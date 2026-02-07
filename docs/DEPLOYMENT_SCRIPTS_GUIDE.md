# OpenWan 本地部署脚本使用说明

**版本**: 2.0  
**更新日期**: 2026-02-07

---

## 📋 脚本清单

### 核心部署脚本

| 脚本 | 用途 | 使用频率 |
|-----|------|---------|
| `setup-local.sh` | 一键部署开发环境 | 初次安装 |
| `start.sh` | 启动所有服务 | 每天 |
| `stop.sh` | 停止所有服务 | 每天 |
| `restart.sh` | 重启所有服务 | 按需 |
| `status.sh` | 查看服务状态 | 随时 |
| `logs.sh` | 查看服务日志 | 调试时 |

### 数据管理脚本

| 脚本 | 用途 | 使用频率 |
|-----|------|---------|
| `backup.sh` | 备份数据库和文件 | 每周 |
| `restore.sh` | 恢复数据 | 需要时 |
| `db-migrate.sh` | 数据库迁移 | 版本更新时 |

---

## 🚀 快速开始

### 首次部署

```bash
# 1. 克隆代码（如果还没有）
git clone https://github.com/yourorg/openwan.git
cd openwan

# 2. 运行一键部署脚本
./scripts/setup-local.sh

# 等待5-10分钟，部署完成后访问:
# 前端: http://localhost:3000
# 后端: http://localhost:8080
# 用户名: admin
# 密码: admin123
```

### 日常使用

```bash
# 启动服务
./scripts/start.sh

# 查看状态
./scripts/status.sh

# 查看日志
./scripts/logs.sh              # 所有服务
./scripts/logs.sh api          # 仅API服务
./scripts/logs.sh worker       # 仅Worker服务

# 重启服务
./scripts/restart.sh

# 停止服务
./scripts/stop.sh
```

---

## 📖 详细说明

### 1. setup-local.sh - 一键部署脚本

**功能**: 自动化部署完整的OpenWan开发环境

**执行步骤**:
1. 检查系统要求（Docker、Docker Compose、磁盘空间、内存）
2. 停止已有容器
3. 可选：清理旧数据
4. 创建必要的目录结构
5. 生成.env配置文件（含随机密码）
6. 拉取Docker镜像
7. 构建应用镜像
8. 启动所有服务
9. 初始化数据库
10. 创建管理员账号
11. 健康检查
12. 显示访问信息

**使用方法**:
```bash
./scripts/setup-local.sh
```

**交互提示**:
- 是否清理旧数据？(y/N)
- 是否覆盖现有.env文件？(y/N)

**输出示例**:
```
╔═══════════════════════════════════════════════════════════╗
║                 部署成功！                                 ║
╚═══════════════════════════════════════════════════════════╝

访问信息：
  📱 前端地址:    http://localhost:3000
  🔌 后端API:     http://localhost:8080
  📊 健康检查:    http://localhost:8080/health

  🔐 管理员账号:
     用户名: admin
     密码:   admin123
```

---

### 2. start.sh - 启动脚本

**功能**: 启动所有Docker容器

**使用方法**:
```bash
./scripts/start.sh
```

**等效命令**:
```bash
docker-compose up -d
```

**启动的服务**:
- MySQL 8.0 (端口3306)
- Redis 7.0 (端口6379)
- RabbitMQ 3.12 (端口5672, 管理界面15672)
- Sphinx 搜索引擎 (端口9306)
- OpenWan API (端口8080)
- OpenWan Worker (转码服务)
- OpenWan Frontend (端口3000)

**输出**:
```
启动OpenWan服务...
[+] Running 7/7
 ✔ Container openwan-mysql      Started
 ✔ Container openwan-redis      Started
 ✔ Container openwan-rabbitmq   Started
 ✔ Container openwan-sphinx     Started
 ✔ Container openwan-api        Started
 ✔ Container openwan-worker     Started
 ✔ Container openwan-frontend   Started

✓ 服务启动成功！

访问地址:
  前端: http://localhost:3000
  后端: http://localhost:8080
```

---

### 3. stop.sh - 停止脚本

**功能**: 停止所有Docker容器

**使用方法**:
```bash
./scripts/stop.sh
```

**等效命令**:
```bash
docker-compose down
```

**注意**: 这不会删除数据，数据保存在Docker volumes中

---

### 4. restart.sh - 重启脚本

