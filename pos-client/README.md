# DeskOrder POS Client

## 目标

- 运行在 Windows 7 收银机上
- 监听后端 POS WebSocket 推单
- 顾客在小程序支付后自动收到订单
- 使用系统默认打印机或指定打印机打印顾客单和后厨单

## 当前实现

- Electron 22，兼容 Windows 7
- 本地配置 `API 地址`、`shop_id`、`POS 密钥`、`终端名称`、`顾客单打印机`、`后厨单打印机`
- 新增 POS 收银台：菜单选品、购物车下单、生成微信支付二维码、轮询支付状态
- 连接 `POST /api/v1/pos/login`
- 监听 `GET /api/v1/pos/ws?shop_id=xxx`
- 收到 `new_order` 后自动提示，并按配置打印顾客单和后厨单
- 支持单打印机连打两张，也支持顾客单/后厨单分配不同打印机
- 支持手动重打最近 50 条订单，可分别打印顾客单、后厨单或全部重打

## 本地启动

```powershell
cd pos-client
$env:ELECTRON_MIRROR='https://npmmirror.com/mirrors/electron/'
npm.cmd install --registry=https://registry.npmmirror.com
npm.cmd start
```

## 打包

64 位 Windows 7:

```powershell
cd pos-client
npm.cmd run dist:x64
```

32 位 Windows 7:

```powershell
cd pos-client
npm.cmd run dist:ia32
```

如果 `dist:ia32` 的 portable 包装步骤在当前开发机失败，可以先生成 32 位可运行目录：

```powershell
cd pos-client
npm.cmd run pack:ia32
```

成功后把整个 `dist/win-ia32-unpacked/` 目录完整复制到 POS 机，再运行其中的 `DeskOrder POS.exe`。这种方式不依赖 portable 外壳，对 Win7 32 位更直接。

注意区分：

- `dist:x64` 产物只能给 64 位 Windows 用。
- `dist:ia32` 或 `pack:ia32` 产物才可以给 32 位 Windows 7 用。
- 当前 `npm.cmd run pack` 默认更偏向当前开发机架构，不适合作为给 32 位 POS 机发包的固定命令。

如果网络无法访问 GitHub，先设置 Electron 镜像：

```powershell
$env:ELECTRON_MIRROR='https://npmmirror.com/mirrors/electron/'
```

## 首次使用

1. 填写后端 API 地址，例如 `https://api.example.com`
2. 填写门店对应的 `shop_id`
3. 填写商户 H5 店铺页展示的 `POS 密钥`
4. 选择顾客单打印机；留空则使用系统默认打印机
5. 选择后厨单打印机；留空则跟随顾客单打印机，如果顾客单也留空则继续使用系统默认打印机
6. 根据门店需要开启或关闭“自动打印顾客单 / 自动打印后厨单”
7. 保存并连接
8. 点击“测试双联打印”先确认顾客单和后厨单链路正常

## 后端依赖

- `POST /api/v1/pos/login`
- `POST /api/v1/pos/menu`
- `POST /api/v1/pos/order`
- `POST /api/v1/pos/order/:id/pay`
- `POST /api/v1/pos/order/:id/cancel`
- `GET /api/v1/pos/order/:id?shop_id=xxx&pos_token=xxx`
- `GET /api/v1/pos/ws?shop_id=xxx`
- 支付回调后会推送完整订单数据到 POS，包括订单号、金额、备注、桌号快照、菜品明细和菜品规格摘要
- POS 登录和 WebSocket 推单连接都要求 `shop_id + POS 密钥`

## 说明

- 当前版本是首版闭环，重点是“收单 + 打印”
- 收银台支持在 POS 端选择商品规格/加料选项并下单
- POS 现在不能再仅凭 `shop_id` 直接连入，必须使用商户 H5 店铺页里的 POS 密钥
- 打印方式为 Electron 静默打印到本机打印机
- 顾客单默认打印金额和合计；后厨单默认突出桌号、菜品、规格摘要和备注，不打印金额
- 如果厨房打印机未配置，后厨单会回退到顾客单打印机继续打印