package clitest

import (
	"context"
	"errors"
	"io"
	"os"

	"github.com/avamsi/climate"
	"github.com/avamsi/climate/internal"
	"github.com/avamsi/ergo/assert"
)

type Got struct {
	Stdout, Stderr string
	Code           int
}

type TestCLI func(ctx context.Context, args []string) Got

func New(p internal.Plan, mods ...func(*internal.RunOptions)) TestCLI {
	return func(ctx context.Context, args []string) Got {
		var (
			// TODO: do these pipes have enough capacity?
			stdoutR, stdoutW, err1 = os.Pipe()
			stderrR, stderrW, err2 = os.Pipe()
			osArgs                 = os.Args
			osStdout, osStderr     = os.Stdout, os.Stderr
		)
		assert.Nil(errors.Join(err1, err2))
		os.Args = append([]string{""}, args...)
		os.Stdout, os.Stderr = stdoutW, stderrW
		defer func() {
			os.Args = osArgs
			os.Stdout, os.Stderr = osStdout, osStderr
		}()
		code := climate.Run(ctx, p, mods...)
		assert.Nil(errors.Join(stdoutW.Close(), stderrW.Close()))
		return Got{
			Stdout: string(assert.Ok(io.ReadAll(stdoutR))),
			Stderr: string(assert.Ok(io.ReadAll(stderrR))),
			Code:   code,
		}
	}
}
