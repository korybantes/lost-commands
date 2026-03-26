package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"lost/internal/db"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	dbPath string
	rootCmd = &cobra.Command{
		Use:   "lost",
		Short: "Never lose a terminal command again",
		Long:  getBanner(),
	}
)

func init() {
	home, _ := os.UserHomeDir()
	defaultDbPath := filepath.Join(home, ".lost", "commands.db")
	
	rootCmd.PersistentFlags().StringVar(&dbPath, "db", defaultDbPath, "Path to the database file")
	
	rootCmd.Version = "1.0.0\nCreated by @korybantes (Ertac Toptutan)"
}

func getBanner() string {
	var b strings.Builder
	
	cyan := color.CyanString
	magenta := color.MagentaString
	yellow := color.YellowString
	green := color.GreenString
	
	b.WriteString("\n")
	b.WriteString(cyan("╔═══════════════════════════════════════════╗\n"))
	b.WriteString(cyan("║ ") + magenta("██╗      ██████╗ ███████╗████████╗") + cyan("  ║\n"))
	b.WriteString(cyan("║ ") + magenta("██║     ██╔═══██╗██╔════╝╚══██╔══╝") + cyan("  ║\n"))
	b.WriteString(cyan("║ ") + magenta("██║     ██║   ██║███████╗   ██║   ") + cyan("  ║\n"))
	b.WriteString(cyan("║ ") + magenta("██║     ██║   ██║╚════██║   ██║   ") + cyan("  ║\n"))
	b.WriteString(cyan("║ ") + magenta("███████╗╚██████╔╝███████║   ██║   ") + cyan("  ║\n"))
	b.WriteString(cyan("║ ") + magenta("╚══════╝ ╚═════╝ ╚══════╝   ╚═╝   ") + cyan("  ║\n"))
	b.WriteString(cyan("╠═══════════════════════════════════════════╣\n"))
	b.WriteString(cyan("║  ") + yellow("Never lose a terminal command again!") + cyan("   ║\n"))
	b.WriteString(cyan("╚═══════════════════════════════════════════╝\n"))
	b.WriteString("\n")
	b.WriteString(green("✨ ") + "Captures, tags & searches your terminal commands\n")
	b.WriteString(yellow("🔧 ") + "Works with PowerShell, Bash, and Zsh\n")
	b.WriteString(magenta("🏷️  ") + "Intelligent auto-tagging for easy retrieval\n")
	b.WriteString("\n")
	b.WriteString(yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	b.WriteString(yellow("QUICK REFERENCE:\n"))
	b.WriteString(yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	b.WriteString(cyan("  lost add \"<cmd>\" -t <tag>     ") + "Add command with tag\n")
	b.WriteString(cyan("  lost search <query>           ") + "Search commands\n")
	b.WriteString(cyan("  lost search -t <tag> -r       ") + "Search by tag & run\n")
	b.WriteString(cyan("  lost run <tag>                ") + "Run command by tag\n")
	b.WriteString(cyan("  lost recent                   ") + "Show recent commands\n")
	b.WriteString(cyan("  lost tags                     ") + "List all tags\n")
	b.WriteString(yellow("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n"))
	b.WriteString("\n")
	b.WriteString(cyan("Created by @korybantes (Ertac Toptutan)\n"))
	
	return b.String()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func getDB() (*db.Database, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}
	return db.New(dbPath)
}
