# UPS Crisis Action Tool

UPS 断电危机处理工具 — 由 UPS PowerManager 软件在断电时调用，SSH 到多台 Linux 服务器并行执行关机脚本。

## 快速开始

### 1. 编辑配置文件

```bash
cp config.json.example config.json
# 修改 servers 列表，填入实际服务器IP、用户名、密码和脚本
```

### 2. 运行

```bash
# Windows (PowerManager 调用)
ups-action.exe config.json

# Linux (直接运行)
./ups-action config.json
```

### 3. 查看日志

日志文件按日期自动生成：`ups-action-YYYY-MM-DD.log`

## 配置文件结构 (config.json)

```json
{
  "global": {
    "ssh_port": 22,                      // 默认 SSH 端口
    "connect_timeout_seconds": 10,       // 连接超时
    "execute_timeout_seconds": 120,      // 脚本执行超时
    "retry_count": 2,                    // 失败重试次数
    "retry_delay_seconds": 5,            // 重试间隔
    "log_dir": ".",                      // 日志目录
    "log_max_days": 30                   // 日志保留天数
  },
  "default_script": [                    // 默认关机脚本（服务器未指定时使用）
    "sync",
    "shutdown -h now"
  ],
  "servers": [                           // 服务器列表
    {
      "host": "192.168.1.10",            // IP 地址
      "port": 22,                        // SSH 端口（可选，默认22）
      "user": "root",                    // SSH 用户
      "password": "xxx",                 // SSH 密码
      "script": [                        // 自定义脚本（可选，不填则用 default_script）
        "systemctl stop nginx",
        "systemctl stop mysql",
        "sync",
        "shutdown -h now"
      ]
    }
  ]
}
```

## 执行逻辑

1. PowerManager 调用 `ups-action.exe config.json`
2. 并行 SSH 连接到所有服务器
3. 通过 `bash -s` 将脚本内容传给远程执行
4. 记录每台服务器的执行结果到日志
5. 全部成功返回退出码 0，任何失败返回 1

## 构建

```bash
# 安装依赖
go mod download

# 构建 Linux 版本
go build -ldflags="-s -w" -o ups-action .

# 交叉编译 Windows 版本
GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ups-action.exe .
```

## 交付物

- `ups-action.exe` — Windows 可执行文件（单文件，零依赖）
- `config.json` — 配置文件
- `ups-action-YYYY-MM-DD.log` — 运行日志

## 测试

```bash
go test -race -cover ./...
```
