# GitHub to GitVerse Mirror

CLI tool for mirroring GitHub repositories (including private) to GitVerse with all branches and tags.

## Features

- Mirror all branches and tags
- Preserve repository privacy settings
- Sync on-demand or via cron
- Support for both creating new repositories and updating existing ones

## Requirements

- Go 1.21+
- Git
- GitHub Personal Access Token with `repo` scope
- GitVerse API token

## Installation

```bash
go build -o mirror ./cmd/mirror/
```

## Configuration

Create `config.yaml`:

```yaml
github:
  token: "${GITHUB_TOKEN}"  # PAT with repo scope

gitverse:
  token: "${GITVERSE_TOKEN}"
  base_url: "https://gitverse.ru/api/v1"

sync:
  timeout_minutes: 30

cron:
  enabled: false
  interval_hours: 6
```

Environment variables are supported via `${VAR_NAME}` syntax.

## Usage

### Sync all repositories

```bash
GITHUB_TOKEN=ghp_xxx GITVERSE_TOKEN=gvt_xxx ./mirror sync
```

Or with config file:

```bash
CONFIG_PATH=/path/to/config.yaml ./mirror sync
```

### Sync specific repository

```bash
GITHUB_TOKEN=ghp_xxx GITVERSE_TOKEN=gvt_xxx ./mirror sync repository-name
```

### List GitHub repositories

```bash
GITHUB_TOKEN=ghp_xxx ./mirror list
```

### Show differences between GitHub and GitVerse

```bash
GITHUB_TOKEN=ghp_xxx GITVERSE_TOKEN=gvt_xxx ./mirror diff
```

## Cron Setup

### systemd timer

Create `/etc/systemd/system/gh-mirror.service`:

```ini
[Unit]
Description=GitHub to GitVerse Mirror

[Service]
Type=oneshot
Environment=CONFIG_PATH=/path/to/config.yaml
ExecStart=/usr/local/bin/mirror sync
```

Create `/etc/systemd/system/gh-mirror.timer`:

```ini
[Unit]
Description=GitHub to GitVerse Mirror Timer

[Timer]
OnCalendar=hourly
Persistent=true

[Install]
WantedBy=timers.target
```

Enable and start:

```bash
systemctl enable gh-mirror.timer
systemctl start gh-mirror.timer
```

### Direct cron

Add to crontab:

```cron
0 */6 * * * /path/to/mirror sync
```

## GitHub Token Setup

1. Go to GitHub Settings → Developer settings → Personal access tokens → Fine-grained tokens
2. Create new token with:
   - Resource owner: your username
   - Permissions: Repository access → All repositories
   - Permissions: Contents → Read and write
3. Copy the token

## GitVerse Token Setup

1. Go to GitVerse → Profile → Settings → Tokens
2. Create new token
3. Copy the token

## How It Works

1. Lists all repositories from GitHub (including private)
2. For each repository:
   - Creates or updates repository on GitVerse (preserving privacy)
   - Uses `git clone --mirror` to clone the full repository
   - Uses `git push --mirror` to push all branches and tags to GitVerse
3. Logs repositories that exist on GitVerse but not on GitHub (never deletes)

## License

MIT
