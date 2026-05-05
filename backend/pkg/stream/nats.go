package stream

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
)

// [prefix.]streamv1.<localAddr>.<remoteAddr>.<messageType>[:sequenceNum]
//
// messageType = enum{
//   data
//   ack
//   hello
//   bye
// }
//
// Subscriptions:
//   listener: [prefix.]streamv1.l.<localAddr>.*.hello
//   session:  [prefix.]streamv1.s.<localAddr>.<remoteAddr>.*

type natsStream struct {
	nc         *nats.Conn
	subSession *nats.Subscription
	msgChan    chan *nats.Msg

	retryDelay time.Duration
	retryTimer *time.Timer

	closed   chan struct{}
	recvChan chan []byte
	ackChan  chan uint32
	sendOnce chan struct{}

	localAddr  string
	remoteAddr string

	remoteSeq uint32
	localSeq  uint32

	closeFunc func()
	termOnce  sync.Once
}

func newStream(nc *nats.Conn, localAddr, remoteAddr string) (*natsStream, error) {
	msgChan := make(chan *nats.Msg, 8)
	sub, err := nc.ChanSubscribe(fmtSessionSubject(localAddr, remoteAddr, "*"), msgChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to session: %w", err)
	}
	ret := &natsStream{
		nc:         nc,
		subSession: sub,
		msgChan:    msgChan,

		retryTimer: time.NewTimer(0),

		localAddr:  localAddr,
		remoteAddr: remoteAddr,
		recvChan:   make(chan []byte, 4),
		ackChan:    make(chan uint32, 4),
		closed:     make(chan struct{}),
		sendOnce:   make(chan struct{}, 1),

		remoteSeq: 1,
		localSeq:  0,
	}
	return ret, nil
}

func fmtHelloSubject(localAddr, remoteAddr string) string {
	return fmt.Sprintf("streamv1.l.%s.%s.hello", localAddr, remoteAddr)
}

func fmtSessionSubject(localAddr, remoteAddr, msgType string) string {
	return fmt.Sprintf("streamv1.s.%s.%s.%s", localAddr, remoteAddr, msgType)
}

func cutLast(s, substr string) (before, after string, found bool) {
	if i := strings.LastIndex(s, substr); i >= 0 {
		return s[:i], s[i+len(substr):], true
	}
	return s, "", false
}

var (
	ErrProtocol          = errors.New("protocol error")
	ErrClosed            = errors.New("stream closed")
	ErrConnectionRefused = errors.New("connection refused")
)

func ConnectNATS(ctx context.Context, nc *nats.Conn, srcAddr, dstAddr string) (Conn, error) {
	stream, err := newStream(nc, srcAddr, dstAddr)
	defer func() {
		if err != nil {
			_ = stream.term()
		}
	}()

	err = nc.PublishRequest(
		fmtHelloSubject(dstAddr, srcAddr),
		fmtSessionSubject(srcAddr, dstAddr, "ack:0"),
		nil)
	if err != nil {
		err = fmt.Errorf("failed to submit handshake to peer: %w", err)
		return nil, err
	}
	if !stream.retryTimer.Stop() {
		<-stream.retryTimer.C
	}
	stream.retryDelay = 100 * time.Millisecond
	stream.retryTimer.Reset(stream.retryDelay)
	for range 3 {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return nil, err
		case msg := <-stream.msgChan:
			err := stream.handleMessage(ctx, msg)
			if err != nil {
				return nil, err
			}
		case ack := <-stream.ackChan:
			if ack == 0 {
				return stream, nil
			}
		case <-stream.retryTimer.C:
			err = nc.PublishRequest(
				fmtHelloSubject(dstAddr, srcAddr),
				fmtSessionSubject(srcAddr, dstAddr, "ack:0"),
				nil)
			if err != nil {
				err = fmt.Errorf("failed to submit handshake to peer: %w", err)
				return nil, err
			}
			stream.retryDelay *= 2
			stream.retryTimer.Reset(stream.retryDelay)
		}
	}
	return nil, fmt.Errorf("failed to establish connection with peer")
}

