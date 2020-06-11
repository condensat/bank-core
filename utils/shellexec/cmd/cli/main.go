// Copyright 2020 Condensat Tech. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/condensat/bank-core/utils/shellexec"
)

func main() {
	ctx := context.Background()

	var program string
	if len(os.Args) > 1 {
		program = os.Args[1]
	}
	var args []string
	if len(os.Args) > 2 {
		args = os.Args[2:]
	}

	result, _ := shellexec.Execute(ctx,
		shellexec.
			DefaultOptions().
			WithProgram(program).
			WithStdin(os.Stdin).
			WithArgs(args...),
	)

	if len(result.Stderr) != 0 {
		fmt.Fprintf(os.Stderr, result.Stderr)
	}
	if len(result.Stdout) != 0 {
		fmt.Fprintf(os.Stdout, result.Stdout)
	}

	os.Exit(int(result.Code))
}
