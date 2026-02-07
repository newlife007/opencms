# 转码预览文件存储位置说明

## 📍 预览文件存储位置

### S3存储路径结构

预览文件存储在与原文件相同的目录中，文件名为原文件名加上`-preview.flv`后缀。

**路径规则**:
```
原文件: openwan/YYYY/MM/DD/{dir_hash}/{file_hash}.{ext}
预览文件: openwan/YYYY/MM/DD/{dir_hash}/{file_hash}-preview.flv
```

**实例**:
```
原文件ID: 30
上传时间: 2026-02-07
原文件路径: openwan/2026/02/07/968765c419dfa6d808f2172548700e94/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4

预览文件路径: openwan/2026/02/07/968765c419dfa6d808f2172548700e94/6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv
```

### 完整的S3路径

```
Bucket: video-bucket-843250590784
Region: us-east-1

原文件完整路径:
s3://video-bucket-843250590784/openwan/2026/02/07/968765c419dfa6d808f2172548700e94/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4

预览文件完整路径:
s3://video-bucket-843250590784/openwan/2026/02/07/968765c419dfa6d808f2172548700e94/6c2c0a46a93a1316d3beb8e2504ebcf7-preview.flv
```

---

## 🔄 转码流程

### 当前实现的流程

#### 1. 文件上传时
```
用户上传视频
  ↓
保存到S3 (原文件)
  ↓
保存数据库记录
  ↓
发布转码任务到RabbitMQ队列
  ↓
API返回成功响应
```

#### 2. Worker处理 (当前状态：失败)
```
Worker从队列获取任务
  ↓
尝试从S3下载原文件 → ❌ 失败 (Worker无S3配置)
  ↓
重试3次后移到死信队列
  ↓
预览文件不存在
```

#### 3. 访问预览时
```
GET /api/v1/files/30/preview
  ↓
查找数据库中的文件记录
  ↓
尝试从S3下载预览文件
  ↓
预览文件不存在 → 返回404
  ↓
(当前没有触发按需转码)
```

---

## ⚠️ 当前问题

### 问题1: Worker无法访问S3

**现象**:
```
[Worker 2] Processing job for file 30
[Worker 2]   Input: openwan/2026/02/07/968765c419dfa6d808f2172548700e94/6c2c0a46a93a1316d3beb8e2504ebcf7.mp4
[Worker 2]   Storage: s3
❌ Transcoding failed: input file does not exist
```

**原因**: Worker配置中缺少S3访问配置（Bucket、Region、Credentials）

**影响**: 队列中的转码任务全部失败，预览文件无法生成

### 问题2: 同步转码Fallback未触发

**现象**: 队列发布成功，但Worker失败后没有fallback

**原因**: 代码逻辑是：
- 队列发布成功 → 不执行同步转码
- 只有队列服务不可用时 → 才执行同步转码

**问题**: Worker虽然接收了任务，但处理失败，API无法感知

### 问题3: 预览请求时没有按需转码

**现象**: 访问`/api/v1/files/30/preview`返回404

**当前行为**: 预览文件不存在时，直接返回错误

**期望行为**: 预览文件不存在时，触发按需转码

---

## ✅ 解决方案

### 方案1: 配置Worker的S3访问 (推荐)

让Worker能够访问S3文件进行转码。

**步骤**:

1. **检查Worker配置文件**
```bash
cat /home/ec2-user/openwan/cmd/worker/main.go | grep -A20 "LoadConfig\|storage"
```

2. **添加S3环境变量**
```bash
export AWS_REGION=us-east-1
export S3_BUCKET=video-bucket-843250590784
export STORAGE_TYPE=s3
```

3. **或修改Worker代码**
在`cmd/worker/main.go`中添加S3配置：
```go
// 初始化S3存储服务
storageConfig := &storage.Config{
    Type: "s3",
    S3: storage.S3Config{
        Bucket: "video-bucket-843250590784",
        Region: "us-east-1",
        // 使用IAM角色或提供credentials
    },
}
storageService, err := storage.NewStorageService(storageConfig)
```

4. **重新编译和启动Worker**
```bash
cd /home/ec2-user/openwan
go build -o bin/openwan-worker ./cmd/worker
pkill -f "openwan-worker"
nohup ./bin/openwan-worker > logs/worker.log 2>&1 &
```

**优点**:
- ✅ 分布式处理，可扩展
- ✅ 异步处理，不阻塞API
- ✅ 多Worker并发转码

**缺点**:
- 需要修改配置或代码
- 需要重新编译

---

### 方案2: 实现按需转码 (快速方案)

在PreviewFile端点中，当预览文件不存在时触发同步转码。

**实现**:

修改`internal/api/handlers/files.go`的PreviewFile方法：