func (stream *natsStream) handleMessage(ctx context.Context, msg *nats.Msg) error {
	if msg.Header.Get("Status") == "503" {
		return ErrConnectionRefused
	}
	_, msgType, _ := cutLast(msg.Subject, ".")
	switch {
	case msgType == "bye":
		if msg.Reply != "" {
			_ = msg.Respond(nil)
		}
		return stream.term()
	case msgType == "byeack":
		return stream.term()

	case strings.HasPrefix(msgType, "data"):
		err := msg.Respond(nil)
		if err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
		_, seqNumStr, _ := cutLast(msg.Subject, ":")
		seqNum, _ := strconv.ParseUint(seqNumStr, 10, 32)
		diff := stream.remoteSeq - uint32(seqNum)
		if diff == 0 {
			stream.remoteSeq++
			select {
			case stream.recvChan <- msg.Data:
			default:
			}
			return nil
		} else if diff == 1 {
			// ignore ack for last package
			return nil
		}
		errClose := stream.Close(ctx)
		if errClose != nil {
			return fmt.Errorf("error closing stream on protocol violation: %w", errClose)
		}
		return ErrProtocol

	case strings.HasPrefix(msgType, "ack"):
		_, seqNumStr, _ := cutLast(msg.Subject, ":")
		seqNum, _ := strconv.ParseUint(seqNumStr, 10, 32)
		// NOTE: "modulo arithmetic"
		diff := stream.localSeq - uint32(seqNum)
		if diff == 0 {
			select {
			case stream.ackChan <- uint32(seqNum):
			default:
			}
			return nil
		} else if diff == 1 {
			// ignore ack for last package
			return nil
		}
		errClose := stream.Close(ctx)
		if errClose != nil {
			return fmt.Errorf("error closing stream on protocol violation: %w", errClose)
		}
		return ErrProtocol
	}
	return nil
}

func (stream *natsStream) term() error {
	err := ErrClosed
	stream.termOnce.Do(func() {
		close(stream.closed)
		if stream.closeFunc != nil {
			stream.closeFunc()
		}
		err = stream.subSession.Unsubscribe()
	})
	return err
}

func (stream *natsStream) sendData(data []byte) error {
	msgTypeData := fmt.Sprintf("data:%d", stream.localSeq)
	msgTypeAck := fmt.Sprintf("ack:%d", stream.localSeq)
	err := stream.nc.PublishRequest(
		fmtSessionSubject(stream.remoteAddr, stream.localAddr, msgTypeData),
		fmtSessionSubject(stream.localAddr, stream.remoteAddr, msgTypeAck),
		data,
	)
	if err != nil {
		return err
	}
	stream.retryDelay *= 2
	stream.retryTimer.Reset(stream.retryDelay)
	return nil
}

