#!/bin/bash

# MailCatch Podman Rootless Systemd Setup Script
# 這個腳本會建立 Podman rootless 容器的 systemd 服務

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 默認配置
DEFAULT_SMTP_PORT=2525
DEFAULT_HTTP_PORT=8080
DEFAULT_DATA_DIR="$HOME/.local/share/mailcatch"
DEFAULT_LOG_DIR="$HOME/.local/share/mailcatch/logs"
DEFAULT_CONTAINER_NAME="mailcatch"
DEFAULT_IMAGE="mailcatch:latest"

# 函數：顯示幫助信息
show_help() {
    echo "MailCatch Podman Rootless Systemd 設置腳本"
    echo ""
    echo "用法: $0 [選項]"
    echo ""
    echo "選項:"
    echo "  -s, --smtp-port PORT     SMTP 端口 (默認: $DEFAULT_SMTP_PORT)"
    echo "  -w, --web-port PORT      Web UI 端口 (默認: $DEFAULT_HTTP_PORT)"
    echo "  -d, --data-dir PATH      數據目錄 (默認: $DEFAULT_DATA_DIR)"
    echo "  -l, --log-dir PATH       日誌目錄 (默認: $DEFAULT_LOG_DIR)"
    echo "  -n, --name NAME          容器名稱 (默認: $DEFAULT_CONTAINER_NAME)"
    echo "  -i, --image IMAGE        容器映像 (默認: $DEFAULT_IMAGE)"
    echo "  -h, --help               顯示此幫助信息"
    echo "  --uninstall              移除服務和容器"
    echo ""
    echo "示例:"
    echo "  $0                       # 使用默認設置"
    echo "  $0 -s 1025 -w 3000       # 自定義端口"
    echo "  $0 --uninstall           # 移除服務"
}

# 解析命令行參數
SMTP_PORT=$DEFAULT_SMTP_PORT
HTTP_PORT=$DEFAULT_HTTP_PORT
DATA_DIR=$DEFAULT_DATA_DIR
LOG_DIR=$DEFAULT_LOG_DIR
CONTAINER_NAME=$DEFAULT_CONTAINER_NAME
IMAGE=$DEFAULT_IMAGE
UNINSTALL=false

while [[ $# -gt 0 ]]; do
    case $1 in
        -s|--smtp-port)
            SMTP_PORT="$2"
            shift 2
            ;;
        -w|--web-port)
            HTTP_PORT="$2"
            shift 2
            ;;
        -d|--data-dir)
            DATA_DIR="$2"
            shift 2
            ;;
        -l|--log-dir)
            LOG_DIR="$2"
            shift 2
            ;;
        -n|--name)
            CONTAINER_NAME="$2"
            shift 2
            ;;
        -i|--image)
            IMAGE="$2"
            shift 2
            ;;
        --uninstall)
            UNINSTALL=true
            shift
            ;;
        -h|--help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}未知選項: $1${NC}"
            show_help
            exit 1
            ;;
    esac
done

# 函數：檢查依賴
check_dependencies() {
    echo -e "${BLUE}檢查依賴...${NC}"

    if ! command -v podman &> /dev/null; then
        echo -e "${RED}錯誤: podman 未安裝${NC}"
        echo "請安裝 podman: sudo dnf install podman (Fedora/RHEL) 或 sudo apt install podman (Ubuntu/Debian)"
        exit 1
    fi

    if ! systemctl --user show-environment &> /dev/null; then
        echo -e "${RED}錯誤: 用戶 systemd 服務不可用${NC}"
        echo "請確保 systemd 用戶服務已啟用"
        exit 1
    fi

    echo -e "${GREEN}依賴檢查通過${NC}"
}

# 函數：移除服務
uninstall_service() {
    echo -e "${YELLOW}移除 MailCatch 服務...${NC}"

    # 停止並禁用服務
    if systemctl --user is-active --quiet "$CONTAINER_NAME.service" 2>/dev/null; then
        echo "停止服務..."
        systemctl --user stop "$CONTAINER_NAME.service"
    fi

    if systemctl --user is-enabled --quiet "$CONTAINER_NAME.service" 2>/dev/null; then
        echo "禁用服務..."
        systemctl --user disable "$CONTAINER_NAME.service"
    fi

    # 移除服務文件
    SERVICE_FILE="$HOME/.config/systemd/user/$CONTAINER_NAME.service"
    if [[ -f "$SERVICE_FILE" ]]; then
        echo "移除服務文件..."
        rm "$SERVICE_FILE"
        systemctl --user daemon-reload
    fi

    # 停止並移除容器
    if podman container exists "$CONTAINER_NAME" 2>/dev/null; then
        echo "停止並移除容器..."
        podman stop "$CONTAINER_NAME" 2>/dev/null || true
        podman rm "$CONTAINER_NAME" 2>/dev/null || true
    fi

    # 移除 volumes（可選）
    read -p "是否同時移除 Podman volumes？這將刪除所有儲存的郵件 (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        DATA_VOLUME_NAME="${CONTAINER_NAME}-data"
        LOG_VOLUME_NAME="${CONTAINER_NAME}-logs"

        if podman volume exists "$DATA_VOLUME_NAME" 2>/dev/null; then
            echo "移除數據 volume: $DATA_VOLUME_NAME"
            podman volume rm "$DATA_VOLUME_NAME" 2>/dev/null || true
        fi

        if podman volume exists "$LOG_VOLUME_NAME" 2>/dev/null; then
            echo "移除日誌 volume: $LOG_VOLUME_NAME"
            podman volume rm "$LOG_VOLUME_NAME" 2>/dev/null || true
        fi
    fi

    echo -e "${GREEN}服務已成功移除${NC}"
}

