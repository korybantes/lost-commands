package tagger

import (
	"path/filepath"
	"strings"
)

// AutoTagRules define pattern → tags mapping
var AutoTagRules = []struct {
	Pattern string
	Tags    []string
	Match   func(cmd string) bool
}{
	{
		Tags: []string{"git", "vcs"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "git ")
		},
	},
	{
		Tags: []string{"docker", "containers", "devops"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "docker ") || strings.HasPrefix(cmd, "docker-compose ")
		},
	},
	{
		Tags: []string{"kubernetes", "k8s", "devops"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "kubectl ") || strings.HasPrefix(cmd, "helm ")
		},
	},
	{
		Tags: []string{"nodejs", "package-manager"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "npm ") || strings.HasPrefix(cmd, "yarn ") || strings.HasPrefix(cmd, "pnpm ")
		},
	},
	{
		Tags: []string{"python", "package-manager"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "python ") || strings.HasPrefix(cmd, "python3 ") ||
				strings.HasPrefix(cmd, "pip ") || strings.HasPrefix(cmd, "pip3 ") ||
				strings.HasPrefix(cmd, "pytest ") || strings.HasPrefix(cmd, "flask ") ||
				strings.HasPrefix(cmd, "django-admin ")
		},
	},
	{
		Tags: []string{"rust", "cargo"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "cargo ") || strings.HasPrefix(cmd, "rustc ")
		},
	},
	{
		Tags: []string{"golang", "go"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "go ")
		},
	},
	{
		Tags: []string{"build"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "make ") || strings.HasPrefix(cmd, "cmake ") ||
				strings.HasPrefix(cmd, "gcc ") || strings.HasPrefix(cmd, "g++ ") ||
				cmd == "make" || cmd == "gcc" || cmd == "g++"
		},
	},
	{
		Tags: []string{"ssh", "remote"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "ssh ") || strings.HasPrefix(cmd, "scp ") || strings.HasPrefix(cmd, "sftp ")
		},
	},
	{
		Tags: []string{"file", "filesystem"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "ls ") || strings.HasPrefix(cmd, "dir ") ||
				strings.HasPrefix(cmd, "cp ") || strings.HasPrefix(cmd, "copy ") ||
				strings.HasPrefix(cmd, "mv ") || strings.HasPrefix(cmd, "move ") ||
				strings.HasPrefix(cmd, "rm ") || strings.HasPrefix(cmd, "del ") ||
				strings.HasPrefix(cmd, "cat ") || strings.HasPrefix(cmd, "type ") ||
				strings.HasPrefix(cmd, "find ") || strings.HasPrefix(cmd, "grep ") ||
				strings.HasPrefix(cmd, "mkdir ") || strings.HasPrefix(cmd, "rmdir ") ||
				strings.HasPrefix(cmd, "touch ") || strings.HasPrefix(cmd, "chmod ")
		},
	},
	{
		Tags: []string{"archive", "compression"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "tar ") || strings.HasPrefix(cmd, "zip ") ||
				strings.HasPrefix(cmd, "unzip ") || strings.HasPrefix(cmd, "gzip ") ||
				strings.HasPrefix(cmd, "gunzip ") || strings.HasPrefix(cmd, "7z ")
		},
	},
	{
		Tags: []string{"database", "sql"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "mysql ") || strings.HasPrefix(cmd, "psql ") ||
				strings.HasPrefix(cmd, "mongo ") || strings.HasPrefix(cmd, "redis-cli ") ||
				strings.HasPrefix(cmd, "sqlite3 ")
		},
	},
	{
		Tags: []string{"text-processing"},
		Match: func(cmd string) bool {
			return strings.HasPrefix(cmd, "sed ") || strings.HasPrefix(cmd, "awk ") ||
				strings.HasPrefix(cmd, "cut ") || strings.HasPrefix(cmd, "sort ") ||
				strings.HasPrefix(cmd, "uniq ") || strings.HasPrefix(cmd, "wc ")
		},
	},
}

// GetAutoTags returns automatically assigned tags for a command
func GetAutoTags(cmd string) []string {
	tagSet := make(map[string]bool)
	
	for _, rule := range AutoTagRules {
		if rule.Match(cmd) {
			for _, tag := range rule.Tags {
				tagSet[tag] = true
			}
		}
	}

	// Detect file types in the command
	tagsFromFiles := detectFileTypes(cmd)
	for _, tag := range tagsFromFiles {
		tagSet[tag] = true
	}

	var tags []string
	for tag := range tagSet {
		tags = append(tags, tag)
	}
	return tags
}

func detectFileTypes(cmd string) []string {
	var tags []string
	
	// Extract file extensions from the command
	words := strings.Fields(cmd)
	for _, word := range words {
		ext := strings.ToLower(filepath.Ext(word))
		switch ext {
		case ".js", ".ts", ".jsx", ".tsx":
			tags = append(tags, "javascript")
		case ".py":
			tags = append(tags, "python")
		case ".go":
			tags = append(tags, "golang")
		case ".rs":
			tags = append(tags, "rust")
		case ".java":
			tags = append(tags, "java")
		case ".c", ".cpp", ".h", ".hpp":
			tags = append(tags, "c-cpp")
		case ".sh", ".bash", ".zsh":
			tags = append(tags, "shell")
		case ".ps1":
			tags = append(tags, "powershell")
		case ".json":
			tags = append(tags, "json")
		case ".yaml", ".yml":
			tags = append(tags, "yaml")
		case ".md":
			tags = append(tags, "markdown")
		case ".sql":
			tags = append(tags, "sql")
		case ".html", ".htm":
			tags = append(tags, "html")
		case ".css":
			tags = append(tags, "css")
		}
	}
	
	return tags
}
