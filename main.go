package main

import (
	"context"
	"os"

	"github.com/cdr/coder-doctor/internal/cmd"
)

func main() {
	command := cmd.NewDefaultDoctorCommand()
	err := command.ExecuteContext(context.Background())
	if err != nil {
		os.Exit(1)
	}

	os.Exit(0)
}
