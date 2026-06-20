# obs

Open-source observability stack CLI — a self-hosted Datadog alternative.

Bundles SigNoz (APM, traces, logs, metrics) + PostHog (session replay, analytics) into a single installable binary.

## Install

```bash
# macOS / Linux (Homebrew)
brew install osillation/tap/obs

# Go toolchain
go install github.com/osillation/obs@latest

# curl (no Go required)
curl -fsSL https://get.obs.sh | sh
```

## Quick Start (local dashboard)

```bash
obs dashboard start
# Open http://localhost:3301 (SigNoz) and http://localhost:8000 (PostHog)
# Set in your app: OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4318
```

## Deploy to a client server

```bash
obs init <client>                  # interactive setup wizard
obs certs init-ca --client <name>  # create mTLS CA
obs certs add ibrahim --client <name>
obs deploy --client <name> --project /path/to/project
```

## AWS CloudWatch integration

```bash
obs cloudwatch connect --client <name>
obs cloudwatch status  --client <name>
```

## Commands

| Command | Description |
|---------|-------------|
| `obs dashboard start` | Run full stack locally |
| `obs init <client>` | Scaffold client config |
| `obs detect <path>` | Auto-detect project stack |
| `obs validate` | Pre-deploy check |
| `obs deploy` | Full deployment |
| `obs instrument` | Copy instrumentation templates |
| `obs certs init-ca` | Create mTLS CA |
| `obs certs add <name>` | Issue employee cert |
| `obs cloudwatch connect` | Connect AWS CloudWatch |

## Platform

Windows requires WSL. Native Windows support planned for v2.

## License

MIT
