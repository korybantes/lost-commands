# Lost - Never Lose a Terminal Command Again

Lost is a cross-platform tool that captures, tags, and searches your terminal commands. It works with PowerShell, Bash, and Zsh to automatically track and categorize your command history with intelligent auto-tagging.

## Features

- **Automatic Command Capture**: Works silently in the background via shell hooks
- **Intelligent Auto-Tagging**: Commands are automatically tagged based on patterns (git, docker, npm, python, etc.)
- **Full-Text Search**: Search through your entire command history
- **Tag-Based Filtering**: Find commands by tags
- **Run by Tag**: Execute the most recent command for a tag with `lost run <tag>`
- **Quick Run Shortcut**: Search and execute in one step with `lost search -t <tag> -r`
- **Cross-Platform**: Works on Windows (PowerShell), macOS, and Linux (Bash/Zsh)

## Installation

### Pre-built Binaries (Recommended)

Download the latest binary for your platform from [GitHub Releases](https://github.com/korybantes/lost-commands/releases):

**Windows (PowerShell):**
```powershell
# 1. Download lost-windows-amd64.exe, rename to lost.exe
# 2. Put it anywhere (e.g., Downloads folder, Desktop, etc.)
# 3. Run the installer - it will add to PATH and setup auto-capture:
.\lost.exe install
# 4. Close and reopen PowerShell - done!
```

**macOS/Linux:**
```bash
# 1. Download appropriate binary
# 2. Put it anywhere (e.g., ~/bin, /tmp, etc.)
# 3. Run the installer - it will add to PATH and setup auto-capture:
chmod +x lost-*
./lost-* install
# 4. Restart your shell or run: source ~/.bashrc (or ~/.zshrc) - done!
```

### Setup Shell Integration

```bash
# Auto-detect and install for your shell
lost install

# Or specify explicitly
lost install powershell
lost install bash
lost install zsh
```

### Build from Source (Linux)

```bash
# 1) Clone
git clone https://github.com/korybantes/lost-commands.git
cd lost-commands

# 2) Build Linux binary
chmod +x scripts/build-linux.sh
./scripts/build-linux.sh

# 3) Run directly
./lost-linux-amd64 --help

# Optional: install globally to /usr/local/bin/lost
./scripts/build-linux.sh --install
```

## Usage

```bash
# Search for git commands
lost search git

# Search with tag filter
lost search --tag docker

# Search and immediately run the most recent match
lost search --tag docker -r

# Show recent commands
lost recent
lost recent 50

# List all tags
lost tags

# Add a command manually
lost add "docker-compose up -d" --tag deploy

# Tag an existing command
lost tag 123 production

# Run most recent command by tag
lost run deploy
```

## Auto-Tagging Rules

Commands are automatically tagged based on:

| Command Prefix | Tags |
|---------------|------|
| `git *` | git, vcs |
| `docker *`, `docker-compose *` | docker, containers, devops |
| `kubectl *`, `helm *` | kubernetes, k8s, devops |
| `npm *`, `yarn *`, `pnpm *` | nodejs, package-manager |
| `python *`, `pip *`, `pytest *` | python, package-manager |
| `cargo *`, `rustc *` | rust, cargo |
| `go *` | golang, go |
| `make *`, `cmake *` | build |
| `ssh *`, `scp *` | ssh, remote |
| File extensions (.py, .js, .go, etc.) | Detected automatically |

## Database Location

Commands are stored in a local SQLite database:
- Windows: `%USERPROFILE%\.lost\commands.db`
- macOS/Linux: `~/.lost/commands.db`

## Commands Reference

| Command | Description |
|---------|-------------|
| `lost install [shell]` | Install shell integration |
| `lost add <cmd>` | Manually add a command |
| `lost capture <cmd> [dir]` | Internal: called by shell hooks |
| `lost search [query]` | Search commands |
| `lost run <tag>` | Run most recent command by tag |
| `lost recent [n]` | Show recent commands |
| `lost tag <id> <tag>` | Add tag to command |
| `lost tags` | List all tags |

## Credits

Created by **Ertac Toptutan** ([@korybantes](https://github.com/korybantes))

## Uninstall

Remove the hook from your shell profile:
- PowerShell: Edit `$PROFILE` and remove the "# Lost" block
- Bash: Edit `~/.bashrc` and remove the "# Lost" block
- Zsh: Edit `~/.zshrc` and remove the "# Lost" block

Then delete the database: `~/.lost/commands.db`