```go
reader, err = h.storageService.Download(c.Request.Context(), previewPath)
if err != nil {
    // Preview not available, trigger on-demand transcoding
    fmt.Printf("⚠ Preview not available for file %d, triggering on-demand transcode\n", file.ID)
    
    // Start async transcoding in background
    go func(f *models.Files, originalPath string) {
        storageType := "s3"
        fmt.Printf("🎬 Starting on-demand transcode for file %d\n", f.ID)
        err := h.syncTranscodeVideo(f, originalPath, storageType)
        if err != nil {
            fmt.Printf("❌ On-demand transcode failed for file %d: %v\n", f.ID, err)
        } else {
            fmt.Printf("✅ On-demand transcode completed for file %d\n", f.ID)
        }
    }(file, file.Path)
    
    // Fall back to original file for now
    reader, err = h.storageService.Download(c.Request.Context(), file.Path)
    // ... rest of code
}
```

**行为**:
1. 第一次访问预览：返回原文件，后台开始转码
2. 转码完成后（约30秒-2分钟）
3. 第二次访问预览：返回转码后的FLV预览文件

**优点**:
- ✅ 简单快速
- ✅ 无需配置Worker
- ✅ 自动按需生成预览

**缺点**:
- ⚠️ 第一次访问较慢（返回原文件）
- ⚠️ API服务器负载较高

---

### 方案3: 手动触发转码

创建一个管理端点，手动触发失败任务的重新转码。

**实现**:

```go
// POST /api/v1/admin/files/:id/retranscode
func (h *FileHandler) RetranscodeFile() gin.HandlerFunc {
    return func(c *gin.Context) {
        fileID, _ := strconv.ParseUint(c.Param("id"), 10, 32)
        file, err := h.fileService.GetFileByID(c.Request.Context(), uint(fileID))
        if err != nil {
            c.JSON(404, gin.H{"error": "File not found"})
            return
        }
        
        // Trigger sync transcode
        go h.syncTranscodeVideo(file, file.Path, "s3")
        
        c.JSON(200, gin.H{"message": "Transcode started"})
    }
}
```

**使用**:
```bash
# 为文件30重新转码
curl -X POST http://localhost:8080/api/v1/admin/files/30/retranscode \
  -H "Authorization: Bearer ADMIN_TOKEN"
```

---

## 📋 检查现有文件

### 查看S3中的文件

```bash
# 查看文件30的存储
aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/968765c419dfa6d808f2172548700e94/

# 预期输出:
# 2026-02-07 09:36:24    8924094 6c2c0a46a93a1316d3beb8e2504ebcf7.mp4  (原文件)
# (预览文件目前不存在)

# 查看所有预览文件
aws s3 ls s3://video-bucket-843250590784/openwan/ --recursive | grep "preview.flv"

# 当前输出: (空，没有预览文件)
```

### 查看数据库中的文件记录

```bash
mysql -h 127.0.0.1 -u openwan -p openwan_db -e "
SELECT id, title, name, ext, type, status, path 
FROM ow_files 
WHERE type IN (1, 2)  -- 视频和音频
ORDER BY id DESC 
LIMIT 5;
"
```

---

## 🎯 推荐操作

### 立即可行的方案 (方案2)

**修改代码实现按需转码**，这样用户访问预览时会自动触发转码：

1. **修改files.go添加按需转码**
   ```bash
   # 修改PreviewFile方法，在预览不存在时触发转码
   # (代码见方案2)
   ```

2. **重新编译API**
   ```bash
   cd /home/ec2-user/openwan
   go build -o bin/openwan ./cmd/api
   ```

3. **重启API服务**
   ```bash
   pkill -f "bin/openwan"
   nohup ./bin/openwan > logs/api.log 2>&1 &
   ```

4. **测试**
   ```bash
   # 第一次访问（会触发后台转码，返回原文件或404）
   curl -I http://localhost:8080/api/v1/files/30/preview
   
   # 等待1-2分钟后再次访问（应该返回转码后的预览）
   curl -I http://localhost:8080/api/v1/files/30/preview
   ```

---

## 📊 转码进度监控

### 查看API日志
```bash
tail -f /home/ec2-user/openwan/logs/api.log | grep -i "transcode\|preview"
```

### 查看Worker日志
```bash
tail -f /home/ec2-user/openwan/logs/worker.log
```

### 查看S3文件变化
```bash
watch -n 5 "aws s3 ls s3://video-bucket-843250590784/openwan/2026/02/07/968765c419dfa6d808f2172548700e94/"
```

---

## 📚 相关文档

- 后端服务启动: `/home/ec2-user/openwan/docs/BACKEND_STARTUP_FIX.md`
- 转码服务状态: `/home/ec2-user/openwan/docs/TRANSCODING_SERVICE_STATUS.md`
- 视频预览404修复: `/home/ec2-user/openwan/docs/VIDEO_PREVIEW_404_FIX.md`

---

**文档创建时间**: 2026-02-07 09:40 UTC  
**当前预览文件状态**: 不存在（Worker无法访问S3）  
**推荐方案**: 实现按需转码（方案2）或配置Worker S3访问（方案1）
