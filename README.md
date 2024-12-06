# MyWallet钱包服务

一个基于 Go 语言实现的高性能数字钱包微服务，提供账户余额管理和转账处理功能。

## 功能特点

- 账户余额的增加和扣减
- 交易历史记录
- 并发安全的余额操作
- 基于 Redis 的数据存储
- RESTful API 接口
- 完整的单元测试覆盖

## 技术栈

- Go 1.20+
- Redis
- Gin Web 框架
- 标准库 context 用于超时控制
- 标准库 sync 用于并发控制 
 

### 运行

```bash
go run cmd/main.go
```

## API 接口

### 查询余额

```http
GET /api/v1/wallet/{account_id}/balance

响应:
{
    "account_id": "123456",
    "balance": 1000.00
}
```

### 增加余额

```http
POST /api/v1/wallet/{account_id}/credit
Content-Type: application/json

请求体:
{
    "amount": 100.00
}

响应:
{
    "account_id": "123456",
    "balance": 1100.00,
    "transaction_id": "tx_123abc"
}
```

### 扣减余额

```http
POST /api/v1/wallet/{account_id}/debit
Content-Type: application/json

请求体:
{
    "amount": 50.00
}

响应:
{
    "account_id": "123456",
    "balance": 1050.00,
    "transaction_id": "tx_456def"
}
```

## 错误处理

所有 API 在发生错误时会返回统一格式的错误响应：

```json
{
    "error": {
        "code": "INSUFFICIENT_BALANCE",
        "message": "账户余额不足"
    }
}
```

常见错误代码：
- `ACCOUNT_NOT_FOUND`: 账户不存在
- `INSUFFICIENT_BALANCE`: 余额不足
- `INVALID_AMOUNT`: 金额无效
- `SYSTEM_ERROR`: 系统内部错误

## 测试

运行所有测试：

```bash
go test ./... -v
```

运行性能测试：

```bash
go test ./... -bench=. -benchmem
```

## 项目结构

```
.
├── cmd/
│   └── main.go                 # 应用程序入口
├── internal/
│   ├── api/
│   │   └── handlers.go         # HTTP 处理器
│   ├── config/
│   │   └── config.go           # 配置管理
│   ├── repository/
│   │   └── redis.go           # Redis 数据访问层
│   ├── routes/
│   │   └── routes.go          # 路由配置
│   └── service/
│       ├── wallet.go          # 业务逻辑层
│       └── wallet_test.go     # 业务逻辑测试
└── README.md
```

## 性能优化

- Redis 连接池配置优化
  - 最大空闲连接数：50
  - 最大活跃连接数：100
  - 空闲超时：5分钟

- 并发控制
  - 使用互斥锁保证账户操作的原子性
  - 采用分布式锁处理跨实例并发

- 超时控制
  - API 请求默认超时时间：5秒
  - Redis 操作超时时间：2秒

## 监控指标

服务暴露以下监控指标：

- HTTP 接口延迟
- Redis 操作延迟
- 交易成功/失败率
- 并发请求数
- 系统资源使用率

## 部署

### Docker 部署

```bash
# 构建镜像
docker build -t wallet-service .

# 运行容器
docker run -d \
  -p 8080:8080 \
  -e REDIS_ADDR=redis:6379 \
  --name wallet-service \
  wallet-service
```

 

 
