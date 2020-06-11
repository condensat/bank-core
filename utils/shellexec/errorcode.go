// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package shellexec

import (
	"os/exec"
	"syscall"
)

// ErrorCode type
type ErrorCode int

const (
	// ErrorCode values
	ErrorCodeUnknown    = ErrorCode(100)
	ErrorCodeReadStdout = ErrorCode(101)
	ErrorCodeReadStderr = ErrorCode(102)
	ErrorCodeTimeout    = ErrorCode(103)
	ErrorCodeKillFailed = ErrorCode(104)
)

// ExitCodeFromError return ErroCode from error
func ExitCodeFromError(err error) (code ErrorCode) {
	if exiterr, ok := err.(*exec.ExitError); ok {
		// The program has exited with an exit code != 0

		// This works on both Unix and Windows. Although package
		// syscall is generally platform dependent, WaitStatus is
		// defined for both Unix and Windows and in both cases has
		// an ExitStatus() method with the same signature.
		if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
			code = ErrorCode(status.ExitStatus())
		}
	}
	return
}
