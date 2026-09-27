# 安全说明

## 报告安全问题

请不要在公开 Issue 中提交密钥、密码、证书内容、真实服务器地址或可复现的敏感数据。请通过仓库维护者提供的私密联系方式报告安全问题。

## 凭据管理

- 仓库只保存 `*.example.*` 示例配置；真实配置、AppID、AppSecret、支付密钥、证书和私钥必须保存在仓库外或已忽略的本地文件中。
- 提交前运行 `pwsh ./scripts/check-secrets.ps1`，并检查 `git diff --cached --name-status`。
- GitHub Actions 会执行仓库规则检查和 Gitleaks 历史扫描。
- 如果仓库属于 GitHub Organization，请在仓库 Actions secrets 中配置有效的 `GITLEAKS_LICENSE`；个人账号仓库不需要该许可证。
- 已经泄漏的凭据必须在对应平台撤销或轮换。删除文件或重写 Git 历史不能使旧凭据自动失效。
