# Claude Insights Agent

Local agent that syncs Claude Code session data to a central dkd team server.

## Installation

```bash
# Build from source
go build -o claude-insights-agent ./cmd/agent

# Or download binary
curl -L https://insights.dkd.internal/install.sh | bash
```

## Configuration

Create `~/.config/claude-insights/config.yaml`:

```yaml
server:
  url: https://insights.dkd.internal
  api_key: dkd_sk_your_key_here

sharing:
  level: metadata  # none | metadata | full

sync:
  interval: 300  # seconds
```

## Usage

```bash
# Initialize config
claude-insights-agent init

# Run sync daemon
claude-insights-agent run

# Check status
claude-insights-agent status
```
