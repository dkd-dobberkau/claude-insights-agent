# Claude Insights Agent

Local agent for syncing Claude Code sessions to the dkd team server.

## Installation

### Quick Install (from releases)

```bash
curl -fsSL https://insights.dkd.internal/install.sh | bash
```

### Build from Source

```bash
git clone https://github.com/dkd/claude-insights-agent.git
cd claude-insights-agent
make install
```

## Usage

### Initialize Configuration

```bash
claude-insights-agent init
```

This creates `~/.config/claude-insights/config.yaml` with your settings.

### Start Continuous Sync

```bash
claude-insights-agent run
```

Runs as a daemon, syncing new sessions every 5 minutes (configurable).

### One-time Sync

```bash
claude-insights-agent sync
```

### Check Status

```bash
claude-insights-agent status
```

## Configuration

Config file: `~/.config/claude-insights/config.yaml`

```yaml
server:
  url: https://insights.dkd.internal
  api_key: dkd_sk_your_api_key_here

sharing:
  level: metadata          # none | metadata | full
  exclude_projects:
    - "**/personal/**"
    - "**/secret-*"
  anonymize_paths: true

sync:
  interval: 300            # Sync every 5 minutes
  retry_attempts: 3

logging:
  level: info
  file: ~/.local/log/claude-insights-agent.log
```

### Share Levels

| Level | What's Shared |
|-------|---------------|
| `none` | Nothing (agent paused) |
| `metadata` | Session stats, token counts, tool names, tags, project name |
| `full` | Everything including message content |

### Excluding Projects

Use glob patterns to exclude sensitive projects:

```yaml
sharing:
  exclude_projects:
    - "**/personal/**"      # Exclude all paths containing 'personal'
    - "**/secret-*"         # Exclude paths with 'secret-' prefix
    - "/Users/me/private/*" # Exclude specific directory
```

## Running as a Service (macOS)

Create `~/Library/LaunchAgents/com.dkd.claude-insights-agent.plist`:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>com.dkd.claude-insights-agent</string>
    <key>ProgramArguments</key>
    <array>
        <string>/Users/YOUR_USERNAME/.local/bin/claude-insights-agent</string>
        <string>run</string>
    </array>
    <key>RunAtLoad</key>
    <true/>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/claude-insights-agent.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/claude-insights-agent.err</string>
</dict>
</plist>
```

Load the service:

```bash
launchctl load ~/Library/LaunchAgents/com.dkd.claude-insights-agent.plist
```

## Running as a Service (Linux systemd)

Create `~/.config/systemd/user/claude-insights-agent.service`:

```ini
[Unit]
Description=Claude Insights Agent
After=network.target

[Service]
Type=simple
ExecStart=%h/.local/bin/claude-insights-agent run
Restart=always
RestartSec=10

[Install]
WantedBy=default.target
```

Enable and start:

```bash
systemctl --user enable claude-insights-agent
systemctl --user start claude-insights-agent
```

## Development

```bash
# Build
make build

# Run tests
make test

# Build all platforms
make release
```

## Architecture

```
~/.claude/projects/
    └── -path-to-project/
        └── session-id.jsonl
                │
                ▼
        ┌───────────────┐
        │ JSONL Parser  │
        └───────────────┘
                │
                ▼
        ┌───────────────┐
        │ Privacy Filter│ (apply share level)
        └───────────────┘
                │
                ▼
        ┌───────────────┐
        │  API Client   │ ──► Team Server
        └───────────────┘
```

## Files

| Path | Purpose |
|------|---------|
| `~/.config/claude-insights/config.yaml` | Configuration |
| `~/.local/state/claude-insights/synced.json` | Sync state |
| `~/.local/log/claude-insights-agent.log` | Logs (if configured) |
