# Claude Code Hook Events

This directory contains the Go structs for the Claude Code hook events.

## Event Types

- `Stop`: Runs when Claude Code finishes responding
- `PreToolUse`: Runs before tool calls (can block them)
- `PostToolUse`: Runs after tool calls
- `SessionEnd`: Runs when the session ends
- `SubagentStop`: Runs when subagent tasks complete
- `UserPromptSubmit`: Runs when user submits a prompt

## How to Update

1. Run `go run` in the `ct_member/shared/tools/hook-schema-analyzer` directory to generate the Go structs.

## How to collect the hook raw events.

In your local claude code, enable hooks for all events to a file.

```json
{
    "PreToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ],
    "PostToolUse": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ],
    "SessionEnd": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ],
    "UserPromptSubmit": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ],
    "PreCompact": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ],
    "SubagentStop": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ],
    "Stop": [
      {
        "matcher": "*",
        "hooks": [
          {
            "type": "command",
            "command": "jq . >> ~/.claude/job_done.log"
          }
        ]
      }
    ]
  }
```

## Reference

1. [Claude Code hooks guide](https://code.claude.com/docs/en/hooks-guide), [Markdown Copy](https://code.claude.com/docs/en/hooks-guide.md)