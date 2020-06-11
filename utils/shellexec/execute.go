// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package shellexec

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"path"

	"github.com/condensat/bank-core/logger"
)

var (
	// Execute Errors

	ErrInvalidProgram = errors.New("Invalid program")
	ErrTimeout        = errors.New("Timeout reached")
	ErrTimeoutKill    = errors.New("Timeout kill failed")
	ErrUnknown        = errors.New("Unknown error")
)

// Output contains program outputs and ErrorCode
type Output struct {
	Stdout string
	Stderr string
	Code   ErrorCode
}

// Execute execute program in options.
// Returns standard outputs or error in Result
// Kill program execution if last more than timeout option
func Execute(ctx context.Context, options Options) (Output, error) {
	log := logger.Logger(ctx).WithField("Method", "shellexec.Execute")

	log.WithField("Args", fmt.Sprintf("%+v", options.Args)).
		Debug("Execute options")

	if len(options.Program) == 0 {
		return Output{}, ErrInvalidProgram
	}
	// create timeout if required
	var cancel context.CancelFunc = func() {} // default empty cancel func
	if options.Timeout >= MinimumTimeout {
		// override provided context
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
	}
	defer cancel()

	// create Cmd
	commandName := path.Join(options.Path, options.Program)
	cmd := exec.Command(commandName, options.Args...)

	cmd.Env = append(cmd.Env, options.Env...)

	if options.Stdin != nil {
		cmd.Stdin = options.Stdin
	}

	// standard outputs
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return Output{}, err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return Output{}, err
	}

	// start program
	if err := cmd.Start(); err != nil {
		return Output{}, err
	}

	// async chan
	type ChanResult struct {
		Output Output
		Error  error
	}
	resultChan := make(chan ChanResult, 1)
	defer func() {
		close(resultChan)
	}()

	// run async for timeout
	go func() {
		var err error
		stdErr, err := ioutil.ReadAll(stderr)
		if err != nil {
			resultChan <- ChanResult{
				Output: Output{
					Code: ErrorCodeReadStderr,
				},
				Error: err,
			}
			return
		}

		stdOut, err := ioutil.ReadAll(stdout)
		if err != nil {
			resultChan <- ChanResult{
				Output: Output{
					Code: ErrorCodeReadStdout,
				},
				Error: err,
			}
			return
		}

		// wait for exec Command
		err = cmd.Wait()

		// grab outputs & return
		resultChan <- ChanResult{
			Output: Output{
				Stdout: string(stdOut),
				Stderr: string(stdErr),
				Code:   ExitCodeFromError(err),
			},
			Error: err,
		}
	}()

	// sync result or timeout
	select {
	// exec.Command returned
	case result := <-resultChan:
		return result.Output, result.Error

	// Timeout reached
	case <-ctx.Done():
		// Try to kill process
		if err := cmd.Process.Kill(); err != nil {
			// failed to kill, returns timeout kill error
			return Output{
				Code: ErrorCodeKillFailed,
			}, ErrTimeoutKill
		}

		// kill ok, returns timeout error
		return Output{
			Code: ErrorCodeTimeout,
		}, ErrTimeout
	}
}
