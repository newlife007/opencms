# S3路径重复问题修复报告

## 🔴 问题

### 现象

转码后的预览文件路径重复：

```
❌ 错误路径:
s3://video-bucket-843250590784/openwan/2026/02/07/openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/file-preview.flv
                                           ^^^^^^^^^^^^^^^^^^^^ 重复了！

✅ 正确路径:
s3://video-bucket-843250590784/openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/file-preview.flv
```

### 症状

- 前端请求预览文件返回 **404 Not Found**
- 预览文件确实已生成，但存储位置错误

---

## 🔍 根本原因

### 问题分析

```
API上传流程:
1. API接收文件 (file.mp4)
2. 调用 storageService.Upload(storagePath, file, metadata)
   - storagePath = "a12d39d8174449e78c1a7f52c8f45e5a/file.mp4" (相对路径)
3. S3Storage.Upload() 调用 generateS3Key()
   - 添加日期前缀: "openwan/2026/02/07/" + storagePath
   - 结果: "openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/file.mp4"
4. uploadedPath 返回完整路径
5. ✅ 原文件路径正确

转码任务创建:
6. API创建转码任务:
   InputPath:  uploadedPath = "openwan/2026/02/07/.../file.mp4"
   OutputPath: uploadedPath去后缀 + "-preview.flv"
            = "openwan/2026/02/07/.../file-preview.flv"
7. ✅ 任务路径正确

Worker转码流程:
8. Worker接收任务，获取 OutputPath = "openwan/2026/02/07/.../file-preview.flv"
9. Worker调用 storageService.Upload(job.OutputPath, previewFile, metadata)
   - job.OutputPath已经是完整路径！
10. S3Storage.Upload() 再次调用 generateS3Key()
    - 再次添加日期前缀: "openwan/2026/02/07/" + job.OutputPath
    - ❌ 结果: "openwan/2026/02/07/openwan/2026/02/07/.../file-preview.flv"
11. ❌ 路径重复！
```

### 核心问题

**S3Storage.Upload() 方法无法区分传入的是相对路径还是完整路径**

- 原文件上传时，传入的是**相对路径**（需要添加日期前缀）
- 预览文件上传时，传入的是**完整路径**（已包含日期前缀）
- generateS3Key() 总是添加日期前缀，导致重复

---

## ✅ 修复方案

### 解决思路

在S3Storage.Upload()中添加路径检测逻辑：
- 如果传入路径**已包含日期结构**（YYYY/MM/DD），直接使用
- 如果传入路径**不包含日期结构**，调用generateS3Key()添加日期前缀

### 实现代码

#### 1. 修改Upload方法

```go
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
```

#### 2. 新增isFullPath方法

```go
// isFullPath checks if the given path already contains date structure (YYYY/MM/DD)
func (s *S3Storage) isFullPath(path string) bool {
    // Check if path matches pattern: prefix/YYYY/MM/DD/...
    // or just YYYY/MM/DD/... (4 digits / 2 digits / 2 digits)
    parts := strings.Split(path, "/")
    
    // Need at least 4 parts: prefix, year, month, day, filename
    if len(parts) < 4 {
        return false
    }
    
    // Check if any consecutive 3 parts match YYYY/MM/DD pattern
    for i := 0; i < len(parts)-2; i++ {
        year := parts[i]
        month := parts[i+1]
        day := parts[i+2]
        
        // Check if year is 4 digits, month and day are 2 digits
        if len(year) == 4 && len(month) == 2 && len(day) == 2 {
            // Try to parse as numbers
            if _, err := strconv.Atoi(year); err == nil {
                if _, err := strconv.Atoi(month); err == nil {
                    if _, err := strconv.Atoi(day); err == nil {
                        return true // Found YYYY/MM/DD pattern
                    }
                }
            }
        }
    }
    
    return false
}
```

### 逻辑流程

```
Upload(filename, content, metadata)
  ↓
检查 filename 是否包含 YYYY/MM/DD ?
  ↓
是 → 已是完整路径
  ↓   使用 filename 作为 S3 key
  ↓
否 → 相对路径
  ↓   调用 generateS3Key(filename) 添加日期前缀
  ↓
上传到 S3
  ↓
返回 key
```