# 函數：創建目錄
create_directories() {
    echo -e "${BLUE}創建必要目錄...${NC}"

    mkdir -p "$DATA_DIR"
    mkdir -p "$LOG_DIR"
    mkdir -p "$HOME/.config/systemd/user"

    echo "數據目錄: $DATA_DIR"
    echo "日誌目錄: $LOG_DIR"
}

# 函數：拉取或建構映像
prepare_image() {
    echo -e "${BLUE}準備容器映像...${NC}"

    if [[ "$IMAGE" == "mailcatch:latest" ]]; then
        # 檢查是否存在本地映像
        if ! podman image exists "$IMAGE" 2>/dev/null; then
            echo "本地映像不存在，嘗試建構..."
            if [[ -f "Dockerfile" ]]; then
                echo "建構 MailCatch 映像..."
                podman build -t "$IMAGE" .
            else
                echo -e "${YELLOW}警告: 找不到 Dockerfile，請確保映像 '$IMAGE' 存在${NC}"
                echo "你可以:"
                echo "1. 在專案根目錄運行此腳本"
                echo "2. 先使用 'make docker-build' 建構映像"
                echo "3. 或指定其他映像名稱"
                read -p "是否繼續? (y/N): " -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                    exit 1
                fi
            fi
        fi
    else
        echo "檢查映像: $IMAGE"
        if ! podman image exists "$IMAGE" 2>/dev/null; then
            echo "拉取映像..."
            podman pull "$IMAGE" || {
                echo -e "${RED}無法拉取映像: $IMAGE${NC}"
                exit 1
            }
        fi
    fi

    echo -e "${GREEN}映像準備完成${NC}"
}

# 函數：創建 Podman volumes
create_podman_volumes() {
    echo -e "${BLUE}創建 Podman volumes...${NC}"

    DATA_VOLUME_NAME="${CONTAINER_NAME}-data"
    LOG_VOLUME_NAME="${CONTAINER_NAME}-logs"

    # 創建 data volume
    if ! podman volume exists "$DATA_VOLUME_NAME" 2>/dev/null; then
        echo "創建數據 volume: $DATA_VOLUME_NAME"
        podman volume create "$DATA_VOLUME_NAME"
    fi

    # 創建 logs volume
    if ! podman volume exists "$LOG_VOLUME_NAME" 2>/dev/null; then
        echo "創建日誌 volume: $LOG_VOLUME_NAME"
        podman volume create "$LOG_VOLUME_NAME"
    fi

    echo "Volumes 創建完成"
}

# 函數：創建 systemd 服務文件
create_systemd_service() {
    echo -e "${BLUE}創建 systemd 服務文件...${NC}"

    SERVICE_FILE="$HOME/.config/systemd/user/$CONTAINER_NAME.service"

    DATA_VOLUME_NAME="${CONTAINER_NAME}-data"
    LOG_VOLUME_NAME="${CONTAINER_NAME}-logs"

    cat > "$SERVICE_FILE" << EOF
[Unit]
Description=MailCatch Container
Wants=network-online.target
After=network-online.target
RequiresMountsFor=%t/containers

[Service]
Environment=PODMAN_SYSTEMD_UNIT=%n
Restart=on-failure
RestartSec=5
TimeoutStopSec=70
ExecStartPre=/bin/rm -f %t/%n.ctr-id
ExecStart=/usr/bin/podman run \\
    --cidfile=%t/%n.ctr-id \\
    --cgroups=no-conmon \\
    --rm \\
    --sdnotify=conmon \\
    --replace \\
    --name $CONTAINER_NAME \\
    --publish $SMTP_PORT:2525 \\
    --publish $HTTP_PORT:8080 \\
    --volume $DATA_VOLUME_NAME:/app/data \\
    --volume $LOG_VOLUME_NAME:/app/logs \\
    --env SMTP_PORT=2525 \\
    --env HTTP_PORT=8080 \\
    --env DB_PATH=/app/data/emails.db \\
    --env LOG_PATH=/app/logs/mailcatch.log \\
    --env CLEAR_ON_SHUTDOWN=true \\
    $IMAGE
ExecStop=/usr/bin/podman stop --ignore --cidfile=%t/%n.ctr-id
ExecStopPost=/usr/bin/podman rm -f --ignore --cidfile=%t/%n.ctr-id
Type=notify
NotifyAccess=all

[Install]
WantedBy=default.target
EOF

    echo "服務文件已創建: $SERVICE_FILE"
}

