# Pump.fun 新币监控 - gRPC Golang Demo

## 功能

- 实时监控新币创建
- 显示代币铸造地址和相关信息
- 连接健康检查 (ping/pong)

## 使用方法

1. 克隆项目

```bash
git clone https://github.com/ChainBuff/grpc-demo-golang.git
```

2. 在项目根目录创建 `.env` 文件

```bash
GRPC_URL=你的_grpc_节点地址
# 可以使用免费节点，比如 solana-yellowstone-grpc.publicnode.com:443
# 注意: 不需要再加 https:// 前缀
```

3. 运行

```bash
go mod tidy # 安装依赖
go run cmd/monitor.go
```

## 输出示例

```
2025/01/09 15:54:35.802204 ----- New Token Created -----
Signature: 4zcaz5NKKKSTUnKp4u4iKhAPNNAQrGhETXp5wW27qLexjpjpFgpThvp9vF9a9QXAdFUZ2GTx38a4yk6BFKmkapoj
Slot: 312906820
Mint: 9S6hpnUdFFRgRVwuwvS4kw4Un16CTMRzbE6AeuMWpump
Bonding Curve: EEjyPV1514jCRRVCn98CcjBWWgcg7SDpPBwJmpnG1qZh
Associated Bonding Curve: BhsTp2BWJKhT9yZZbbz9gQgE1dMoiAv4xrzQjtrtwQLu
```

## 环境要求

- Go 1.21 或更高版本
- Solana gRPC 节点

## 贡献

如果发现任何 **bug** 或有功能改进建议，欢迎提交 Issues 或 Pull Requests 至本仓库。
