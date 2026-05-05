# driftwatch

Lightweight daemon that detects config file changes across services and alerts via webhook.

---

## Installation

```bash
go install github.com/yourorg/driftwatch@latest
```

Or build from source:

```bash
git clone https://github.com/yourorg/driftwatch.git && cd driftwatch && go build -o driftwatch .
```

---

## Usage

Create a `driftwatch.yaml` configuration file:

```yaml
webhook: "https://hooks.example.com/alert"
interval: 30s
paths:
  - /etc/nginx/nginx.conf
  - /etc/app/config.toml
  - /etc/systemd/system/myservice.service
```

Run the daemon:

```bash
driftwatch --config driftwatch.yaml
```

When a monitored file changes, driftwatch sends a POST request to your webhook with details about the file, the timestamp, and a checksum diff.

**Example webhook payload:**

```json
{
  "file": "/etc/nginx/nginx.conf",
  "changed_at": "2024-11-03T14:22:01Z",
  "previous_hash": "a3f1c2...",
  "current_hash": "b9e4d7..."
}
```

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--config` | `./driftwatch.yaml` | Path to config file |
| `--interval` | `30s` | Poll interval |
| `--dry-run` | `false` | Log changes without sending alerts |

---

## License

MIT © yourorg