# 函數：啟用並啟動服務
enable_service() {
    echo -e "${BLUE}啟用並啟動服務...${NC}"

    # 重新載入 systemd
    systemctl --user daemon-reload

    # 啟用開機自啟動
    systemctl --user enable "$CONTAINER_NAME.service"

    # 啟動服務
    systemctl --user start "$CONTAINER_NAME.service"

    # 檢查狀態
    sleep 3
    if systemctl --user is-active --quiet "$CONTAINER_NAME.service"; then
        echo -e "${GREEN}服務已成功啟動${NC}"
    else
        echo -e "${RED}服務啟動失敗${NC}"
        echo "檢查日誌: journalctl --user -u $CONTAINER_NAME.service"
        exit 1
    fi
}

# 函數：顯示安裝完成信息
show_completion_info() {
    echo -e "${GREEN}"
    echo "=================================================="
    echo "MailCatch 已成功設置為 systemd 服務!"
    echo "=================================================="
    echo -e "${NC}"

    echo "服務信息:"
    echo "  服務名稱: $CONTAINER_NAME.service"
    echo "  SMTP 端口: $SMTP_PORT"
    echo "  Web UI 端口: $HTTP_PORT"
    echo "  Web UI 地址: http://localhost:$HTTP_PORT"
    echo ""

    echo "數據位置:"
    echo "  數據 volume: ${CONTAINER_NAME}-data"
    echo "  日誌 volume: ${CONTAINER_NAME}-logs"
    echo "  本地備份目錄: $DATA_DIR (未使用)"
    echo "  本地日誌目錄: $LOG_DIR (未使用)"
    echo ""

    echo "Volume 管理:"
    echo "  查看 volumes: podman volume ls"
    echo "  查看數據位置: podman volume inspect ${CONTAINER_NAME}-data"
    echo "  備份數據: podman run --rm -v ${CONTAINER_NAME}-data:/data -v \$(pwd):/backup alpine tar czf /backup/mailcatch-data.tar.gz /data"
    echo ""

    echo "常用命令:"
    echo "  查看狀態: systemctl --user status $CONTAINER_NAME.service"
    echo "  查看日誌: journalctl --user -u $CONTAINER_NAME.service -f"
    echo "  重啟服務: systemctl --user restart $CONTAINER_NAME.service"
    echo "  停止服務: systemctl --user stop $CONTAINER_NAME.service"
    echo "  啟動服務: systemctl --user start $CONTAINER_NAME.service"
    echo ""

    echo "移除服務:"
    echo "  $0 --uninstall"
    echo ""

    echo -e "${BLUE}注意: 要使用戶服務在開機時自動啟動，請運行:${NC}"
    echo "  sudo loginctl enable-linger $USER"
}

# 主函數
main() {
    echo -e "${BLUE}MailCatch Podman Rootless Systemd 設置${NC}"
    echo "======================================"

    if [[ "$UNINSTALL" == "true" ]]; then
        check_dependencies
        uninstall_service
        exit 0
    fi

    echo "配置:"
    echo "  SMTP 端口: $SMTP_PORT"
    echo "  Web UI 端口: $HTTP_PORT"
    echo "  數據目錄: $DATA_DIR"
    echo "  日誌目錄: $LOG_DIR"
    echo "  容器名稱: $CONTAINER_NAME"
    echo "  映像: $IMAGE"
    echo ""

    read -p "是否繼續安裝? (Y/n): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Nn]$ ]]; then
        echo "安裝已取消"
        exit 0
    fi

    check_dependencies
    create_directories
    prepare_image

    # 如果服務已存在，先停止
    if systemctl --user is-active --quiet "$CONTAINER_NAME.service" 2>/dev/null; then
        echo -e "${YELLOW}停止現有服務...${NC}"
        systemctl --user stop "$CONTAINER_NAME.service"
    fi

    # 如果容器已存在，先移除
    if podman container exists "$CONTAINER_NAME" 2>/dev/null; then
        echo -e "${YELLOW}移除現有容器...${NC}"
        podman stop "$CONTAINER_NAME" 2>/dev/null || true
        podman rm "$CONTAINER_NAME" 2>/dev/null || true
    fi

    create_podman_volumes
    create_systemd_service
    enable_service
    show_completion_info
}

# 運行主函數
main "$@"