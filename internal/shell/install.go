package shell

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// Install sets up shell hooks for automatic command capture
func Install(shell string) error {
	switch shell {
	case "powershell", "pwsh":
		return installPowerShell()
	case "bash":
		return installBash()
	case "zsh":
		return installZsh()
	case "auto":
		return installAuto()
	default:
		return fmt.Errorf("unsupported shell: %s", shell)
	}
}

func installAuto() error {
	switch runtime.GOOS {
	case "windows":
		return installPowerShell()
	default:
		// Check for zsh first, then bash
		if _, err := os.Stat(os.ExpandEnv("$HOME/.zshrc")); err == nil {
			return installZsh()
		}
		return installBash()
	}
}

func installPowerShell() error {
	profilePath := getPowerShellProfile()
	if profilePath == "" {
		return fmt.Errorf("could not find PowerShell profile")
	}

	// Ensure directory exists
	dir := filepath.Dir(profilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create profile directory: %w", err)
	}

	hook := `
# Lost - Command Tracker Hook
$global:LostLastCommand = ""
$global:LostPromptOriginal = $function:prompt

function global:prompt {
    $currentCmd = (Get-History -Count 1).CommandLine
    if ($global:LostLastCommand -and $global:LostLastCommand -ne $currentCmd -and 
        -not $global:LostLastCommand.StartsWith("lost ")) {
        $exe = "lost"
        $location = (Get-Location).Path
        Start-Process -FilePath $exe -ArgumentList "capture", $global:LostLastCommand, $location -WindowStyle Hidden
    }
    $global:LostLastCommand = $currentCmd
    & $global:LostPromptOriginal
}
# End Lost Hook
`

	return appendToFile(profilePath, hook)
}

func installBash() error {
	rcPath := filepath.Join(os.ExpandEnv("$HOME"), ".bashrc")
	
	hook := `
# Lost - Command Tracker Hook
__lost_capture() {
    local last_cmd=$(history 1 | sed 's/^[ ]*[0-9]*[ ]*//')
    if [[ -n "$last_cmd" && ! "$last_cmd" =~ ^lost[[:space:]] ]]; then
        (lost capture "$last_cmd" "$PWD" &)
    fi
}
PROMPT_COMMAND="${PROMPT_COMMAND:+$PROMPT_COMMAND$'\\n'}__lost_capture"
# End Lost Hook
`

	return appendToFile(rcPath, hook)
}

func installZsh() error {
	rcPath := filepath.Join(os.ExpandEnv("$HOME"), ".zshrc")
	
	hook := `
# Lost - Command Tracker Hook
__lost_capture() {
    local last_cmd=$(fc -ln -1)
    if [[ -n "$last_cmd" && ! "$last_cmd" =~ ^lost[[:space:]] ]]; then
        (lost capture "$last_cmd" "$PWD" &)
    fi
}
autoload -U add-zsh-hook
add-zsh-hook precmd __lost_capture
# End Lost Hook
`

	return appendToFile(rcPath, hook)
}

func getPowerShellProfile() string {
	// Try to get the profile path from environment
	if profile := os.Getenv("USERPROFILE"); profile != "" {
		// PowerShell 7+ profile
		ps7Profile := filepath.Join(profile, "Documents", "PowerShell", "Microsoft.PowerShell_profile.ps1")
		if _, err := os.Stat(ps7Profile); err == nil || os.IsNotExist(err) {
			return ps7Profile
		}
		// Windows PowerShell profile
		ps5Profile := filepath.Join(profile, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")
		return ps5Profile
	}
	return ""
}

// DetectShell returns the current shell type
func DetectShell() string {
	// Check for PowerShell first (Windows-specific env vars)
	if os.Getenv("PSMODULEPATH") != "" {
		return "powershell"
	}
	// Check shell via process inspection or environment
	shell := os.Getenv("SHELL")
	if strings.Contains(shell, "zsh") {
		return "zsh"
	}
	if strings.Contains(shell, "bash") {
		return "bash"
	}
	return "bash" // Default
}

func appendToFile(path, content string) error {
	// Check if already installed
	existing, err := os.ReadFile(path)
	if err == nil && strings.Contains(string(existing), "Lost - Command Tracker Hook") {
		fmt.Println("Lost hook already installed in", path)
		return nil
	}

	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	if err == nil {
		fmt.Printf("✓ Installed lost hook to %s\n", path)
		fmt.Println("  Please restart your shell or run: source", path)
	}
	return err
}
