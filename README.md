project-root/
├── cmd/                    # 启动入口
│   └── server/             # 主服务
│       └── main.go
├── configs/                # 配置文件
│   └── config.yaml
├── internal/               # 内部逻辑（对外隐藏）
│   ├── api/                # API 层
│   │   ├── handlers/       # 业务处理函数
│   │   │   ├── health.go
│   │   │   ├── auth.go
│   │   │   └── user.go
│   │   └── router.go       # gin 路由注册
│   ├── middleware/         # 中间件（插件化）
│   │   ├── jwt.go
│   │   ├── limiter.go
│   │   ├── recovery.go
│   │   └── cors.go
│   ├── plugin/             # 插件（Redis/MySQL/…）
│   │   ├── redis.go
│   │   ├── mysql.go
│   │   └── plugin.go       # 插件加载器
│   ├── service/            # 业务逻辑
│   │   └── user.go
│   ├── metrics/            # 监控相关
│   │   ├── prometheus.go   # 指标采集
│   │   └── middleware.go   # 请求指标中间件
│   └── app.go              # 应用初始化
├── pkg/                    # 公共库（通用工具）
│   ├── config/             # viper 配置封装
│   │   └── config.go
│   ├── logger/             # zap 日志封装
│   │   └── logger.go
│   └── utils/              # 通用工具
├── go.mod
└── go.sum
