# Podman Rootless Systemd 部署指南

這份指南說明如何使用 Podman rootless 容器和 systemd 用戶服務來部署 MailCatch。

## 功能特色

- **Rootless 容器**: 不需要 root 權限，提升安全性
- **Systemd 整合**: 使用 systemd 管理容器生命週期
- **自動重啟**: 服務失敗時自動重啟
- **開機啟動**: 可選的開機自動啟動
- **日誌管理**: 整合 journald 日誌
- **健康檢查**: 內建容器健康監控

## 前置需求

### 系統需求
- Linux 發行版 (Fedora, RHEL, Ubuntu, Debian 等)
- systemd (大多數現代 Linux 發行版預設安裝)
- Podman 3.0+

### 安裝 Podman

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

## 快速開始

### 1. 基本安裝
```bash
# 使用預設設置
./scripts/setup-podman-systemd.sh
```

### 2. 自訂配置安裝
```bash
# 自訂端口和目錄
./scripts/setup-podman-systemd.sh \
    --smtp-port 1025 \
    --web-port 3000 \
    --data-dir "$HOME/mailcatch-data"
```

### 3. 啟用開機自動啟動
```bash
# 允許用戶服務在開機時啟動（即使用戶未登入）
sudo loginctl enable-linger $USER
```

## 腳本選項

```bash
選項:
  -s, --smtp-port PORT     SMTP 端口 (預設: 2525)
  -w, --web-port PORT      Web UI 端口 (預設: 8080)
  -d, --data-dir PATH      數據目錄 (預設: ~/.local/share/mailcatch)
  -l, --log-dir PATH       日誌目錄 (預設: ~/.local/share/mailcatch/logs)
  -n, --name NAME          容器名稱 (預設: mailcatch)
  -i, --image IMAGE        容器映像 (預設: mailcatch:latest)
  -h, --help               顯示幫助信息
  --uninstall              移除服務和容器
```

## 服務管理

### 基本命令
```bash
# 查看服務狀態
systemctl --user status mailcatch.service

# 啟動服務
systemctl --user start mailcatch.service

# 停止服務
systemctl --user stop mailcatch.service

# 重啟服務
systemctl --user restart mailcatch.service

# 啟用開機啟動
systemctl --user enable mailcatch.service

# 禁用開機啟動
systemctl --user disable mailcatch.service
```

### 查看日誌
```bash
# 查看所有日誌
journalctl --user -u mailcatch.service

# 實時查看日誌
journalctl --user -u mailcatch.service -f

# 查看最近的日誌
journalctl --user -u mailcatch.service --since "1 hour ago"

# 查看容器內部日誌
podman logs mailcatch
```

## 故障排除

### 服務無法啟動

1. 檢查服務狀態:
```bash
systemctl --user status mailcatch.service
```

2. 查看詳細日誌:
```bash
journalctl --user -u mailcatch.service --no-pager
```

3. 檢查容器狀態:
```bash
podman ps -a
podman logs mailcatch
```

### 端口衝突

如果端口被佔用，使用不同端口重新安裝:
```bash
./scripts/setup-podman-systemd.sh --uninstall
./scripts/setup-podman-systemd.sh --smtp-port 1025 --web-port 3000
```

### 權限問題

確保數據目錄權限正確:
```bash
ls -la ~/.local/share/mailcatch
# 應該屬於你的用戶
```

### SELinux 問題 (Fedora/RHEL)

如果遇到 SELinux 權限問題:
```bash
# 檢查 SELinux 狀態
getenforce

# 暫時禁用 SELinux (不建議)
sudo setenforce 0

# 或設置適當的 SELinux 上下文
setsebool -P container_manage_cgroup true
```

## 進階配置

### 自訂環境變數

編輯服務檔案以添加更多環境變數:
```bash
# 編輯服務檔案
systemctl --user edit mailcatch.service
```

添加內容:
```ini
[Service]
Environment=CLEAR_ON_SHUTDOWN=false
Environment=DAEMON=true
```

### 網路配置

預設情況下，容器使用主機網路。如需自訂網路:

1. 創建 Podman 網路:
```bash
podman network create mailcatch-network
```

2. 修改服務檔案中的 `--network` 參數

### 資源限制

在服務檔案中添加資源限制:
```bash
# 編輯服務檔案
systemctl --user edit mailcatch.service
```

添加內容:
```ini
[Service]
Environment=PODMAN_EXTRA_ARGS=--memory=256m --cpus=1.0
```

## 安全考量

### Rootless 優勢
- 容器以非特權用戶運行
- 減少攻擊面
- 不需要 sudo 權限

### 建議的安全設置
1. 僅綁定到 localhost (預設行為)
2. 使用防火牆限制訪問
3. 定期更新容器映像
4. 監控日誌文件

### 防火牆設置
```bash
# 僅允許本地連接 (預設)
# 如需從外部訪問，配置防火牆:

# firewalld (Fedora/RHEL)
sudo firewall-cmd --add-port=2525/tcp --permanent
sudo firewall-cmd --add-port=8080/tcp --permanent
sudo firewall-cmd --reload

# ufw (Ubuntu)
sudo ufw allow 2525/tcp
sudo ufw allow 8080/tcp
```

## 移除服務

完全移除 MailCatch 服務:
```bash
./scripts/setup-podman-systemd.sh --uninstall
```

這將:
- 停止並禁用 systemd 服務
- 移除服務檔案
- 停止並刪除容器
- 保留數據檔案 (需手動刪除)

手動清理數據:
```bash
rm -rf ~/.local/share/mailcatch
```

## 與 Docker Compose 比較

| 特性 | Docker Compose | Podman + Systemd |
|------|----------------|-------------------|
| 權限需求 | 需要 Docker daemon (通常需要 root) | Rootless |
| 系統整合 | 較少 | 深度整合 systemd |
| 自動重啟 | 透過 restart 策略 | systemd 管理 |
| 日誌管理 | Docker 日誌驅動 | journald 整合 |
| 開機啟動 | 需要額外配置 | systemd 原生支持 |
| 資源控制 | Docker 限制 | systemd + cgroups |

## 參考資源

- [Podman 官方文檔](https://podman.io/getting-started/)
- [systemd 用戶服務](https://www.freedesktop.org/software/systemd/man/systemd.service.html)
- [Podman systemd 整合](https://github.com/containers/podman/blob/main/docs/source/markdown/podman-generate-systemd.1.md)