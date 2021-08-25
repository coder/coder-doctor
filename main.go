package main

import (
	"context"
	"fmt"
	"os"

	"github.com/cdr/coder-doctor/internal/cmd"
)

func main() {
	defer func() {
		if r := recover(); r != nil {
			_, _ = fmt.Fprintln(os.Stderr, "fatal:", r.(error))
		}
		os.Exit(1)
	}()
	command := cmd.NewDefaultDoctorCommand()
	err := command.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