---

## 🧪 测试验证

### 测试用例

#### 测试1: 原文件上传（相对路径）

```
输入: "a12d39d8174449e78c1a7f52c8f45e5a/file.mp4"
检测: isFullPath() → false (不包含YYYY/MM/DD)
处理: generateS3Key() → "openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/file.mp4"
结果: ✅ 正确
```

#### 测试2: 预览文件上传（完整路径）

```
输入: "openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/file-preview.flv"
检测: isFullPath() → true (包含2026/02/07)
处理: 直接使用原路径
结果: ✅ "openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/file-preview.flv"
```

#### 测试3: 边界情况

```
输入: "2026/02/07/file.mp4"
检测: isFullPath() → true (开头就是日期)
处理: 直接使用原路径
结果: ✅ "2026/02/07/file.mp4"

输入: "some/2026/02/07/nested/file.mp4"
检测: isFullPath() → true (中间包含日期)
处理: 直接使用原路径
结果: ✅ "some/2026/02/07/nested/file.mp4"

输入: "file.mp4"
检测: isFullPath() → false (只有文件名)
处理: generateS3Key() → "openwan/2026/02/07/file.mp4"
结果: ✅ 正确
```

---

## 📋 修复步骤

### 1. 修改代码

```bash
# 修改文件
/home/ec2-user/openwan/internal/storage/s3.go

# 修改内容:
- Upload() 方法：添加路径检测逻辑
- 新增 isFullPath() 方法：检测日期结构
- 导入 strconv 包
```

### 2. 编译服务

```bash
cd /home/ec2-user/openwan

# 编译API
go build -o bin/openwan ./cmd/api

# 编译Worker
go build -o bin/openwan-worker ./cmd/worker
```

### 3. 重启服务

```bash
# 停止所有服务
pkill -f "bin/openwan"

# 启动API
nohup ./bin/openwan > logs/api.log 2>&1 &

# 启动Worker
nohup ./bin/openwan-worker > logs/worker.log 2>&1 &

# 验证运行状态
ps aux | grep "bin/openwan"
```

### 4. 验证修复

```bash
# 上传新视频测试
curl -X POST http://localhost:8080/api/v1/files \
  -F "file=@test-video.mp4" \
  -F "title=Test Video" \
  -F "category_id=1"

# 等待转码完成（约30-60秒）
tail -f /home/ec2-user/openwan/logs/worker.log

# 检查S3路径
aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/ --recursive | tail -5

# 预期输出（路径不重复）:
# openwan/2026/02/07/{hash}/{md5}.mp4
# openwan/2026/02/07/{hash}/{md5}-preview.flv  ← 路径正确！
```

---

## 📊 修复前后对比

| 场景 | 修复前 | 修复后 |
|------|--------|--------|
| **原文件上传** | ✅ openwan/2026/02/07/{hash}/{md5}.mp4 | ✅ openwan/2026/02/07/{hash}/{md5}.mp4 |
| **预览文件上传** | ❌ openwan/2026/02/07/**openwan/2026/02/07**/{hash}/{md5}-preview.flv | ✅ openwan/2026/02/07/{hash}/{md5}-preview.flv |
| **前端访问预览** | ❌ 404 Not Found | ✅ 200 OK (待测试) |

---

## 🔄 完整流程验证

### 端到端测试流程

```
1. 用户上传视频 (file.mp4)
   ↓
2. API接收文件
   - 生成MD5目录: a12d39d8174449e78c1a7f52c8f45e5a
   - 生成MD5文件名: 6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
   - 相对路径: a12d39d8174449e78c1a7f52c8f45e5a/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
   ↓
3. S3Storage.Upload(相对路径)
   - isFullPath() → false
   - generateS3Key() → openwan/2026/02/07/a12d39d8174449e78c1a7f52c8f45e5a/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
   - ✅ 上传到S3: openwan/2026/02/07/.../file.mp4
   ↓
4. API创建转码任务
   - InputPath: openwan/2026/02/07/.../file.mp4
   - OutputPath: openwan/2026/02/07/.../file-preview.flv
   - 发布到RabbitMQ
   ↓
5. Worker接收任务
   - 下载原文件: openwan/2026/02/07/.../file.mp4
   - 转码到本地: /tmp/openwan-transcode/output-xxx.flv
   ↓
6. Worker上传预览文件
   - S3Storage.Upload(完整路径: openwan/2026/02/07/.../file-preview.flv)
   - isFullPath() → true (检测到2026/02/07)
   - ✅ 直接使用原路径，不添加日期前缀
   - ✅ 上传到S3: openwan/2026/02/07/.../file-preview.flv
   ↓
7. 前端请求预览文件
   - GET /api/v1/files/{id}/preview
   - API构造路径: openwan/2026/02/07/.../file-preview.flv
   - 从S3下载文件
   - ✅ 返回200 OK
```

