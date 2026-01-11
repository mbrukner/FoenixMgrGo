package util

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// Confirm prompts the user for confirmation (y/n) and returns true if confirmed
// This is used for destructive operations like flash erase
func Confirm(prompt string) bool {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print(prompt)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	// Trim whitespace and convert to lowercase
	response = strings.TrimSpace(strings.ToLower(response))

	// Accept 'y' or 'yes'
	return response == "y" || response == "yes"
}

// ConfirmDanger prompts for a more serious confirmation with a warning message
// Returns true only if the user explicitly types "yes"
func ConfirmDanger(operation string) bool {
	fmt.Printf("\n⚠️  WARNING: %s\n", operation)
	fmt.Println("This operation cannot be undone.")
	fmt.Print("\nType 'yes' to confirm: ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	// Trim whitespace and convert to lowercase
	response = strings.TrimSpace(strings.ToLower(response))

	return response == "yes"
}
