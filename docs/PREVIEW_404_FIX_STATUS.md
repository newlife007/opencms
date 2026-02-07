# 预览文件404问题修复状态报告

**日期**: 2026-02-07  
**问题**: 前端访问预览文件返回404，但预览文件已在S3上成功生成  
**状态**: 部分修复（1/2完成）

---

## 📊 修复进度

### ✅ 1. S3路径重复问题（已修复）

**问题描述**:
Worker上传预览文件时，S3Storage.Upload()方法重复添加日期前缀，导致路径错误。

**修复位置**: `internal/storage/s3.go`

**修复内容**:
```go
// 在Upload()方法中添加路径检测逻辑
func (s *S3Storage) Upload(ctx context.Context, filename string, content io.Reader, metadata map[string]string) (string, error) {
    // Check if filename already contains date path structure (e.g., starts with prefix/YYYY/MM/DD/)
    // If so, use it as-is; otherwise generate S3 key with date structure
    var key string
    if s.isFullPath(filename) {
        // Already a full path (e.g., from transcoding job), use as-is
        key = filename
    } else {
        // Generate S3 key with date structure for new uploads
        key = s.generateS3Key(filename)
    }
    
    // ... rest of upload logic
}

// 新增方法: 检测路径是否已包含YYYY/MM/DD结构
func (s *S3Storage) isFullPath(path string) bool {
    parts := strings.Split(path, "/")
    if len(parts) < 4 {
        return false
    }
    
    // Check if any consecutive 3 parts match YYYY/MM/DD pattern
    for i := 0; i < len(parts)-2; i++ {
        year, month, day := parts[i], parts[i+1], parts[i+2]
        if len(year) == 4 && len(month) == 2 && len(day) == 2 {
            if _, err := strconv.Atoi(year); err == nil {
                if _, err := strconv.Atoi(month); err == nil {
                    if _, err := strconv.Atoi(day); err == nil {
                        return true
                    }
                }
            }
        }
    }
    return false
}
```

**验证结果**:
```bash
# Worker日志显示上传成功，路径正确
[Worker 1] ✓ Uploaded 8.14 MB to S3: openwan/2026/02/07/33ab512143b66df625abaec6521383a3/6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv

# S3文件确认存在
$ aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/33ab512143b66df625abaec6521383a3/
2026-02-07 10:04:12    8538824 openwan/.../6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv ✅
2026-02-07 10:04:09    8924094 openwan/.../6c2c0a46a93a1316d3beb8e2504ebcf7.mp4         ✅
```

**状态**: ✅ **已完成并验证**

---

### ❌ 2. PreviewFile API路径生成问题（未修复）

**问题描述**:
`internal/api/handlers/files.go`的PreviewFile函数构建预览路径的逻辑可能有问题（待验证）。

**问题位置**: `internal/api/handlers/files.go` 第560-563行

**当前代码**:
```go
// Try preview file: {name}-preview.flv
dir := filepath.Dir(file.Path)
previewFilename := file.Name + "-preview.flv"
previewPath := filepath.Join(dir, previewFilename)
```

**问题分析**:
根据路径逻辑分析，当前代码应该能正确生成路径：
- `file.Path = "openwan/2026/02/07/.../6c2c0a46a93a1316d3beb8e2504ebcf7.mp4"`
- `file.Name = "6c2c0a46a93a1316d3beb8e2504ebcf7"`
- `dir = "openwan/2026/02/07/..."`
- `previewPath = "openwan/2026/02/07/.../6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv"` ✅

**但为了确保S3路径处理的健壮性，建议修改为**:

```go
// Try preview file: replace extension with -preview.flv
// Use string manipulation to handle S3 paths correctly
ext := filepath.Ext(file.Path)
previewPath := strings.TrimSuffix(file.Path, ext) + "-preview.flv"
```

**修复状态**: ❌ **未完成**

**原因**: 
- 文件太大（约76KB），editor工具修改失败
- 多次尝试使用sed/patch/Python脚本均因shell heredoc语法问题失败

**测试结果**:
```bash
$ curl -I http://localhost:8080/api/v1/files/32/preview
HTTP/1.1 404 Not Found ❌
```

---

## 🔍 根本原因分析

### 可能的原因

#### 1. ✅ S3路径重复（已排除）
- Worker上传时路径重复
- **已修复并验证**

#### 2. ⚠️ API路径生成错误（需验证）
- PreviewFile构建的路径与S3实际路径不匹配
- **需要添加日志确认**

#### 3. ⚠️ 数据库path字段不正确（需验证）
- 数据库中存储的file.Path可能与S3实际路径不一致
- **需要查询数据库确认**

#### 4. ⚠️ S3Download方法问题（待排查）
- S3Storage.Download()方法可能处理路径有误
- **需要检查代码逻辑**

---

## 📋 手动修复步骤

### 方法1: 直接编辑文件（推荐）

