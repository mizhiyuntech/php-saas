# 鱼跃授权系统

程序授权管理系统，支持授权码生成、在线注入授权、支付对接等功能。

## 技术栈

**后端**: Go + Gin + GORM + MySQL + Redis

**前端**: Vue 3 + Nuxt 3 + Nuxt UI v3

## 快速开始

### 环境要求

- Go 1.22+
- Node.js 18+
- MySQL 5.7+
- Redis 6+

### 构建前端

```bash
cd frontend
npm install
npm run generate
```

### 构建后端

```bash
go build -o yuyue-auth .
```

### 运行

```bash
./yuyue-auth
```

访问 `http://your-ip:3132/install` 进行系统安装配置。

## 功能列表

- 系统安装向导（MySQL、Redis、管理员配置）
- 程序管理（添加/编辑/删除程序）
- 授权码管理（批量生成、状态管理、验证）
- 在线注入授权（上传ZIP，自动注入授权验证代码）
- 财务管理（微信支付、支付宝、易支付配置）
- 订单查询
- 系统设置（网站标题、SEO、备案信息、图标）
- 标签页导航 + 面包屑

## API接口

- `POST /api/license/verify` - 授权码验证（外部调用）
- `POST /api/auth/login` - 管理员登录
- `GET /api/programs` - 程序列表
- `GET /api/licenses` - 授权码列表
- 更多接口请参考 `routes/routes.go`

## 端口

默认运行在 **3132** 端口。
