package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"lost/internal/shell"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var installCmd = &cobra.Command{
	Use:   "install [shell]",
	Short: "Install lost to PATH and setup shell integration",
	Long: `Installs lost by:
1. Adding lost to your PATH (if not already there)
2. Setting up shell hooks for automatic command tracking

Supported shells: powershell, bash, zsh
If no shell is specified, it will auto-detect.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		shellName := "auto"
		if len(args) > 0 {
			shellName = args[0]
		}

		// Step 1: Ensure lost is in PATH
		if err := ensureInPath(); err != nil {
			fmt.Printf("Warning: Could not add to PATH: %v\n", err)
			fmt.Println("You may need to manually add lost to your PATH.")
		}

		green := color.GreenString
		cyan := color.CyanString
		yellow := color.YellowString

		// Step 2: Install shell hooks
		if err := shell.Install(shellName); err != nil {
			return fmt.Errorf("shell integration failed: %w", err)
		}

		fmt.Println()
		fmt.Println(green("✓ Shell integration installed successfully!"))
		fmt.Println()
		fmt.Println(yellow("To finish setup:"))
		switch runtime.GOOS {
		case "windows":
			fmt.Println(cyan("  1. Close and reopen PowerShell"))
			fmt.Println(cyan("  2. Run 'lost' to verify it's working"))
			fmt.Println(cyan("  3. Your commands will be automatically tracked!"))
		default:
			fmt.Println(cyan("  1. Run: source ~/.bashrc  (or ~/.zshrc)"))
			fmt.Println(cyan("  2. Run 'lost' to verify it's working"))
			fmt.Println(cyan("  3. Your commands will be automatically tracked!"))
		}

		return nil
	},
}

func ensureInPath() error {
	// Get the directory where lost binary is located
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not find executable path: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// Check if already in PATH
	pathEnv := os.Getenv("PATH")
	pathDirs := strings.Split(pathEnv, string(os.PathListSeparator))
	for _, dir := range pathDirs {
		if strings.EqualFold(filepath.Clean(dir), filepath.Clean(exeDir)) {
			return nil // Already in PATH
		}
	}

	// Not in PATH, need to add it
	fmt.Printf("Adding %s to PATH...\n", exeDir)

	if runtime.GOOS == "windows" {
		return addToWindowsPath(exeDir)
	}

	return addToUnixPath(exeDir)
}

func addToWindowsPath(dir string) error {
	// Use PowerShell to add to user PATH persistently
	cmd := exec.Command("powershell", "-Command",
		fmt.Sprintf(`[Environment]::SetEnvironmentVariable("Path", $env:Path + ";%s", "User")`, dir))
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to update PATH: %w", err)
	}
	green := color.GreenString
	fmt.Println(green("✓ Added to user PATH"))
	return nil
}

func addToUnixPath(dir string) error {
	// For Unix, we'll add to shell rc files
	shell := shell.DetectShell()
	var rcFile string
	switch shell {
	case "zsh":
		rcFile = filepath.Join(os.Getenv("HOME"), ".zshrc")
	case "bash":
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
	default:
		rcFile = filepath.Join(os.Getenv("HOME"), ".bashrc")
	}

	// Check if already added
	content, err := os.ReadFile(rcFile)
	if err == nil && strings.Contains(string(content), dir) {
		return nil // Already in PATH via rc file
	}

	// Add export statement
	exportLine := fmt.Sprintf("\n# Added by lost installer\nexport PATH=\"$PATH:%s\"\n", dir)
	f, err := os.OpenFile(rcFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(exportLine)
	if err == nil {
		green := color.GreenString
		fmt.Printf(green("✓ Added PATH export to %s\n"), rcFile)
	}
	return err
}

func init() {
	rootCmd.AddCommand(installCmd)
}

// Fix for getenv function used in commands.go
func getenv(key string) string {
	return os.Getenv(key)
}
