package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

// validateCmd represents the validate command
var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "A validation command that uses gator",
	Run: func(cmd *cobra.Command, args []string) {
		useGator, _ := cmd.Flags().GetBool("gator")

		if useGator {
			checkAndInstallGator()
			fmt.Println("Running validation with gator...")
			// Add the logic to call and use gator here
		} else {
			fmt.Println("Running validation without gator...")
			// Add the logic for validation without gator here
		}
	},
}

func init() {
	rootCmd.AddCommand(validateCmd)
	validateCmd.Flags().BoolP("gator", "g", false, "Use gator for validation")
}

func checkAndInstallGator() {
	if !isCommandAvailable("gator") {
		fmt.Println("Gator is not installed.")
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Do you want to install gator? (Y/n): ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" || strings.ToLower(input) == "y" {
			installGator()
		} else if strings.ToLower(input) == "n" {
			fmt.Println("Installation aborted.")
			os.Exit(0)
		} else {
			fmt.Println("Invalid input. Installation aborted.")
			os.Exit(1)
		}
	} else {
		fmt.Println("Gator is already installed.")
	}
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func installGator() {
	fmt.Println("Installing gator...")
	if runtime.GOOS == "darwin" || runtime.GOOS == "linux" {
		cmd := exec.Command("go", "install", "github.com/open-policy-agent/gatekeeper/v3/cmd/gator@master")

		// Set the GOMODPROXY environment variable
		cmd.Env = append(os.Environ(), "GOMODPROXY=http://forwardproxy:3128")

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Printf("Failed to install gator: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Gator installed successfully.")
	} else {
		fmt.Println("Installation is only supported on MacOS or Linux.")
		os.Exit(1)
	}
}
