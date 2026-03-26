package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Command struct {
	ID        int64     `json:"id"`
	Command   string    `json:"command"`
	Tags      []string  `json:"tags"`
	Timestamp time.Time `json:"timestamp"`
	Directory string    `json:"directory"`
	Shell     string    `json:"shell"`
}

type Database struct {
	db *sql.DB
}

func New(path string) (*Database, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	d := &Database{db: db}
	if err := d.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return d, nil
}

func (d *Database) migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS commands (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		tags TEXT DEFAULT '[]',
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
		directory TEXT,
		shell TEXT,
		UNIQUE(command, directory, timestamp)
	);
	CREATE INDEX IF NOT EXISTS idx_commands_command ON commands(command);
	CREATE INDEX IF NOT EXISTS idx_commands_timestamp ON commands(timestamp);
	`
	_, err := d.db.Exec(query)
	return err
}

func (d *Database) Close() error {
	return d.db.Close()
}

func (d *Database) AddCommand(cmd string, tags []string, directory, shell string) (*Command, error) {
	tagsJSON, err := json.Marshal(tags)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal tags: %w", err)
	}

	result, err := d.db.Exec(
		"INSERT INTO commands (command, tags, directory, shell) VALUES (?, ?, ?, ?)",
		cmd, string(tagsJSON), directory, shell,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert command: %w", err)
	}

	id, _ := result.LastInsertId()
	return &Command{
		ID:        id,
		Command:   cmd,
		Tags:      tags,
		Timestamp: time.Now(),
		Directory: directory,
		Shell:     shell,
	}, nil
}

func (d *Database) Search(query string, tags []string) ([]Command, error) {
	var args []interface{}
	sqlQuery := "SELECT id, command, tags, timestamp, directory, shell FROM commands WHERE 1=1"

	if query != "" {
		sqlQuery += " AND command LIKE ?"
		args = append(args, "%"+query+"%")
	}

	if len(tags) > 0 {
		for _, tag := range tags {
			sqlQuery += " AND tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
	}

	sqlQuery += " ORDER BY timestamp DESC LIMIT 50"

	rows, err := d.db.Query(sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to search: %w", err)
	}
	defer rows.Close()

	return d.scanCommands(rows)
}

func (d *Database) GetRecent(limit int) ([]Command, error) {
	rows, err := d.db.Query(
		"SELECT id, command, tags, timestamp, directory, shell FROM commands ORDER BY timestamp DESC LIMIT ?",
		limit,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent: %w", err)
	}
	defer rows.Close()

	return d.scanCommands(rows)
}

func (d *Database) GetByID(id int64) (*Command, error) {
	row := d.db.QueryRow(
		"SELECT id, command, tags, timestamp, directory, shell FROM commands WHERE id = ?",
		id,
	)

	var cmd Command
	var tagsJSON string
	err := row.Scan(&cmd.ID, &cmd.Command, &tagsJSON, &cmd.Timestamp, &cmd.Directory, &cmd.Shell)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal([]byte(tagsJSON), &cmd.Tags)
	return &cmd, nil
}

func (d *Database) TagCommand(id int64, newTag string) error {
	cmd, err := d.GetByID(id)
	if err != nil {
		return err
	}
	if cmd == nil {
		return fmt.Errorf("command not found")
	}

	// Check if tag already exists
	for _, t := range cmd.Tags {
		if t == newTag {
			return nil // Already tagged
		}
	}

	cmd.Tags = append(cmd.Tags, newTag)
	tagsJSON, _ := json.Marshal(cmd.Tags)

	_, err = d.db.Exec("UPDATE commands SET tags = ? WHERE id = ?", string(tagsJSON), id)
	return err
}

func (d *Database) GetAllTags() ([]string, error) {
	rows, err := d.db.Query("SELECT tags FROM commands")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tagSet := make(map[string]bool)
	for rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&tagsJSON); err != nil {
			continue
		}
		var tags []string
		json.Unmarshal([]byte(tagsJSON), &tags)
		for _, t := range tags {
			tagSet[t] = true
		}
	}

	var tags []string
	for t := range tagSet {
		tags = append(tags, t)
	}
	return tags, nil
}

func (d *Database) GetByTag(tag string) (*Command, error) {
	row := d.db.QueryRow(
		"SELECT id, command, tags, timestamp, directory, shell FROM commands WHERE tags LIKE ? ORDER BY timestamp DESC LIMIT 1",
		"%"+tag+"%",
	)

	var cmd Command
	var tagsJSON string
	err := row.Scan(&cmd.ID, &cmd.Command, &tagsJSON, &cmd.Timestamp, &cmd.Directory, &cmd.Shell)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	json.Unmarshal([]byte(tagsJSON), &cmd.Tags)
	return &cmd, nil
}

func (d *Database) scanCommands(rows *sql.Rows) ([]Command, error) {
	var commands []Command
	for rows.Next() {
		var cmd Command
		var tagsJSON string
		if err := rows.Scan(&cmd.ID, &cmd.Command, &tagsJSON, &cmd.Timestamp, &cmd.Directory, &cmd.Shell); err != nil {
			continue
		}
		json.Unmarshal([]byte(tagsJSON), &cmd.Tags)
		commands = append(commands, cmd)
	}
	return commands, nil
}
