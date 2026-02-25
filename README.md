# 鱼跃授权系统

程序授权管理系统，支持授权码生成、在线注入授权、支付对接等功能。

## 技术栈

| 技术 | 版本 |
|------|------|
| Go | 1.26 |
| Gin | v1.10.0 |
| GORM | v1.25.10 |
| go-redis | v9.5.1 |
| Nuxt | 4.3.1 |
| Nuxt UI | v4.5.0 |
| Vue | 3.5.x |

## 快速开始

### 环境要求

- Go 1.26+
- Node.js 20+
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

## 项目结构

```
.
├── main.go                          # 后端入口
├── config/config.go                 # 配置管理
├── models/models.go                 # 数据模型
├── controllers/                     # 控制器
│   ├── install.go                   # 安装向导
│   ├── auth.go                      # 认证
│   ├── program.go                   # 程序管理
│   ├── license.go                   # 授权码管理
│   ├── injection.go                 # 在线注入
│   ├── finance.go                   # 财务/支付
│   └── setting.go                   # 系统设置
├── middleware/auth.go               # JWT认证中间件
├── routes/routes.go                 # 路由配置
├── utils/response.go                # 工具函数
└── frontend/                        # Nuxt 4 前端
    ├── nuxt.config.ts
    ├── app/
    │   ├── app.vue
    │   ├── app.config.ts
    │   ├── pages/                   # 页面
    │   ├── layouts/                 # 布局
    │   ├── composables/             # 组合式函数
    │   ├── stores/                  # Pinia状态管理
    │   └── middleware/              # 路由中间件
    └── package.json
```

## API接口

- `POST /api/license/verify` - 授权码验证（外部调用）
- `POST /api/auth/login` - 管理员登录
- `GET /api/programs` - 程序列表
- `GET /api/licenses` - 授权码列表
- 更多接口请参考 `routes/routes.go`

## 端口

默认运行在 **3132** 端口。
