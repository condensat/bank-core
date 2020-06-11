// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package shellexec

import (
	"io"
	"time"
)

const (
	// DefaultTimeout for exiting
	DefaultTimeout = 5 * time.Second
	MinimumTimeout = 100 * time.Millisecond
)

// Options for Execute
type Options struct {
	Timeout time.Duration // Timeout execution
	Program string        // program name
	Path    string        // program path (optional)
	Env     []string      // Program env (optional)
	Stdin   io.Reader     // Program Stdin (optional)
	Args    []string      // Program arguments (optional)
}

// DefaultOptions return defaults
func DefaultOptions() Options {
	return Options{
		Timeout: DefaultTimeout, // DefaultTimeout
	}
}

// WithTimeout add timeout to options
func (option Options) WithTimeout(timeout time.Duration) Options {
	if timeout < MinimumTimeout {
		timeout = DefaultTimeout
	}
	result := option
	result.Timeout = timeout
	return result
}

// WithProgram add program to options
func (option Options) WithProgram(program string) Options {
	result := option
	result.Program = program
	return result
}

// WithPath add path to options
func (option Options) WithPath(path string) Options {
	result := option
	result.Path = path
	return result
}

// WithEnv add env to options
func (option Options) WithEnv(env ...string) Options {
	result := option
	result.Env = env
	return result
}

// WithStdin add stdin to options
func (option Options) WithStdin(stdin io.Reader) Options {
	result := option
	result.Stdin = stdin
	return result
}

// WithArgs add args to options
func (option Options) WithArgs(args ...string) Options {
	result := option
	result.Args = args
	return result
}
