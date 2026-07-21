// Copyright 2026 Northern.tech AS
//
//	Licensed under the Apache License, Version 2.0 (the "License");
//	you may not use this file except in compliance with the License.
//	You may obtain a copy of the License at
//
//	    http://www.apache.org/licenses/LICENSE-2.0
//
//	Unless required by applicable law or agreed to in writing, software
//	distributed under the License is distributed on an "AS IS" BASIS,
//	WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//	See the License for the specific language governing permissions and
//	limitations under the License.

package stream

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type recvResult struct {
	data []byte
	err  error
}

// addrSeq backs uniqueAddr so every stream pair uses distinct NATS subjects,
// letting tests run in parallel without cross-talk.
var addrSeq atomic.Uint64

func uniqueAddr() string {
	return fmt.Sprintf("peer-%d", addrSeq.Add(1))
}

// startTestNATS runs an in-process NATS server on a random port and returns a
// connected client. Everything is torn down via t.Cleanup.
func startTestNATS(t *testing.T) *nats.Conn {
	t.Helper()
	srv, err := natsserver.NewServer(&natsserver.Options{
		Host: "127.0.0.1",
		Port: -1, // pick a random free port
	})
	require.NoError(t, err)
	go srv.Start()
	if !srv.ReadyForConnections(5 * time.Second) {
		t.Fatal("in-process NATS server did not become ready")
	}
	t.Cleanup(srv.Shutdown)

	nc, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	return nc
}

// connectPair establishes a sender/receiver stream pair over nc: the sender is
// the ConnectNATS side, the receiver is the stream Accept'ed by a listener.
func connectPair(t *testing.T, ctx context.Context, nc *nats.Conn) (sender, receiver Conn) {
	t.Helper()
	addrA := uniqueAddr()
	addrB := uniqueAddr()
	ln, err := ListenNATS(nc, addrA)
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close(context.Background()) })

	type acceptResult struct {
		conn Conn
		err  error
	}
	accepted := make(chan acceptResult, 1)
	go func() {
		c, err := ln.Accept(ctx)
		accepted <- acceptResult{conn: c, err: err}
	}()

	sender, err = ConnectNATS(ctx, nc, addrB, addrA)
	require.NoError(t, err)
	t.Cleanup(func() { _ = sender.Close(context.Background()) })

	select {
	case res := <-accepted:
		require.NoError(t, res.err, "listener failed to accept the stream")
		receiver = res.conn
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for the listener to accept the stream")
	}
	t.Cleanup(func() { _ = receiver.Close(context.Background()) })
	return sender, receiver
}

// TestNATSStreamSlowConsumerNoDrop is the regression guard for the file-transfer
// data loss under an enabled bandwidth limit: a consumer that reads slower than
// the sender produces must never cause the stream to drop or reorder data.
//
// The bug only manifests when a SECOND goroutine drives the receiver stream's
// Send path concurrently with the slow Recv consumer. That Send loop
// independently drains the stream's msgChan and, with the pre-fix code, acks
// the forward chunks BEFORE they are buffered — decoupling the sender's pacing
// from the slow consumer, so recvChan overflows and already-acked chunks are
// silently dropped. This mirrors deviceconnect, where handleDeviceMessages
// calls Send on the same session stream that handleManagementMessages reads
// with Recv. A single-consumer test does NOT reproduce this, because there the
// ack (inside handleMessage) is coupled to the same Recv loop that drains
// recvChan, keeping the sender paced.
func TestNATSStreamSlowConsumerNoDrop(t *testing.T) {
	t.Parallel()
	nc := startTestNATS(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	sender, receiver := connectPair(t, ctx, nc)

	// Background reverse traffic to force the concurrent Send path on both
	// streams. Stopped once the forward transfer is verified.
	noiseCtx, stopNoise := context.WithCancel(ctx)
	defer stopNoise()
	// Sender side keeps draining so the receiver's reverse Sends get acked.
	go func() {
		for {
			if _, err := sender.Recv(noiseCtx); err != nil {
				return
			}
		}
	}()
	// Receiver side keeps sending back, which drives its Send loop to drain
	// msgChan concurrently with the slow forward consumer below.
	go func() {
		for i := 0; ; i++ {
			if err := receiver.Send(noiseCtx, []byte(fmt.Sprintf("noise-%d", i))); err != nil {
				return
			}
		}
	}()

	const n = 500

	// Slow forward consumer: sleep before every Recv so recvChan fills while
	// the reverse-traffic Send loop keeps prematurely acking forward chunks.
	received := make(chan recvResult, n)
	go func() {
		for i := 0; i < n; i++ {
			time.Sleep(time.Millisecond)
			data, err := receiver.Recv(ctx)
			received <- recvResult{data: data, err: err}
			if err != nil {
				return
			}
		}
	}()

	// Fast forward producer.
	sendErr := make(chan error, 1)
	go func() {
		for i := 0; i < n; i++ {
			if err := sender.Send(ctx, []byte(fmt.Sprintf("msg-%d", i))); err != nil {
				sendErr <- err
				return
			}
		}
		sendErr <- nil
	}()

	got := make([]string, 0, n)
	for i := 0; i < n; i++ {
		select {
		case res := <-received:
			require.NoError(t, res.err, "Recv failed after %d messages", i)
			got = append(got, string(res.data))
		case <-ctx.Done():
			t.Fatalf("timed out after receiving %d/%d messages "+
				"(dropped chunks are the pre-fix bug)", i, n)
		}
	}
	require.NoError(t, <-sendErr)
	stopNoise()

	// Every forward message arrives exactly once and in order — nothing dropped.
	require.Len(t, got, n)
	for i := 0; i < n; i++ {
		assert.Equalf(t, fmt.Sprintf("msg-%d", i), got[i],
			"message %d dropped, reordered or corrupted", i)
	}
}

// TestNATSStreamRoundTrip covers the basic bidirectional happy path.
func TestNATSStreamRoundTrip(t *testing.T) {
	t.Parallel()
	nc := startTestNATS(t)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	sender, receiver := connectPair(t, ctx, nc)

	payloads := [][]byte{
		[]byte("hello"),
		[]byte(""),
		make([]byte, 4096), // a full-size chunk
	}
	for i := range payloads[2] {
		payloads[2][i] = byte(i % 251)
	}

	received := make(chan recvResult, len(payloads))
	go func() {
		for range payloads {
			data, err := receiver.Recv(ctx)
			received <- recvResult{data: data, err: err}
			if err != nil {
				return
			}
		}
	}()

	for _, p := range payloads {
		require.NoError(t, sender.Send(ctx, p))
	}
	for i, want := range payloads {
		select {
		case res := <-received:
			require.NoError(t, res.err)
			assert.Equalf(t, want, res.data, "payload %d mismatch", i)
		case <-ctx.Done():
			t.Fatalf("timed out waiting for payload %d", i)
		}
	}
}
