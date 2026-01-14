package stream

import (
	"context"
)

//go:generate ../../utils/mockgen.sh
type Conn interface {
	Send(ctx context.Context, data []byte) error
	Recv(ctx context.Context) ([]byte, error)
	Close(ctx context.Context) error

	LocalAddr() string
	RemoteAddr() string
}

//go:generate ../../utils/mockgen.sh
type Listener interface {
	Accept(ctx context.Context) (Conn, error)
	Close(ctx context.Context) error
	LocalAddr() string
}