```bash
# 1. 备份文件
cp /home/ec2-user/openwan/internal/api/handlers/files.go \
   /home/ec2-user/openwan/internal/api/handlers/files.go.backup

# 2. 使用vim或nano编辑文件
vim /home/ec2-user/openwan/internal/api/handlers/files.go

# 3. 跳转到第560行
:560

# 4. 将以下4行:
		// Try preview file: {name}-preview.flv
		dir := filepath.Dir(file.Path)
		previewFilename := file.Name + "-preview.flv"
		previewPath := filepath.Join(dir, previewFilename)

# 5. 替换为:
		// Try preview file: replace extension with -preview.flv
		// Use string manipulation to handle S3 paths correctly
		ext := filepath.Ext(file.Path)
		previewPath := strings.TrimSuffix(file.Path, ext) + "-preview.flv"

# 6. 保存并退出
:wq

# 7. 重新编译
cd /home/ec2-user/openwan
go build -o bin/openwan ./cmd/api

# 8. 重启API服务
pkill -f "bin/openwan" && nohup ./bin/openwan > logs/api.log 2>&1 &

# 9. 测试
curl -I http://localhost:8080/api/v1/files/32/preview
```

### 方法2: 添加调试日志（诊断用）

在第563行后添加日志：

```go
previewPath := filepath.Join(dir, previewFilename)

// 添加调试日志
log.Printf("[PreviewFile DEBUG] fileID=%d file.Name=%s file.Path=%s previewPath=%s", 
    file.ID, file.Name, file.Path, previewPath)

reader, err = h.storageService.Download(c.Request.Context(), previewPath)
```

然后：
```bash
# 重新编译和重启
go build -o bin/openwan ./cmd/api
pkill -f "bin/openwan" && nohup ./bin/openwan > logs/api.log 2>&1 &

# 测试并查看日志
curl -I http://localhost:8080/api/v1/files/32/preview
tail -f logs/api.log | grep "PreviewFile DEBUG"
```

---

## 🔧 验证检查清单

### 1. 检查数据库中的路径
```bash
# 连接MySQL查看file 32的path字段
docker exec openwan-mysql mysql -u root -prootpassword openwan_db \
  -e "SELECT id, name, path FROM ow_files WHERE id=32\G"
```

**预期结果**:
```
id: 32
name: 6c2c0a46a93a1316d3beb8e2504ebcf7
path: openwan/2026/02/07/33ab512143b66df625abaec6521383a3/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
```

### 2. 检查S3文件是否可访问
```bash
# 直接下载预览文件测试
aws s3 cp s3://video-bucket-843250590784/openwan/2026/02/07/33ab512143b66df625abaec6521383a3/6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv /tmp/test.flv

# 检查文件大小
ls -lh /tmp/test.flv
```

### 3. 检查S3Storage.Download()方法
查看 `internal/storage/s3.go` 的Download方法是否正确处理路径：

```go
func (s *S3Storage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
    // 确认key没有被再次修改
    log.Printf("[S3Storage.Download] key=%s", key)
    
    input := &s3.GetObjectInput{
        Bucket: aws.String(s.bucket),
        Key:    aws.String(key),  // ← 确认这里直接使用传入的key
    }
    
    // ...
}
```

---

## 📝 修复完成后的验证

### 验证步骤

1. **重新编译并重启服务**
```bash
cd /home/ec2-user/openwan
go build -o bin/openwan ./cmd/api
pkill -f "bin/openwan"
nohup ./bin/openwan > logs/api.log 2>&1 &
```

2. **测试预览文件访问**
```bash
# HEAD请求
curl -I http://localhost:8080/api/v1/files/32/preview

# 预期结果:
HTTP/1.1 200 OK
Content-Type: video/x-flv
Content-Length: 8538824
```

3. **实际下载预览文件**
```bash
curl -o /tmp/preview.flv http://localhost:8080/api/v1/files/32/preview

# 检查文件大小
ls -lh /tmp/preview.flv
# 预期: 8.1M

# 播放测试
ffplay /tmp/preview.flv
```

4. **前端测试**
- 访问前端页面
- 找到file 32
- 点击预览按钮
- 确认视频能正常播放

---

## 📊 当前状态总结

| 问题 | 状态 | 优先级 |
|------|------|--------|
| S3路径重复 | ✅ 已修复 | 高 |
| API路径生成 | ❌ 未修复 | 高 |
| 预览文件404 | ❌ 未解决 | 高 |

**下一步行动**:
1. ✅ 手动编辑 `files.go` 修复PreviewFile函数
2. 📋 添加调试日志确认路径生成正确
3. 🔍 检查数据库path字段
4. ✅ 测试验证修复效果

---

## 🎯 成功标准

- [x] Worker转码成功
- [x] 预览文件上传到S3正确路径
- [ ] API能找到并下载预览文件
- [ ] 前端能正常播放预览视频
- [ ] 路径生成逻辑经过测试验证

**修复完成率**: 50% (1/2)

---

**最后更新**: 2026-02-07 10:18 UTC  
**更新人**: AWS Transform CLI Agent
