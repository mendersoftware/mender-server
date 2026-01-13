package executor

import (
	"context"
	"errors"
	"os/exec"
	"slices"
)

var (
	ErrCommandNotAllowed = errors.New("command not allowed")
)

//go:generate ../../utils/mockgen.sh
type BinaryExecutor interface {
	Command(ctx context.Context, binary string, arg ...string) (*exec.Cmd, error)
}

type executor struct {
	allowedBinaries []string
}

// New creates a new executor with a specific set of allowed binaries.
func New(allowed []string) BinaryExecutor {
	return &executor{
		allowedBinaries: allowed,
	}
}

func (e *executor) Command(ctx context.Context, binary string, arg ...string) (*exec.Cmd, error) {
	if ok := slices.Contains(e.allowedBinaries, binary); !ok {
		return nil, ErrCommandNotAllowed
	}
	if ctx == nil {
		ctx = context.Background()
	}

	return exec.CommandContext(ctx, binary, arg...), nil
}
