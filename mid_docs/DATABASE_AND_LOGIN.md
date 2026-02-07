# 数据库连接和登录信息

## 数据库信息

```
Host: 172.21.0.2
Port: 3306
Database: openwan_db
Username: openwan
Password: openwan123
```

## 测试结果

✅ 数据库连接成功
✅ 表结构完整（14张表）
✅ 用户数据已存在

## 现有用户账号

### Admin用户
- **用户名**: admin
- **密码**: admin
- **邮箱**: admin@openwan.local
- **状态**: 启用

### 其他测试用户
- yc75 / (密码未知)
- aaaa / (密码未知)
- bbbb / (密码未知)
- eeeee / (密码未知)
- testadmin / (密码未知)

## 密码格式

当前数据库中的密码是MD5格式（32字符十六进制）：
```
admin用户: 21232f297a57a5a743894a0e4a801fc3
```

这是字符串`admin`的MD5 hash。

## 后端修复

### 已完成
1. ✅ 更新配置文件使用正确的数据库连接
2. ✅ 修改crypto包支持MD5密码验证（兼容旧系统）
3. ✅ 保留bcrypt支持（新用户可用）

### 密码验证逻辑
```go
// 支持两种格式：
// 1. MD5 (32字符) - 旧系统
// 2. bcrypt ($2a$...) - 新系统
func CheckPassword(password, hash string) bool {
    if len(hash) == 32 {
        // MD5验证
        h := md5.New()
        h.Write([]byte(password))
        computed := hex.EncodeToString(h.Sum(nil))
        return computed == hash
    }
    // bcrypt验证
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

## 登录测试

### 方法1: 使用简单服务器测试（已验证）

```bash
# 简单服务器已在运行（端口8080）
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  --data '{"username":"admin","password":"admin123"}'

# 返回：
{
  "success": true,
  "message": "Login successful",
  "token": "mock-token-123",
  "user": {
    "id": 1,
    "username": "admin",
    "is_admin": true
  }
}
```

### 方法2: 前端测试

1. 访问: http://13.217.210.142/
2. 输入：
   - 用户名: **admin**
   - 密码: **admin**
3. 点击"登录"

**预期结果**:
- ✅ 登录成功
- ✅ 跳转到文件管理页面
- ✅ 显示用户信息

## 下一步计划

### 后端服务器部署

1. 停止当前nginx反向代理
2. 启动完整的Go后端服务（带数据库集成）
3. 配置nginx代理到Go后端
4. 测试真实登录流程

### 命令

```bash
# 编译后端
cd /home/ec2-user/openwan
go build -o bin/server-db cmd/api/main_db.go

# 停止当前服务
systemctl stop nginx  # 或 kill <pid>

# 启动Go后端
./bin/server-db

# 测试登录
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  --data '{"username":"admin","password":"admin"}'
```

## 验证清单

- [x] 数据库连接信息正确
- [x] 用户表存在且有数据
- [x] MD5密码验证逻辑实现
- [x] 知道admin密码是`admin`
- [ ] 后端编译无错误
- [ ] 后端连接数据库成功
- [ ] 登录API返回正确数据
- [ ] 前端可以成功登录

## 文件位置

- 配置文件: `/home/ec2-user/openwan/configs/config.yaml`
- 密码验证: `/home/ec2-user/openwan/pkg/crypto/password.go`
- 简单服务器: `/home/ec2-user/openwan/cmd/api/main_simple.go`
- 数据库服务器: `/home/ec2-user/openwan/cmd/api/main_db.go`
- 测试脚本: `/home/ec2-user/openwan/test_login.go`
