# 前端部署快速指南

## 📦 部署流程（3步完成）

### 1️⃣ 编译前端
```bash
cd /home/ec2-user/openwan/frontend
npm run build
```
**输出**: `dist/` 目录

---

### 2️⃣ 配置已就绪
```bash
# Nginx配置已指向 dist/ 目录
# 无需手动操作
```

---

### 3️⃣ 访问应用
```
http://13.217.210.142/
```

---

## 🔄 更新部署（当前端代码修改后）

```bash
# 1. 重新编译
cd /home/ec2-user/openwan/frontend
npm run build

# 2. 完成！（Nginx自动服务新文件）
```

**说明**: 
- 静态资源文件名包含哈希（自动版本控制）
- index.html不缓存（更新立即生效）
- 无需重启Nginx

---

## ✅ 当前状态

| 组件 | 状态 | 说明 |
|------|------|------|
| **前端静态文件** | ✅ 已部署 | `/home/ec2-user/openwan/frontend/dist/` |
| **Nginx配置** | ✅ 已配置 | 服务静态文件 + API代理 |
| **访问地址** | ✅ 可用 | http://13.217.210.142/ |
| **Vite服务器** | ✅ 已停止 | 不再需要 |

---

## 🎯 架构

```
浏览器
  ↓
Nginx :80
  ├─ / → 静态文件 (dist/)
  └─ /api → 代理到 :8080 (API)
```

---

## 📝 开发模式（可选）

如需使用HMR等开发功能：

```bash
cd frontend
npm run dev
# 访问 http://localhost:3000
```

**注意**: 生产环境不使用开发模式

---

## 🔍 验证

```bash
# 测试前端
curl http://localhost/ | grep "<title>"

# 测试API
curl http://localhost/api/v1/ping

# 测试健康检查
curl http://localhost/health
```

---

**部署完成**: ✅  
**状态**: 🟢 可用