**功能**: 重启所有服务（保留数据）

**使用方法**:
```bash
./scripts/restart.sh
```

**等效命令**:
```bash
docker-compose restart
```

**适用场景**:
- 修改配置文件后
- 服务异常需要重启
- 更新代码后

---

### 5. status.sh - 状态查看脚本

**功能**: 查看所有服务运行状态和健康检查

**使用方法**:
```bash
./scripts/status.sh
```

**输出示例**:
```
OpenWan 服务状态：

NAME                   STATUS    PORTS
openwan-mysql          running   0.0.0.0:3306->3306/tcp
openwan-redis          running   0.0.0.0:6379->6379/tcp
openwan-rabbitmq       running   0.0.0.0:5672->5672/tcp, 15672/tcp
openwan-api            running   0.0.0.0:8080->8080/tcp
openwan-worker         running   
openwan-frontend       running   0.0.0.0:3000->3000/tcp

健康检查：
  API:      ✓ 健康
  Frontend: ✓ 可访问
```

---

### 6. logs.sh - 日志查看脚本

**功能**: 查看服务日志，支持实时跟踪

**使用方法**:
```bash
# 查看所有服务日志
./scripts/logs.sh

# 查看特定服务日志
./scripts/logs.sh api       # API服务
./scripts/logs.sh worker    # Worker服务
./scripts/logs.sh mysql     # MySQL
./scripts/logs.sh redis     # Redis
./scripts/logs.sh frontend  # 前端
```

**快捷键**:
- `Ctrl + C`: 退出日志查看
- `Ctrl + S`: 暂停滚动
- `Ctrl + Q`: 恢复滚动

**日志过滤**:
```bash
# 过滤错误日志
./scripts/logs.sh api | grep ERROR

# 过滤特定关键词
./scripts/logs.sh api | grep "upload"

# 查看最近100行
docker-compose logs --tail=100 api
```

---

### 7. backup.sh - 备份脚本

**功能**: 备份数据库、上传文件和配置

**使用方法**:
```bash
./scripts/backup.sh
```

**备份内容**:
1. MySQL数据库（完整SQL dump）
2. 上传文件目录（tar.gz压缩）
3. .env配置文件
4. config.yaml配置文件

**备份位置**:
```
./backups/openwan_backup_YYYYMMDD_HHMMSS/
├── database.sql              # 数据库备份
├── uploads.tar.gz            # 文件备份
├── env.backup                # 环境变量
├── config.yaml.backup        # 配置文件
└── backup_info.txt           # 备份信息
```

**输出示例**:
```
开始备份OpenWan数据...
备份数据库...
备份上传文件...
备份配置文件...

✓ 备份完成！
备份位置: ./backups/openwan_backup_20260207_153045

备份时间: Fri Feb  7 15:30:45 UTC 2026
数据库大小: 45M
文件大小: 2.3G
总大小: 2.4G
```

**自动化备份**:
```bash
# 添加到crontab，每天凌晨2点备份
crontab -e

# 添加以下行:
0 2 * * * cd /path/to/openwan && ./scripts/backup.sh >> /var/log/openwan-backup.log 2>&1
```

---

### 8. restore.sh - 恢复脚本

**功能**: 从备份恢复数据

**使用方法**:
```bash
# 列出可用备份
ls ./backups/

# 恢复指定备份
./scripts/restore.sh openwan_backup_20260207_153045
```

**交互确认**:
```
警告: 这将覆盖当前数据！
确认恢复？(yes/no): yes
```

**恢复流程**:
1. 恢复数据库
2. 恢复上传文件
3. 提示重启服务

**注意**:
- ⚠️ 恢复会覆盖当前数据，请谨慎操作
- ⚠️ 建议恢复前先做一次备份
- ✅ 恢复后需要重启服务

---

### 9. db-migrate.sh - 数据库迁移脚本

**功能**: 运行数据库迁移（Schema变更）

**使用方法**:
```bash
# 执行迁移（up）
./scripts/db-migrate.sh up

# 回滚迁移（down）
./scripts/db-migrate.sh down

# 查看迁移状态
./scripts/db-migrate.sh status
```

**适用场景**:
- 首次部署（setup-local.sh会自动执行）
- 版本升级（SQL schema变更）
- 添加新表或字段
- 回滚数据库变更

