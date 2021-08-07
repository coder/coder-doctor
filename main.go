package main

import (
	"context"

	"github.com/cdr/doctor/internal/cmd"
)

func main() {
	command := cmd.NewDefaultDoctorCommand()
	command.ExecuteContext(context.Background())
}