func (stream *natsStream) Send(ctx context.Context, data []byte) error {
	select {
	case stream.sendOnce <- struct{}{}:
		defer func() { <-stream.sendOnce }()
	case <-ctx.Done():
		return ctx.Err()
	}
	stream.localSeq++
	stream.retryDelay = 100 * time.Millisecond
	if !stream.retryTimer.Stop() {
		<-stream.retryTimer.C
	}
	err := stream.sendData(data)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}
	for {
		select {
		case <-stream.closed:
			return ErrClosed

		case <-stream.retryTimer.C:
			err = stream.sendData(data)
			if err != nil {
				return fmt.Errorf("failed to publish message: %w", err)
			}

		case msg := <-stream.msgChan:
			err := stream.handleMessage(ctx, msg)
			if err != nil {
				return err
			}

		case <-stream.ackChan:
			return nil

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (stream *natsStream) Recv(ctx context.Context) ([]byte, error) {
	for {
		select {
		case data := <-stream.recvChan:
			return data, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		case msg := <-stream.msgChan:
			err := stream.handleMessage(ctx, msg)
			if err != nil {
				return nil, err
			}
		}
	}
}

func (stream *natsStream) Close(ctx context.Context) error {
	err := stream.nc.PublishRequest(
		fmtSessionSubject(stream.remoteAddr, stream.localAddr, "bye"),
		fmtSessionSubject(stream.localAddr, stream.remoteAddr, "byeack"),
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to send close to peer: %w", err)
	}
	if !stream.retryTimer.Stop() {
		<-stream.retryTimer.C
	}
	stream.retryDelay = 100 * time.Millisecond
	stream.retryTimer.Reset(stream.retryDelay)
	for range 3 {
		select {
		case msg := <-stream.msgChan:
			err = stream.handleMessage(ctx, msg)
			if err != nil {
				return err
			}
		case <-stream.closed:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case <-stream.retryTimer.C:
			err := stream.nc.PublishRequest(
				fmtSessionSubject(stream.remoteAddr, stream.localAddr, "bye"),
				fmtSessionSubject(stream.localAddr, stream.remoteAddr, "byeack"),
				nil,
			)
			if err != nil {
				return fmt.Errorf("failed to send close to peer: %w", err)
			}
			stream.retryDelay *= 2
			stream.retryTimer.Reset(stream.retryDelay)
		}
	}
	_ = stream.term()
	return fmt.Errorf("failed to notify peer to close connection")
}

func (stream *natsStream) RemoteAddr() string {
	return stream.remoteAddr
}

func (stream *natsStream) LocalAddr() string {
	return stream.localAddr
}

type natsListener struct {
	nc        *nats.Conn
	subListen *nats.Subscription

	msgChan chan *nats.Msg
	addr    string

	closed    chan struct{}
	closeOnce sync.Once

	// openConns track open connections
	openConns map[string]Conn
	mu        sync.Mutex
}

func ListenNATS(nc *nats.Conn, addr string) (Listener, error) {
	msgChan := make(chan *nats.Msg, 3)
	sub, err := nc.ChanSubscribe(fmtHelloSubject(addr, "*"), msgChan)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to listener subject: %w", err)
	}

	return &natsListener{
		nc:        nc,
		subListen: sub,
		addr:      addr,
		msgChan:   msgChan,
		openConns: make(map[string]Conn),

		closed: make(chan struct{}),
	}, nil
}

func (l *natsListener) Close(ctx context.Context) error {
	err := l.subListen.Unsubscribe()
	if err != nil {
		return err
	}
	for _, stream := range l.openConns {
		err = stream.Close(ctx)
		if err != nil {
			return err
		}
	}
	l.closeOnce.Do(func() {
		close(l.closed)
	})
	return nil
}

func (l *natsListener) Accept(ctx context.Context) (Conn, error) {
	select {
	case <-l.closed:
		return nil, ErrClosed
	case <-ctx.Done():
		return nil, ctx.Err()
	case msg := <-l.msgChan:
		rest, msgType, _ := cutLast(msg.Subject, ".")
		if msgType != "hello" {
			return nil, ErrProtocol
		}
		_, addr, _ := cutLast(rest, ".")

		_, ok := l.openConns[addr]
		if ok {
			return nil, ErrProtocol
		}
		stream, err := newStream(l.nc, l.addr, addr)
		if err != nil {
			return nil, err
		}
		l.mu.Lock()
		if _, ok := l.openConns[addr]; ok {
			l.mu.Unlock()
			_ = stream.term()
			return nil, ErrProtocol
		} else {
			l.openConns[addr] = stream
		}
		l.mu.Unlock()
		stream.closeFunc = func() {
			l.mu.Lock()
			defer l.mu.Unlock()
			delete(l.openConns, addr)
		}
		err = msg.Respond(nil)
		if err != nil {
			l.openConns[addr].Close(ctx)
			return nil, fmt.Errorf("failed to complete handshake: %w", err)
		}
		return stream, nil
	}
}

func (l *natsListener) LocalAddr() string {
	return l.addr
}
