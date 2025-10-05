# Podman Rootless Systemd Deployment Guide

This guide explains how to deploy MailCatch using Podman rootless containers and systemd user services.

## Features

- **Rootless Containers**: No root privileges required, enhanced security
- **Systemd Integration**: Use systemd to manage container lifecycle
- **Auto Restart**: Automatically restart service on failure
- **Boot Startup**: Optional automatic startup on boot
- **Log Management**: Integrated with journald logging
- **Health Checks**: Built-in container health monitoring

## Prerequisites

### System Requirements
- Linux distribution (Fedora, RHEL, Ubuntu, Debian, etc.)
- systemd (pre-installed on most modern Linux distributions)
- Podman 3.0+

### Install Podman

#### Fedora/RHEL/CentOS
```bash
sudo dnf install podman
```

#### Ubuntu/Debian
```bash
sudo apt update
sudo apt install podman
```

#### Arch Linux
```bash
sudo pacman -S podman
```

## Quick Start

### 1. Basic Installation
```bash
# Use default settings
./scripts/setup-podman-systemd.sh
```

### 2. Custom Configuration Installation
```bash
# Customize ports and directories
./scripts/setup-podman-systemd.sh \
    --smtp-port 1025 \
    --web-port 3000 \
    --data-dir "$HOME/mailcatch-data"
```

### 3. Enable Boot Startup
```bash
# Allow user services to start at boot (even when user is not logged in)
sudo loginctl enable-linger $USER
```

## Script Options

```bash
Options:
  -s, --smtp-port PORT     SMTP port (default: 2525)
  -w, --web-port PORT      Web UI port (default: 8080)
  -d, --data-dir PATH      Data directory (default: ~/.local/share/mailcatch)
  -l, --log-dir PATH       Log directory (default: ~/.local/share/mailcatch/logs)
  -n, --name NAME          Container name (default: mailcatch)
  -i, --image IMAGE        Container image (default: mailcatch:latest)
  -h, --help               Show help information
  --uninstall              Remove service and container
```

## Service Management

### Basic Commands
```bash
# View service status
systemctl --user status mailcatch.service

# Start service
systemctl --user start mailcatch.service

# Stop service
systemctl --user stop mailcatch.service

# Restart service
systemctl --user restart mailcatch.service

# Enable boot startup
systemctl --user enable mailcatch.service

# Disable boot startup
systemctl --user disable mailcatch.service
```

### View Logs
```bash
# View all logs
journalctl --user -u mailcatch.service

# View logs in real-time
journalctl --user -u mailcatch.service -f

# View recent logs
journalctl --user -u mailcatch.service --since "1 hour ago"

# View container internal logs
podman logs mailcatch
```

## Troubleshooting

### Service Won't Start

1. Check service status:
```bash
systemctl --user status mailcatch.service
```

2. View detailed logs:
```bash
journalctl --user -u mailcatch.service --no-pager
```

3. Check container status:
```bash
podman ps -a
podman logs mailcatch
```

### Port Conflicts

If ports are occupied, reinstall with different ports:
```bash
./scripts/setup-podman-systemd.sh --uninstall
./scripts/setup-podman-systemd.sh --smtp-port 1025 --web-port 3000
```

### Permission Issues

Ensure data directory permissions are correct:
```bash
ls -la ~/.local/share/mailcatch
# Should be owned by your user
```

### SELinux Issues (Fedora/RHEL)

If encountering SELinux permission problems:
```bash
# Check SELinux status
getenforce

# Temporarily disable SELinux (not recommended)
sudo setenforce 0

# Or set appropriate SELinux context
setsebool -P container_manage_cgroup true
```

## Advanced Configuration

### Custom Environment Variables

Edit service file to add more environment variables:
```bash
# Edit service file
systemctl --user edit mailcatch.service
```

Add content:
```ini
[Service]
Environment=CLEAR_ON_SHUTDOWN=false
Environment=DAEMON=true
```

### Network Configuration

By default, the container uses host network. For custom network:

1. Create Podman network:
```bash
podman network create mailcatch-network
```

2. Modify `--network` parameter in service file

### Resource Limits

Add resource limits in service file:
```bash
# Edit service file
systemctl --user edit mailcatch.service
```

Add content:
```ini
[Service]
Environment=PODMAN_EXTRA_ARGS=--memory=256m --cpus=1.0
```

## Security Considerations

### Rootless Advantages
- Container runs as unprivileged user
- Reduced attack surface
- No sudo privileges required

### Recommended Security Settings
1. Bind only to localhost (default behavior)
2. Use firewall to restrict access
3. Regularly update container images
4. Monitor log files

### Firewall Setup
```bash
# Only allow local connections (default)
# If external access is needed, configure firewall:

# firewalld (Fedora/RHEL)
sudo firewall-cmd --add-port=2525/tcp --permanent
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload

# ufw (Ubuntu)
sudo ufw allow 2525/tcp
sudo ufw allow 8080/tcp
```

## Remove Service

Completely remove MailCatch service:
```bash
./scripts/setup-podman-systemd.sh --uninstall
```

This will:
- Stop and disable systemd service
- Remove service file
- Stop and delete container
- Preserve data files (manual deletion required)

Manual data cleanup:
```bash
rm -rf ~/.local/share/mailcatch
```

## Comparison with Docker Compose

| Feature | Docker Compose | Podman + Systemd |
|---------|----------------|-------------------|
| Privilege Requirements | Needs Docker daemon (usually requires root) | Rootless |
| System Integration | Less | Deep systemd integration |
| Auto Restart | Via restart policy | Managed by systemd |
| Log Management | Docker log drivers | journald integration |
| Boot Startup | Needs extra configuration | Native systemd support |
| Resource Control | Docker limits | systemd + cgroups |

## References

- [Podman Official Documentation](https://podman.io/getting-started/)
- [systemd User Services](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [Podman systemd Integration](https://github.com/containers/podman/blob/main/docs/source/markdown/podman-generate-systemd.1.md)
