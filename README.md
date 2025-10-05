# MailCatch

A lightweight, cross-platform fake SMTP server for email testing and development.

![Go Version](https://img.shields.io/badge/go-1.21+-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

## Features

- 🚀 **Lightweight & Fast** - Single binary (~8MB) with zero dependencies
- 🌐 **Cross-platform** - Windows, macOS (Intel/Apple Silicon), Linux (x64/ARM64)
- 📧 **SMTP Server** - Receive emails on any port (default: 2525)
- 🖥️ **Web Interface** - Modern React-based UI for viewing emails
- ⚡ **Real-time Updates** - WebSocket integration for instant notifications
- 💾 **Persistent Storage** - BoltDB embedded database (no CGO required)
- 🐳 **Docker Ready** - Pre-built Docker/Podman images (30MB)
- 🧹 **Auto Cleanup** - Automatically clears test emails on shutdown
- 🌙 **Daemon Mode** - Run in background with custom logging
- 🔧 **Flexible Config** - Command-line flags and environment variables

## Quick Start

### Download Binary

| Platform | Architecture | Size |
|----------|-------------|------|
| macOS | Apple Silicon (M1/M2/M3) | 8.3MB |
| macOS | Intel | 8.8MB |
| Linux | x64 | 8.6MB |
| Linux | ARM64 | 8.1MB |
| Windows | x64 | 8.9MB |

```bash
# macOS
chmod +x mailcatch-darwin-arm64
./mailcatch-darwin-arm64

# Linux
chmod +x mailcatch-linux-amd64
./mailcatch-linux-amd64

# Windows
mailcatch-windows-amd64.exe
```

### Docker/Podman

```bash
# Quick start
docker run -p 2525:2525 -p 8080:8080 mailcatch:latest

# With persistent storage
docker run -p 2525:2525 -p 8080:8080 \
  -v ./data:/app/data \
  -v ./logs:/app/logs \
  mailcatch:latest

# Podman (same commands)
podman run -p 2525:2525 -p 8080:8080 mailcatch:latest

# Podman + Systemd (rootless, recommended)
./scripts/setup-podman-systemd.sh
```

### Access

- 🌐 **Web UI**: http://localhost:8080
- 📧 **SMTP**: localhost:2525

## Configuration

### Command Line Options

```bash
./mailcatch [OPTIONS]

Options:
  --smtp-port=2525              SMTP server port
  --http-port=8080              Web UI port  
  --db-path=./data/emails.db    Database file path
  --log-path=/tmp/mailcatch.log  Log file path
  --clear-on-shutdown=true      Clear emails on shutdown
  --daemon=false                Run in background mode
  --help                        Show help
```

### Environment Variables

```bash
export SMTP_PORT=1025
export HTTP_PORT=3000
export LOG_PATH=/var/log/mailcatch.log
export CLEAR_ON_SHUTDOWN=false
export DAEMON=true
```

### Usage Examples

```bash
# Basic usage
./mailcatch

# Custom ports
./mailcatch --smtp-port=1025 --http-port=3000

# Background mode
./mailcatch --daemon --log-path=/var/log/mailcatch.log

# Keep emails between restarts
./mailcatch --clear-on-shutdown=false
```

## Sending Test Emails

### Python

```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText("Hello from MailCatch!")
msg['Subject'] = 'Test Email'
msg['From'] = 'sender@example.com'
msg['To'] = 'recipient@example.com'

with smtplib.SMTP('localhost', 2525) as server:
    server.send_message(msg)
print("Email sent!")
```

### Node.js

```javascript
const nodemailer = require('nodemailer');

const transporter = nodemailer.createTransporter({
  host: 'localhost',
  port: 2525,
  secure: false,
  auth: false
});

transporter.sendMail({
  from: 'sender@example.com',
  to: 'recipient@example.com',
  subject: 'Test Email',
  text: 'Hello from Node.js!'
});
```

### cURL/Telnet

```bash
telnet localhost 2525
# Commands:
HELO localhost
MAIL FROM:<sender@example.com>
RCPT TO:<recipient@example.com>
DATA
Subject: Test Email

This is a test!
.
QUIT
```

## Docker/Podman Usage

### Docker Compose

```yaml
version: '3.8'
services:
  mailcatch:
    image: mailcatch:latest
    ports:
      - "2525:2525"
      - "8080:8080"
    volumes:
      - ./data:/app/data
      - ./logs:/app/logs
    environment:
      - CLEAR_ON_SHUTDOWN=true
    restart: unless-stopped
```

Run: `docker-compose up -d`

### Podman + Systemd (Recommended)

Using rootless containers with systemd service management:

```bash
# Quick setup
./scripts/setup-podman-systemd.sh

# Custom configuration
./scripts/setup-podman-systemd.sh --smtp-port 1025 --web-port 3000

# Enable boot startup
sudo loginctl enable-linger $USER
```

Detailed documentation: [PODMAN_SYSTEMD.md](PODMAN_SYSTEMD.md)

## API Reference

### REST API

- `GET /api/emails` - List emails
- `GET /api/emails/:id` - Get email details
- `DELETE /api/emails/:id` - Delete email
- `DELETE /api/emails` - Clear all emails
- `GET /api/stats` - Server statistics

### WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'new_email') {
    console.log('New email:', message.data);
  }
};
```

## Troubleshooting

### Port Issues

```bash
# Check port usage
lsof -i :2525
netstat -tulpn | grep 2525

# Use different port
./mailcatch --smtp-port=1025
```

### Permission Issues

```bash
# Make executable
chmod +x mailcatch-*

# Fix Docker volumes
sudo chown -R 1000:1000 ./data ./logs
```

### Database Issues

```bash
# Clear database
rm -f data/emails.bolt

# Check logs
tail -f /tmp/mailcatch.log
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

## Support This Project

If MailCatch has been helpful for your development workflow, please consider supporting its development:

[![Ko-fi](https://img.shields.io/badge/Ko--fi-Support-ff5f5f?logo=ko-fi)](https://ko-fi.com/ivanh0906)

### Other Ways to Support

- ⭐ **Star this repository** on GitHub
- 🐛 **Report bugs** and request features
- 🤝 **Contribute code** - see [CONTRIBUTING.md](.github/CONTRIBUTING.md)
- 📢 **Share** with your team and community

Your support helps maintain and improve MailCatch for the entire community!

---

**🌟 Star this project if it helped you!**