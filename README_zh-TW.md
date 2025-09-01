# FakeSMTP

輕量級跨平台假 SMTP 伺服器，專為郵件測試與開發而設計。

![Go Version](https://img.shields.io/badge/go-1.21+-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Platform](https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-lightgrey)

## 功能特色

- 🚀 **輕量高效** - 單一執行檔 (~8MB)，零外部依賴
- 🌐 **跨平台** - 支援 Windows、macOS (Intel/Apple Silicon)、Linux (x64/ARM64)
- 📧 **SMTP 伺服器** - 可配置埠號接收郵件 (預設: 2525)
- 🖥️ **Web 介面** - 現代化 React UI 查看郵件
- ⚡ **即時更新** - WebSocket 整合，即時顯示新郵件
- 💾 **持久儲存** - 內建 BoltDB 資料庫 (無需 CGO)
- 🐳 **Docker 支援** - 預建 Docker/Podman 映像檔 (30MB)
- 🧹 **自動清理** - 程式停止時自動清空測試郵件
- 🌙 **背景模式** - 可在背景執行並自訂日誌位置
- 🔧 **靈活配置** - 命令列參數和環境變數配置

## 快速開始

### 下載執行檔

| 平台 | 架構 | 大小 |
|------|-----|------|
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
# 快速啟動
docker run -p 2525:2525 -p 8080:8080 fakesmtp:latest

# 持久化資料
docker run -p 2525:2525 -p 8080:8080 \
  -v ./data:/app/data \
  -v ./logs:/app/logs \
  fakesmtp:latest

# Podman (相同指令)
podman run -p 2525:2525 -p 8080:8080 fakesmtp:latest
```

### 存取方式

- 🌐 **Web 介面**: http://localhost:8080
- 📧 **SMTP 埠號**: localhost:2525

## 配置選項

### 命令列參數

```bash
./fakesmtp [選項]

選項:
  --smtp-port=2525              SMTP 伺服器埠號
  --http-port=8080              Web 介面埠號
  --db-path=./data/emails.db    資料庫檔案路徑
  --log-path=/tmp/fakesmtp.log  日誌檔案路徑
  --clear-on-shutdown=true      程式停止時清空郵件
  --daemon=false                背景執行模式
  --help                        顯示幫助資訊
```

### 環境變數

```bash
export SMTP_PORT=1025
export HTTP_PORT=3000
export LOG_PATH=/var/log/fakesmtp.log
export CLEAR_ON_SHUTDOWN=false
export DAEMON=true
```

### 使用範例

```bash
# 基本使用
./fakesmtp

# 自訂埠號
./fakesmtp --smtp-port=1025 --http-port=3000

# 背景執行
./fakesmtp --daemon --log-path=/var/log/fakesmtp.log

# 重啟時保留郵件
./fakesmtp --clear-on-shutdown=false
```

## 發送測試郵件

### Python 範例

```python
import smtplib
from email.mime.text import MIMEText

msg = MIMEText("來自 FakeSMTP 的問候！")
msg['Subject'] = '測試郵件'
msg['From'] = 'sender@example.com'
msg['To'] = 'recipient@example.com'

with smtplib.SMTP('localhost', 2525) as server:
    server.send_message(msg)
print("郵件發送成功！")
```

### Node.js 範例

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
  subject: '測試郵件',
  text: '來自 Node.js 的問候！'
});
```

### cURL/Telnet 範例

```bash
telnet localhost 2525
# 指令:
HELO localhost
MAIL FROM:<sender@example.com>
RCPT TO:<recipient@example.com>
DATA
Subject: 測試郵件

這是一封測試郵件！
.
QUIT
```

## Docker 使用

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

執行: `docker-compose up -d`

## API 參考

### REST API

- `GET /api/emails` - 列出郵件
- `GET /api/emails/:id` - 取得郵件詳情
- `DELETE /api/emails/:id` - 刪除郵件
- `DELETE /api/emails` - 清空所有郵件
- `GET /api/stats` - 伺服器統計

### WebSocket

```javascript
const ws = new WebSocket('ws://localhost:8080/ws');
ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  if (message.type === 'new_email') {
    console.log('新郵件:', message.data);
  }
};
```

## 故障排除

### 埠號問題

```bash
# 檢查埠號使用狀況
lsof -i :2525
netstat -tulpn | grep 2525

# 使用不同埠號
./fakesmtp --smtp-port=1025
```

### 權限問題

```bash
# 設定執行權限
chmod +x fakesmtp-*

# 修正 Docker volume 權限
sudo chown -R 1000:1000 ./data ./logs
```

### 資料庫問題

```bash
# 清除資料庫
rm -f data/emails.bolt

# 檢查日誌
tail -f /tmp/fakesmtp.log
```

## 授權條款

MIT 授權 - 詳見 [LICENSE](LICENSE) 檔案。

---

**🌟 如果這個專案對你有幫助，請給個星星！**