**迁移文件位置**:
```
./migrations/
├── 000001_init_schema.up.sql      # 初始化Schema
├── 000001_init_schema.down.sql    # 回滚脚本
├── 000002_add_audit_logs.up.sql   # 新增审计日志表
└── 000002_add_audit_logs.down.sql # 回滚
```

---

## 🛠️ 故障排查

### 问题1: 端口被占用

**错误**:
```
Error: bind: address already in use
```

**解决**:
```bash
# 查看占用端口的进程
sudo lsof -i :8080
sudo lsof -i :3000
sudo lsof -i :3306

# 停止占用端口的进程
sudo kill -9 <PID>

# 或使用不同端口（修改docker-compose.yaml）
```

### 问题2: Docker权限错误

**错误**:
```
permission denied while trying to connect to the Docker daemon socket
```

**解决**:
```bash
# 将当前用户添加到docker组
sudo usermod -aG docker $USER

# 重新登录或执行
newgrp docker

# 或使用sudo运行脚本
sudo ./scripts/setup-local.sh
```

### 问题3: 容器无法启动

**排查步骤**:
```bash
# 1. 查看容器日志
./scripts/logs.sh <service_name>

# 2. 查看详细错误
docker-compose ps
docker-compose logs <service_name>

# 3. 检查资源使用
docker stats

# 4. 重新构建镜像
docker-compose build --no-cache <service_name>

# 5. 完全清理后重新部署
docker-compose down -v
docker system prune -a
./scripts/setup-local.sh
```

### 问题4: 数据库连接失败

**排查步骤**:
```bash
# 1. 检查MySQL容器状态
docker-compose ps mysql

# 2. 检查MySQL日志
docker-compose logs mysql

# 3. 手动连接测试
docker-compose exec mysql mysql -u root -p

# 4. 检查.env配置
cat .env | grep DB_

# 5. 重启MySQL
docker-compose restart mysql
```

### 问题5: 前端无法访问后端

**排查步骤**:
```bash
# 1. 检查API健康
curl http://localhost:8080/health

# 2. 检查CORS配置
# 查看configs/config.yaml中的cors设置

# 3. 检查前端配置
cat frontend/.env | grep VITE_API_BASE_URL

# 4. 检查网络连接
docker-compose exec frontend ping api
```

---

## 📚 常用命令参考

### Docker Compose命令

```bash
# 查看所有容器
docker-compose ps

# 查看日志
docker-compose logs [service]

# 进入容器
docker-compose exec [service] bash

# 重启单个服务
docker-compose restart [service]

# 停止单个服务
docker-compose stop [service]

# 删除所有容器和数据
docker-compose down -v

# 重新构建镜像
docker-compose build [service]
```

### Docker命令

```bash
# 查看所有容器
docker ps -a

# 查看镜像
docker images

# 清理未使用的资源
docker system prune -a

# 查看资源使用
docker stats

# 查看日志
docker logs [container_id]

# 进入容器
docker exec -it [container_id] bash
```

### 数据库命令

```bash
# 连接MySQL
docker-compose exec mysql mysql -u root -p

# 导出数据库
docker-compose exec mysql mysqldump -u root -p openwan_db > backup.sql

# 导入数据库
docker-compose exec -T mysql mysql -u root -p openwan_db < backup.sql

# 查看数据库列表
docker-compose exec mysql mysql -u root -p -e "SHOW DATABASES;"
```

---

## 🔧 高级配置

### 自定义端口

编辑`docker-compose.yaml`:
```yaml
services:
  api:
    ports:
      - "8081:8080"  # 改为8081
  
  frontend:
    ports:
      - "3001:3000"  # 改为3001
```

### 增加Worker数量

编辑`docker-compose.yaml`:
```yaml
services:
  worker:
    deploy:
      replicas: 4  # 增加到4个worker
```

### 修改数据库配置

编辑`.env`:
```bash
DB_DATABASE=my_openwan
DB_USERNAME=my_user
DB_PASSWORD=my_secure_password
```

---

## 📞 获取帮助

**文档**:
- 用户手册: `docs/USER_MANUAL.md`
- API文档: `docs/API.md`
- 部署指南: `docs/DEPLOYMENT.md`

**社区**:
- GitHub Issues: https://github.com/openwan/openwan/issues
- 邮箱: support@openwan.com

---

**脚本版本**: 2.0  
**最后更新**: 2026-02-07  
**维护者**: OpenWan开发团队
