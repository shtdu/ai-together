// Copyright (c) 2025 Code Together
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.


package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/code-together/shared/tools/hook-collector/commands"
	"github.com/code-together/shared/tools/hook-collector/internal/config"
)

var (
	// Global flags
	dbPath     string
	configPath string
	toolName   string
	verbose    bool
	quiet      bool

	// Collect flags
	collectOverwrite bool

	// Export flags
	exportFormat    string
	exportSince     string
	exportUntil     string
	exportSessionID string
	exportEventType string
	exportCompress  bool

	// Clean flags
	cleanOlderThan string
	cleanDryRun    bool
	cleanCompact   bool
)

func main() {
	// Global flags
	flag.StringVar(&dbPath, "db-path", "", "Path to SQLite database")
	flag.StringVar(&configPath, "config", "", "Path to config file")
	flag.StringVar(&toolName, "tool-name", "", "Tool name for bucket isolation (e.g., claude, codex)")
	flag.BoolVar(&verbose, "verbose", false, "Enable verbose logging")
	flag.BoolVar(&quiet, "quiet", false, "Suppress non-error output")

	// Collect flags
	flag.BoolVar(&collectOverwrite, "overwrite", false, "Overwrite existing events")

	// Export flags
	flag.StringVar(&exportFormat, "format", "jsonl", "Export format: json, jsonl, csv")
	flag.StringVar(&exportSince, "since", "", "Export events since this time")
	flag.StringVar(&exportUntil, "until", "", "Export events until this time")
	flag.StringVar(&exportSessionID, "session-id", "", "Export only this session")
	flag.StringVar(&exportEventType, "event-type", "", "Export only this event type")
	flag.BoolVar(&exportCompress, "compress", false, "Compress output with gzip")

	// Clean flags
	flag.StringVar(&cleanOlderThan, "older-than", "", "Remove events older than")
	flag.BoolVar(&cleanDryRun, "dry-run", false, "Show what would be deleted without deleting")
	flag.BoolVar(&cleanCompact, "compact", false, "Compact database after cleanup")

	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
		os.Exit(1)
	}

	command := flag.Arg(0)

	// Initialize config and get db path
	var dbPathResolved string
	cfg, dbP, err := config.InitConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}
	if dbPath == "" {
		dbPath = dbP
	}
	dbPathResolved = dbPath

	// Route to command handler
	switch command {
	case "collect":
		commands.RunCollectCommand(dbPathResolved, toolName, collectOverwrite, verbose)
	case "export":
		outputFile := ""
		if flag.NArg() >= 2 {
			outputFile = flag.Arg(1)
		}
		commands.RunExportCommand(dbPathResolved, toolName, outputFile, exportFormat, exportSince, exportUntil,
			exportSessionID, exportEventType, exportCompress)
	case "clean":
		commands.RunCleanCommand(dbPathResolved, toolName, cleanOlderThan, cleanDryRun, cleanCompact, cfg.RetentionDays)
	case "help", "-h", "--help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown command '%s'\n", command)
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `Hook Event Collector - Collect and manage Claude Code hook events

USAGE:
  hook-collector [global-flags] <command> [command-flags]

GLOBAL FLAGS:
  -db-path <path>       Path to sqlite database (default: $HOME/.code-together/hook-events.sqlite)
  -config <path>        Path to config file (default: $HOME/.code-together/hook-collector.json)
  -tool-name <name>     Tool name for bucket isolation (e.g., claude, codex). Uses default bucket if not specified.
  -verbose              Enable verbose logging
  -quiet                Suppress non-error output

COMMANDS:
  collect               Collect an event from stdin and store in database
  export [output-file]  Export events from database (default: stdout)
  clean                 Clean up old events from database

COLLECT FLAGS:
  -overwrite            Overwrite existing events (default: skip duplicates)

EXPORT FLAGS:
  -format <format>      Export format: json, jsonl, csv (default: jsonl)
  -since <time>         Export events since this time (e.g., 7d, 30d, 2024-01-01, 2024-01-01T10:00:00Z)
  -until <time>         Export events until this time
  -session-id <id>      Export only this session
  -event-type <type>    Export only this event type
  -compress             Compress output with gzip

CLEAN FLAGS:
  -older-than <time>    Remove events older than (e.g., 30d, 1y)
  -dry-run              Show what would be deleted without deleting
  -compact              Compact database after cleanup

EXAMPLES:
  # Collect event from file (uses default bucket)
  cat event.json | hook-collector collect

  # Collect event from Claude Code hook (uses sessions.claude bucket)
  # In ~/.claude/settings.json: "command": "jq -cM | hook-collector --tool-name claude collect"

  # Export all events from default bucket to stdout
  hook-collector export

  # Export events from claude bucket
  hook-collector --tool-name claude export

  # Export last 7 days to file from claude bucket
  hook-collector --tool-name claude --since 7d export events.jsonl

  # Export events from a specific date
  hook-collector --since 2024-01-01 export

  # Export events from a specific datetime
  hook-collector --since "2024-01-01T10:00:00Z" export

  # Export last 30 days, compressed
  hook-collector --since 30d --compress export events.jsonl.gz

  # Export specific event type to stdout
  hook-collector --event-type UserPromptSubmit export | jq '.prompt'

  # Clean events older than 90 days from claude bucket
  hook-collector --tool-name claude clean --older-than 90d --compact

`)
}
