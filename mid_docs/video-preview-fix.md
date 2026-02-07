# Video Preview Issue and Resolution

## Problem
Video files are not displaying or playing in the frontend file detail view.

## Root Causes Identified

### 1. Component Naming Issue (FIXED)
**Problem:** Template used `<video-player>` (kebab-case) but component was registered as `VideoPlayer` (PascalCase).

**Fix Applied:**
- Updated `frontend/src/views/files/FileDetail.vue` line 18
- Changed `<video-player>` to `<VideoPlayer>` to match the component registration

### 2. FLV.js Tech Registration Issue (FIXED)
**Problem:** The `videojs-flvjs-es6` plugin requires `flvjs` to be available globally on the `window` object.

**Fix Applied:**
- Updated `frontend/src/components/VideoPlayer.vue`
- Added explicit flv.js import: `import flvjs from 'flv.js'`
- Added global registration: `window.flvjs = flvjs`

### 3. Missing Preview Files (PRIMARY ISSUE)
**Problem:** Video preview files (`{filename}-preview.flv`) do not exist because transcoding workers are not running.

**Evidence:**
```bash
# Check for video files in database
SELECT id, title, name, type FROM ow_files WHERE type=1 LIMIT 5;
# Results show 5 video files (IDs: 17, 18, 20, 21, 25)

# Check storage for preview files
find /home/ec2-user/openwan/storage/data1 -name "*-preview.flv"
# No results - preview files don't exist

# Check for running worker services
docker ps | grep worker
# No worker containers running
```

**Backend Behavior:**
When preview endpoint is called (`/api/v1/files/{id}/preview`), the backend:
1. Determines preview path as `{dir}/{filename}-preview.flv`
2. Attempts to download from storage
3. Returns 404 error: "Preview file not available. Transcoding may be in progress."

## Solution

### Immediate Fix: Enhanced Error Handling (COMPLETED)

**VideoPlayer Component Improvements:**
```javascript
// Added comprehensive logging
console.log('Initializing video player with src:', props.src)

// Added error event handlers
player.on('error', (error) => {
  const err = player.error()
  console.error('Video player error:', {
    code: err?.code,
    message: err?.message,
    type: err?.type,
    raw: error
  })
})

// Added lifecycle event logging
player.on('loadstart', () => console.log('Video load started'))
player.on('canplay', () => console.log('Video can play'))
player.on('play', () => console.log('Video playing'))
```

### Long-term Fix: Enable Transcoding Workers

**Option 1: Start Worker in Docker Compose**
```bash
# Check if worker service is defined
docker-compose config | grep -A5 worker

# Start worker service
cd /home/ec2-user/openwan
docker-compose up -d worker
```

**Option 2: Build and Run Worker Manually**
```bash
# Build worker application
cd /home/ec2-user/openwan
go build -o bin/worker ./cmd/worker

# Run worker
./bin/worker
```

**Option 3: Trigger Transcoding via API**
If transcoding endpoint exists:
```bash
curl -X POST http://localhost:8080/api/v1/files/17/transcode \
  -H "Cookie: openwan_session=<session_id>"
```

## Transcoding Workflow

Based on code analysis (`internal/transcoding/` and `cmd/worker/`):

1. **File Upload**: User uploads video file
2. **Message Queue**: Upload handler publishes transcode job to RabbitMQ queue
3. **Worker Consumption**: Worker service consumes job from queue
4. **FFmpeg Execution**: Worker runs FFmpeg to generate preview FLV:
   ```
   ffmpeg -i "{inputFile}" -y -ab 56 -ar 22050 -r 15 -b 500 -s 320x240 "{outputFile}"
   ```
5. **Storage**: Preview file saved to `{dir}/{filename}-preview.flv`
6. **Frontend Access**: VideoPlayer loads preview via `/api/v1/files/{id}/preview`

## Testing After Fix

### 1. Verify Frontend Build
```bash
cd /home/ec2-user/openwan/frontend
npm run build
# ✅ Build completed successfully with videojs chunks
```

### 2. Test VideoPlayer Component
1. Navigate to file detail page: `http://localhost:3000/files/17`
2. Open browser console (F12)
3. Look for messages:
   - "Initializing video player with src: /api/v1/files/17/preview"
   - "Video.js player is ready"
   - "Tech in use: flvjs" (if FLV file available)
   - OR error message if preview file missing

### 3. Check Backend Logs
```bash
# Backend should log preview file access attempts
grep "preview" /path/to/backend/logs
```

### 4. Verify Preview File Creation
After enabling workers and transcoding:
```bash
# Check for preview files
find /home/ec2-user/openwan/storage -name "*-preview.flv"

# Check specific video's preview
ls -lh /home/ec2-user/openwan/storage/data1/{hash_dir}/*-preview.flv
```

## Fallback for Testing

If transcoding cannot be enabled immediately, implement a fallback in FileDetail.vue:

```vue
<div v-if="fileInfo.type === 1" class="video-preview">
  <VideoPlayer
    v-if="previewUrl && previewExists"
    :src="previewUrl"
    type="video/x-flv"
  />
  <div v-else class="preview-pending">
    <el-icon :size="80"><VideoCamera /></el-icon>
    <p>视频预览正在生成中...</p>
    <el-button @click="checkPreview" type="primary" size="small">
      刷新状态
    </el-button>
  </div>
</div>
```

## Related Files Modified
1. `/home/ec2-user/openwan/frontend/src/components/VideoPlayer.vue`
   - Added flv.js global registration
   - Enhanced error handling and logging
   - Added lifecycle event tracking

2. `/home/ec2-user/openwan/frontend/src/views/files/FileDetail.vue`
   - Fixed component naming from `<video-player>` to `<VideoPlayer>`

3. Frontend rebuilt: `npm run build` completed successfully

## Next Steps
1. ✅ **Fixed**: Component naming and flv.js registration
2. ✅ **Fixed**: Enhanced error handling in VideoPlayer
3. ⚠️ **Pending**: Start transcoding worker service
4. ⚠️ **Pending**: Generate preview files for existing videos
5. ⚠️ **Pending**: Test video playback with actual FLV preview files

## Status
- **Component Issues**: RESOLVED ✅
- **Preview File Generation**: BLOCKED (workers not running) ⚠️
- **Estimated Time to Full Resolution**: 30 minutes (after starting workers and running transcoding)
