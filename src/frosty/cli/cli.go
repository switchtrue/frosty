package cli

import (
	"os"
)

const (
	CommandBackup = "backup"
)

func Main() {
	runCommand := os.Args[1]

	if runCommand == CommandBackup {
		Read(os.Args[2])
	}
}
