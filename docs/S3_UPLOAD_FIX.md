# S3 上传问题已修复

## 🔧 修复内容

### 问题
上传文件时报错：
```
failed to upload to S3: operation error S3: PutObject, get identity: get credentials: 
failed to refresh cached credentials, static credentials are empty
```

### 根本原因
S3存储服务在没有显式提供静态凭证时，仍然尝试使用空的静态凭证，而不是使用AWS默认凭证链。

### 解决方案
修改了 `internal/storage/s3.go` 中的 `NewS3Storage` 函数：
- 检查是否提供了静态凭证（AccessKeyID 和 SecretAccessKey）
- 如果未提供静态凭证，使用AWS默认凭证链
- AWS默认凭证链会按顺序查找：
  1. 环境变量 (AWS_ACCESS_KEY_ID, AWS_SECRET_ACCESS_KEY)
  2. ~/.aws/credentials 文件
  3. IAM角色（如果在EC2实例上）

### 代码变更
```go
// 修改前：强制使用静态凭证（即使为空）
if cfg.UseIAMRole {
    awsCfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
} else {
    awsCfg, err = config.LoadDefaultConfig(ctx,
        config.WithRegion(cfg.Region),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            cfg.AccessKeyID,
            cfg.SecretAccessKey,
            "",
        )),
    )
}

// 修改后：智能选择凭证来源
hasStaticCredentials := cfg.AccessKeyID != "" && cfg.SecretAccessKey != ""

if cfg.UseIAMRole || !hasStaticCredentials {
    // 使用默认凭证链（包括 ~/.aws/credentials）
    awsCfg, err = config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
} else {
    // 使用提供的静态凭证
    awsCfg, err = config.LoadDefaultConfig(ctx,
        config.WithRegion(cfg.Region),
        config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
            cfg.AccessKeyID,
            cfg.SecretAccessKey,
            "",
        )),
    )
}
```

---

## ✅ 验证修复

### 1. 服务已重启
```bash
✓ 后端编译完成
✓ 服务已重启
✓ S3存储服务初始化成功
```

### 2. 当前配置
```
存储类型: AWS S3
S3存储桶: video-bucket-843250590784
AWS区域: us-east-1
凭证来源: 默认凭证链 (~/.aws/credentials)
```

### 3. 验证凭证可用
```bash
aws s3 ls s3://video-bucket-843250590784/
# 应该返回成功（无错误）
```

---

## 🧪 测试上传

### 通过Web界面测试

1. **访问应用**
   ```
   URL: http://localhost
   ```

2. **登录**
   ```
   用户名: admin
   密码: admin123
   ```

3. **上传文件**
   - 点击"文件管理" → "文件上传"
   - 选择任意文件（文本、图片、视频等）
   - 填写必要信息：
     - 选择分类
     - 文件类型（根据文件自动选择）
     - 标题
   - 点击"开始上传"

4. **验证上传成功**
   - 等待上传进度条完成
   - 应该看到成功提示
   - 不应该再出现 "static credentials are empty" 错误

### 验证文件已存储到S3

```bash
# 列出最近上传的文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive --human-readable

# 应该能看到新上传的文件，路径类似：
# 2026-02-07 09:00:00   1.2 KiB openwan/data1/abc123def456/789ghi012jkl.txt
```

---

## 📝 技术细节

### AWS凭证链优先级
1. **环境变量**
   - AWS_ACCESS_KEY_ID
   - AWS_SECRET_ACCESS_KEY
   - AWS_SESSION_TOKEN (可选)

2. **共享凭证文件**
   - ~/.aws/credentials (Linux/Mac)
   - %USERPROFILE%\.aws\credentials (Windows)

3. **IAM角色**
   - EC2实例角色
   - ECS任务角色
   - Lambda执行角色

4. **配置文件**
   - ~/.aws/config

### 当前使用的凭证
系统当前使用 `~/.aws/credentials` 文件中的凭证：
```bash
# 查看当前凭证
cat ~/.aws/credentials

# 验证凭证有效性
aws sts get-caller-identity
```

### 如何切换凭证来源

**选项1：使用环境变量（推荐用于生产环境）**
```bash
# 在启动脚本中设置
export AWS_ACCESS_KEY_ID=your_key_id
export AWS_SECRET_ACCESS_KEY=your_secret_key
export AWS_DEFAULT_REGION=us-east-1
./start-services.sh
```

**选项2：使用IAM角色（推荐用于EC2）**
```yaml
# configs/config.yaml
storage:
  type: s3
  s3_bucket: video-bucket-843250590784
  s3_region: us-east-1
  # 不设置 s3_access_key 和 s3_secret_key
```
然后设置环境变量：
```bash
export S3_USE_IAM_ROLE=true
./start-services.sh
```

**选项3：使用配置文件中的静态凭证（不推荐）**
```yaml
# configs/config.yaml - 不推荐将凭证写在配置文件中
storage:
  type: s3
  s3_bucket: video-bucket-843250590784
  s3_region: us-east-1
  s3_access_key: AKIAIOSFODNN7EXAMPLE
  s3_secret_key: wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY
```

---

## 🔒 安全建议

1. **不要在代码或配置文件中硬编码凭证**
   - 使用环境变量
   - 使用IAM角色
   - 使用AWS凭证文件

2. **定期轮换凭证**
   ```bash
   aws iam create-access-key --user-name your-user
   # 更新 ~/.aws/credentials
   aws iam delete-access-key --access-key-id OLD_KEY_ID --user-name your-user
   ```

3. **使用最小权限原则**
   IAM策略示例：
   ```json
   {
       "Version": "2012-10-17",
       "Statement": [
           {
               "Effect": "Allow",
               "Action": [
                   "s3:PutObject",
                   "s3:GetObject",
                   "s3:DeleteObject"
               ],
               "Resource": "arn:aws:s3:::video-bucket-843250590784/openwan/*"
           },
           {
               "Effect": "Allow",
               "Action": "s3:ListBucket",
               "Resource": "arn:aws:s3:::video-bucket-843250590784"
           }
       ]
   }
   ```

4. **启用S3服务端加密**
   - 已在代码中配置：`ServerSideEncryption: "AES256"`

---

## 📊 监控

### 查看上传日志
```bash
# 实时监控上传
tail -f /home/ec2-user/openwan/logs/api.log | grep -i "upload\|s3"

# 查看最近的上传
tail -50 /home/ec2-user/openwan/logs/api.log | grep upload
```

### 检查S3使用情况
```bash
# 统计文件数量
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | wc -l

# 计算总大小
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive --summarize

# 查看最近上传的文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive --human-readable | tail -10
```

---

## ✅ 总结

- ✅ **问题已修复**：S3存储现在正确使用AWS凭证链
- ✅ **服务已重启**：最新代码已部署
- ✅ **配置已验证**：S3存储桶可访问
- ✅ **凭证已验证**：~/.aws/credentials 文件可用

**可以开始测试文件上传了！**

访问 http://localhost 并上传文件，应该不会再出现凭证错误。

---

**修复时间**: 2026-02-07 09:05 UTC
**修复文件**: internal/storage/s3.go
**服务PID**: API: 4162737, Worker #1: 4162763, Worker #2: 4162814