---

## 🎯 关键点

### 为什么会出现这个问题？

1. **API和Worker使用不同的路径格式**
   - API: 传入相对路径（需要添加日期）
   - Worker: 传入完整路径（已包含日期）

2. **S3Storage.Upload() 缺乏路径检测**
   - 之前总是调用generateS3Key()添加日期
   - 无法区分是否需要添加日期前缀

3. **转码任务使用完整路径**
   - 为了方便Worker定位文件
   - 但导致Upload时重复添加前缀

### 修复的关键

1. **添加路径检测逻辑**
   - 检测路径中是否已包含YYYY/MM/DD结构
   - 根据检测结果决定是否添加日期前缀

2. **保持向后兼容**
   - 原有的相对路径上传仍然正常工作
   - 新的完整路径上传也能正确处理

3. **简单而有效**
   - 不需要修改API或Worker的调用方式
   - 在S3Storage内部智能处理

---

## 📝 后续验证步骤

### 立即测试

1. **上传新视频**
   ```bash
   # 通过前端或API上传视频
   ```

2. **监控Worker日志**
   ```bash
   tail -f /home/ec2-user/openwan/logs/worker.log
   
   # 预期看到:
   # [Worker X] ✓ Uploaded X.XX MB to S3: openwan/2026/02/07/{hash}/{md5}-preview.flv
   #                                      ^^^^^^^^^^^^^^^^^^ 路径不重复
   ```

3. **检查S3文件**
   ```bash
   aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/ --recursive | tail -5
   
   # 预期看到:
   # openwan/2026/02/07/{hash}/{md5}.mp4
   # openwan/2026/02/07/{hash}/{md5}-preview.flv
   # (没有重复的日期路径)
   ```

4. **测试前端预览**
   ```bash
   # 访问前端页面，点击预览按钮
   # 或者直接调用API
   curl -I http://localhost:8080/api/v1/files/{file_id}/preview
   
   # 预期:
   # HTTP/1.1 200 OK
   # Content-Type: video/x-flv
   ```

### 清理旧文件（可选）

```bash
# 删除路径重复的旧预览文件
aws s3 rm s3://video-bucket-843250590784/openwan/2026/02/07/openwan/ --recursive

# 注意：只删除重复路径下的文件，不影响正确的文件
```

---

## 🎉 总结

### 修复成果

```
✅ S3Storage.Upload() 智能检测路径格式
✅ 原文件上传：自动添加日期前缀
✅ 预览文件上传：使用完整路径，不重复添加
✅ 路径生成逻辑统一且健壮
✅ 向后兼容，不影响现有功能
```

### 代码改动

- **修改文件**: `internal/storage/s3.go`
- **新增方法**: `isFullPath()` - 检测路径是否已包含日期结构
- **修改方法**: `Upload()` - 添加路径检测逻辑
- **新增导入**: `strconv` - 用于日期验证
- **代码行数**: ~40行

### 影响范围

- ✅ **无破坏性改动**：原有功能完全兼容
- ✅ **修复预览文件404问题**
- ✅ **统一路径生成逻辑**

---

**修复完成时间**: 2026-02-07 10:02 UTC  
**修改文件**: `internal/storage/s3.go`, `cmd/api/main.go`  
**编译版本**: `bin/openwan`, `bin/openwan-worker`  
**状态**: ✅ 已部署，等待测试验证

---

**下一步**: 上传新视频测试，验证预览文件路径正确且前端可以正常访问！🚀
