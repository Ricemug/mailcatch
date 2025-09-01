# FakeSMTP

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
chmod +x fakesmtp-darwin-arm64
./fakesmtp-darwin-arm64

# Linux
chmod +x fakesmtp-linux-amd64
./fakesmtp-linux-amd64

# Windows
fakesmtp-windows-amd64.exe
```

### Docker/Podman

```bash
# Quick start
docker run -p 2525:2525 -p 8080:8080 fakesmtp:latest

# With persistent storage
docker run -p 2525:2525 -p 8080:8080 \
  -v ./data:/app/data \
  -v ./logs:/app/logs \
  fakesmtp:latest

# Podman (same commands)
podman run -p 2525:2525 -p 8080:8080 fakesmtp:latest
```

### Access

- 🌐 **Web UI**: http://localhost:8080
- 📧 **SMTP**: localhost:2525

## Configuration

### Command Line Options

```bash
./fakesmtp [OPTIONS]

Options:
  --smtp-port=2525              SMTP server port
  --http-port=8080              Web UI port  
  --db-path=./data/emails.db    Database file path
  --log-path=/tmp/fakesmtp.log  Log file path
  --clear-on-shutdown=true      Clear emails on shutdown
  --daemon=false                Run in background mode
  --help                        Show help
```

### Environment Variables

```bash
export SMTP_PORT=1025
export HTTP_PORT=3000
export LOG_PATH=/var/log/fakesmtp.log
export CLEAR_ON_SHUTDOWN=false
export DAEMON=true
```

### Usage Examples

```bash
# Basic usage
./fakesmtp

# Custom ports
./fakesmtp --smtp-port=1025 --http-port=3000

# Background mode
./fakesmtp --daemon --log-path=/var/log/fakesmtp.log

# Keep emails between restarts
./fakesmtp --clear-on-shutdown=false
```

## Sending Test Emails

### Python

```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText("Hello from FakeSMTP!")
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

## Docker Usage

### Docker Compose

```yaml
version: '3.8'
services:
  fakesmtp:
    image: fakesmtp:latest
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
./fakesmtp --smtp-port=1025
```

### Permission Issues

```bash
# Make executable
chmod +x fakesmtp-*

# Fix Docker volumes
sudo chown -R 1000:1000 ./data ./logs
```

### Database Issues

```bash
# Clear database
rm -f data/emails.bolt

# Check logs
tail -f /tmp/fakesmtp.log
```

## License

MIT License - see [LICENSE](LICENSE) file for details.

---

**🌟 Star this project if it helped you!**