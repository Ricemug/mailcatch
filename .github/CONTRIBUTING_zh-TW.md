# 為 MailCatch 貢獻

感謝您對 MailCatch 的關注！我們歡迎社群的貢獻。

## 如何貢獻

### 回報問題

在建立新 issue 之前，請先:

1. 檢查問題是否已存在於我們的 [issue tracker](../../issues)
2. 確保您使用的是最新版本
3. 提供關於您的環境和問題的詳細資訊

建立 issue 時，請包含:

- 作業系統和版本
- Go 版本
- 重現問題的步驟
- 預期行為與實際行為
- 相關的日誌輸出或錯誤訊息

### 提交 Pull Request

1. Fork 這個倉庫
2. 從 `main` 建立功能分支: `git checkout -b feature/your-feature-name`
3. 使用清晰、描述性的提交訊息進行變更
4. 為新功能添加測試
5. 確保所有測試通過: `make test`
6. 執行 linter: `make lint`
7. 如有需要，更新文件
8. 提交 pull request 並清楚描述您的變更

### 開發環境設置

1. Clone 您的 fork:
   ```bash
   git clone https://github.com/YOUR-USERNAME/mailcatch.git
   cd mailcatch
   ```

2. 安裝依賴:
   ```bash
   go mod download
   ```

3. 建置專案:
   ```bash
   make build
   ```

4. 執行測試:
   ```bash
   make test
   ```

### 程式碼風格

- 遵循 Go 最佳實踐和慣例
- 使用 `gofmt` 格式化程式碼
- 撰寫清晰、自我說明的程式碼
- 為複雜邏輯添加註解
- 保持函數專注和精簡

### 提交訊息

使用清晰且具描述性的提交訊息:

```
feat: add SMTP authentication support

- Implement PLAIN and LOGIN mechanisms
- Add configuration options for auth
- Update documentation

Closes #123
```

格式: `type: description`

類型:
- `feat`: 新功能
- `fix`: 錯誤修復
- `docs`: 文件變更
- `style`: 程式碼風格變更
- `refactor`: 程式碼重構
- `test`: 添加或更新測試
- `chore`: 維護任務

### 測試

- 為新功能撰寫單元測試
- 確保現有測試仍然通過
- 測試成功和錯誤情況
- 適當時包含整合測試

### 文件

- 如果添加新功能，請更新 README.md
- 為複雜程式碼添加內聯註解
- 更新配置文件
- 在有幫助的地方包含範例

## 專案結構

```
mailcatch/
├── cmd/server/          # 主應用程式入口點
├── internal/
│   ├── config/         # 配置管理
│   ├── models/         # 資料模型
│   ├── smtp/           # SMTP 伺服器實作
│   ├── storage/        # 儲存後端
│   └── web/            # 網頁介面
├── scripts/            # 建置和部署腳本
└── web/static/         # 靜態網頁資源
```

## 取得協助

如果您需要幫助或有問題:

- 查看現有的 [issues](../../issues) 和 [discussions](../../discussions)
- 加入我們的社群聊天 (如果可用)
- 直接聯絡維護者

## 行為準則

請在所有互動中保持尊重和體諒。我們希望為所有貢獻者維護一個友善的環境。

## 授權

透過為 MailCatch 貢獻，您同意您的貢獻將使用與專案相同的授權 (MIT License)。

感謝您的貢獻！🚀
