# DeskOrder - 桌面扫码点餐系统

## 项目简介

DeskOrder 是一个完整的桌面扫码点餐系统，包含用户端小程序、商家端小程序、商户端 H5、POS 终端、后台管理系统和后端服务。

## 演示视频

[▶ 查看或下载 DeskOrder 功能演示视频](https://github.com/Xiaeer/desk-order/releases/download/v0.1.0/desk-order-demo.mp4)

视频以 MP4 格式托管在 GitHub Release 中，大小约为 20 MB。

## 项目结构

```
DeskOrder/
├── backend/          # 后端服务（Golang + Gin + GORM）
├── admin-web/        # 后台管理系统（Vue3 + Element Plus）
├── merchant-h5/      # 商户端 H5（Vue3 + Vue Router + Axios + Vite）
├── mini-user/        # 用户端微信小程序
├── mini-merchant/    # 商家端微信小程序
├── pos-client/       # POS 终端客户端（Electron）
└── docs/             # 项目文档
```

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Golang, Gin, GORM, MySQL, Redis, WebSocket, Viper, Zerolog |
| 用户端 | 微信小程序原生 |
| 商家端 | 微信小程序原生 |
| 商户端 H5 | Vue3, Vue Router, Axios, Vite |
| 后台管理 | Vue3, Element Plus, Vite, Pinia |
| POS 终端 | Electron, Vue3 |

## 本地配置

仓库只提供不含凭据的示例文件。首次运行前请复制示例并填写自己的本地配置：

```powershell
Copy-Item .\backend\configs\config.example.yaml .\backend\configs\config.yaml
Copy-Item .\admin-web\.env.production.example .\admin-web\.env.production
Copy-Item .\merchant-h5\.env.production.example .\merchant-h5\.env.production
Copy-Item .\mini-user\utils\config.local.example.js .\mini-user\utils\config.local.js
Copy-Item .\mini-merchant\utils\config.local.example.js .\mini-merchant\utils\config.local.js
```

微信小程序的真实 AppID 请放在本机的 `project.private.config.json` 中，可参考各小程序目录下的 `project.private.config.example.json`。数据库密码、JWT Secret、微信 AppSecret、支付 API Key、私钥和证书不得提交到 Git。

生产部署建议把真实配置和运行证书保存在仓库外，并通过 `deploy/config.example.psd1` 中的 `Backend.ConfigDir` 与 `Backend.RuntimeAssetsDir` 引用。详细说明见 `deploy/README.md`。

## 安全提醒

- 客户端代码中的地址和 AppID 对最终用户可见，不能把 AppSecret、支付密钥或私钥放入小程序、H5 或桌面客户端。
- 提交前运行 `pwsh ./scripts/check-secrets.ps1`，并检查暂存区，确认没有 `.env`、`config.yaml`、`project.private.config.json`、证书或私钥。
- GitHub Actions 会对每次推送和 Pull Request 执行仓库规则检查与 Gitleaks 扫描；组织账号仓库需额外配置 `GITLEAKS_LICENSE` Actions secret。
- 如果凭据曾进入公开仓库，应立即撤销或轮换；仅删除文件不能让旧提交中的凭据失效。

安全问题的报告方式与凭据管理约定见 [SECURITY.md](SECURITY.md)。

## 许可证

本项目采用 [MIT License](LICENSE) 开源。
