package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"lost/internal/db"
	"lost/internal/tagger"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <command>",
	Short: "Add a command to lost",
	Long:  `Add a command manually with optional tags.`,
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := strings.Join(args, " ")
		tags, _ := cmd.Flags().GetStringArray("tag")
		directory, _ := cmd.Flags().GetString("dir")
		shell, _ := cmd.Flags().GetString("shell")

		// Auto-detect tags
		autoTags := tagger.GetAutoTags(command)
		allTags := mergeTags(autoTags, tags)

		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		entry, err := database.AddCommand(command, allTags, directory, shell)
		if err != nil {
			return err
		}

		green := color.GreenString
		cyan := color.CyanString
		fmt.Printf(green("✓")+" Added command %s: %s\n", cyan(fmt.Sprintf("[%d]", entry.ID)), command)
		if len(allTags) > 0 {
			fmt.Printf("  Tags: %s\n", cyan(strings.Join(allTags, ", ")))
		}
		return nil
	},
}

var captureCmd = &cobra.Command{
	Use:    "capture <command> [directory]",
	Short:  "Internal: capture a command from shell hook",
	Hidden: true,
	Args:   cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		command := args[0]
		directory := ""
		if len(args) > 1 {
			directory = args[1]
		}
		shell := detectShell()

		// Skip empty commands and lost commands
		if command == "" || strings.HasPrefix(command, "lost ") {
			return nil
		}

		autoTags := tagger.GetAutoTags(command)

		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		_, err = database.AddCommand(command, autoTags, directory, shell)
		return err
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for commands",
	Long:  `Search through your command history by query string or tags.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		query := strings.Join(args, " ")
		tags, _ := cmd.Flags().GetStringArray("tag")
		runMode, _ := cmd.Flags().GetBool("run")

		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		results, err := database.Search(query, tags)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No commands found.")
			return nil
		}

		if runMode {
			// Execute the most recent matching command
			mostRecent := results[0]
			fmt.Printf("Running: %s\n\n", mostRecent.Command)
			return executeCommand(mostRecent.Command, mostRecent.Directory)
		}

		cyan := color.CyanString
		fmt.Printf(cyan("Found %d command(s):\n\n"), len(results))
		for _, r := range results {
			printCommand(r)
		}
		return nil
	},
}

var recentCmd = &cobra.Command{
	Use:   "recent [limit]",
	Short: "Show recent commands",
	Long:  `Display the most recently captured commands.`,
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		limit := 20
		if len(args) > 0 {
			fmt.Sscanf(args[0], "%d", &limit)
		}

		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		results, err := database.GetRecent(limit)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("No commands found. Start using your terminal or run 'lost install'!")
			return nil
		}

		cyan := color.CyanString
		fmt.Printf(cyan("Recent %d command(s):\n\n"), len(results))
		for _, r := range results {
			printCommand(r)
		}
		return nil
	},
}

var tagCmd = &cobra.Command{
	Use:   "tag <id> <tag>",
	Short: "Add a tag to an existing command",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		var id int64
		fmt.Sscanf(args[0], "%d", &id)
		tag := args[1]

		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		if err := database.TagCommand(id, tag); err != nil {
			return err
		}

		green := color.GreenString
		cyan := color.CyanString
		fmt.Printf(green("✓")+" Tagged command %s with '%s'\n", cyan(fmt.Sprintf("%d", id)), cyan(tag))
		return nil
	},
}

var tagsCmd = &cobra.Command{
	Use:   "tags",
	Short: "List all available tags",
	RunE: func(cmd *cobra.Command, args []string) error {
		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		tags, err := database.GetAllTags()
		if err != nil {
			return err
		}

		if len(tags) == 0 {
			fmt.Println("No tags found yet. Commands will be auto-tagged as you use them!")
			return nil
		}

		cyan := color.CyanString
		yellow := color.YellowString
		fmt.Println(yellow("Available tags:"))
		for _, t := range tags {
			fmt.Printf("  %s %s\n", cyan("•"), t)
		}
		return nil
	},
}

var runCmd = &cobra.Command{
	Use:   "run <tag>",
	Short: "Run the most recent command with a given tag",
	Long:  `Finds and executes the most recently used command that has the specified tag.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		tag := args[0]

		database, err := getDB()
		if err != nil {
			return err
		}
		defer database.Close()

		command, err := database.GetByTag(tag)
		if err != nil {
			return err
		}
		if command == nil {
			return fmt.Errorf("no command found with tag '%s'", tag)
		}

		magenta := color.MagentaString
		cyan := color.CyanString
		fmt.Printf(magenta("Running:")+" %s\n\n", cyan(command.Command))

		// Execute the command in the stored directory
		return executeCommand(command.Command, command.Directory)
	},
}

func init() {
	addCmd.Flags().StringArrayP("tag", "t", []string{}, "Add custom tags (can be used multiple times)")
	addCmd.Flags().StringP("dir", "d", "", "Directory where command was run")
	addCmd.Flags().StringP("shell", "s", "", "Shell used")

	searchCmd.Flags().StringArrayP("tag", "t", []string{}, "Filter by tag (can be used multiple times)")
	searchCmd.Flags().BoolP("run", "r", false, "Execute the most recent matching command")

	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(captureCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(recentCmd)
	rootCmd.AddCommand(tagCmd)
	rootCmd.AddCommand(tagsCmd)
	rootCmd.AddCommand(runCmd)
}

func mergeTags(auto, manual []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, t := range auto {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}

	for _, t := range manual {
		if !seen[t] {
			seen[t] = true
			result = append(result, t)
		}
	}

	return result
}

func printCommand(cmd db.Command) {
	fmt.Printf("[%d] %s\n", cmd.ID, cmd.Command)
	if len(cmd.Tags) > 0 {
		fmt.Printf("    Tags: %s\n", strings.Join(cmd.Tags, ", "))
	}
	fmt.Printf("    %s @ %s\n", cmd.Timestamp.Format("2006-01-02 15:04"), cmd.Directory)
	fmt.Println()
}

func detectShell() string {
	// Simple detection based on environment
	if strings.Contains(getenv("PSMODULEPATH"), "PowerShell") {
		return "powershell"
	}
	if getenv("ZSH_VERSION") != "" {
		return "zsh"
	}
	if getenv("BASH_VERSION") != "" {
		return "bash"
	}
	return "unknown"
}

func executeCommand(command string, directory string) error {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		if _, err := exec.LookPath("powershell"); err == nil {
			cmd = exec.Command("powershell", "-NoProfile", "-Command", command)
		} else {
			cmd = exec.Command("cmd", "/C", command)
		}
	default:
		shell := getenv("SHELL")
		if shell != "" {
			cmd = exec.Command(shell, "-lc", command)
		} else if _, err := exec.LookPath("bash"); err == nil {
			cmd = exec.Command("bash", "-lc", command)
		} else {
			cmd = exec.Command("sh", "-c", command)
		}
	}

	if directory != "" {
		cmd.Dir = directory
	}

	// Connect stdin/stdout/stderr so user can interact
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

