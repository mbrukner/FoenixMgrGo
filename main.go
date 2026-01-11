// FoenixMgr - Command-line tool for managing Foenix retro computers
//
// This tool enables uploading binaries, programming flash memory,
// reading/writing memory, and controlling CPU state over serial or TCP.
package main

import (
	"fmt"
	"os"

	"github.com/daschewie/foenixmgr/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
