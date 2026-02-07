#!/bin/bash

echo "========================================="
echo "OpenWan VideoPlayer 问题诊断工具"
echo "========================================="
echo ""

# 检查前端服务
echo "1. 检查前端开发服务器..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "✓ 前端服务运行正常 (http://localhost:3000)"
else
    echo "✗ 前端服务未运行"
    echo "  启动命令: cd /home/ec2-user/openwan/frontend && screen -dmS vite npm run dev"
fi
echo ""

# 检查后端服务
echo "2. 检查后端API服务器..."
HEALTH=$(curl -s http://localhost:8080/health | jq -r '.status' 2>/dev/null)
if [ -n "$HEALTH" ]; then
    echo "✓ 后端服务运行正常 (状态: $HEALTH)"
else
    echo "✗ 后端服务未运行"
    echo "  启动命令: cd /home/ec2-user/openwan && screen -dmS openwan ./bin/openwan"
fi
echo ""

# 测试登录和文件API
echo "3. 测试API端点..."
LOGIN_RESULT=$(curl -s -c /tmp/test-cookies.txt -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}' 2>&1)

if echo "$LOGIN_RESULT" | grep -q "success\|token\|user"; then
    echo "✓ 登录API工作正常"
    
    # 测试文件详情API
    FILE_RESULT=$(curl -s -b /tmp/test-cookies.txt http://localhost:8080/api/v1/files/71 2>&1)
    if echo "$FILE_RESULT" | grep -q '"id"'; then
        echo "✓ 文件详情API工作正常"
        FILE_TYPE=$(echo "$FILE_RESULT" | jq -r '.data.type // .type' 2>/dev/null)
        echo "  文件类型: $FILE_TYPE (1=视频, 2=音频, 3=图片, 4=富媒体)"
    else
        echo "✗ 文件详情API失败"
        echo "  响应: $FILE_RESULT"
    fi
    
    # 测试预览API
    PREVIEW_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -b /tmp/test-cookies.txt \
      http://localhost:8080/api/v1/files/71/preview)
    if [ "$PREVIEW_STATUS" == "200" ]; then
        echo "✓ 预览API工作正常 (HTTP $PREVIEW_STATUS)"
        
        # 检查文件头
        FILE_HEADER=$(curl -s -b /tmp/test-cookies.txt http://localhost:8080/api/v1/files/71/preview | \
          head -c 20 | od -An -tx1 | head -1)
        echo "  文件头: $FILE_HEADER"
    else
        echo "✗ 预览API失败 (HTTP $PREVIEW_STATUS)"
    fi
else
    echo "✗ 登录API失败"
    echo "  响应: $LOGIN_RESULT"
fi
echo ""

# 检查VideoPlayer组件
echo "4. 检查VideoPlayer组件文件..."
if [ -f "/home/ec2-user/openwan/frontend/src/components/VideoPlayer.vue" ]; then
    echo "✓ VideoPlayer.vue 存在"
    LINE_COUNT=$(wc -l < /home/ec2-user/openwan/frontend/src/components/VideoPlayer.vue)
    echo "  文件大小: $LINE_COUNT 行"
else
    echo "✗ VideoPlayer.vue 不存在"
fi
echo ""

# 检查依赖
echo "5. 检查video.js依赖..."
cd /home/ec2-user/openwan/frontend
if npm list video.js 2>&1 | grep -q "video.js@"; then
    VIDEO_VERSION=$(npm list video.js 2>&1 | grep "video.js@" | head -1 | sed 's/.*video.js@//' | sed 's/ .*//')
    echo "✓ video.js 已安装 (版本: $VIDEO_VERSION)"
else
    echo "✗ video.js 未安装"
fi

if npm list videojs-flvjs-es6 2>&1 | grep -q "videojs-flvjs-es6@"; then
    FLV_VERSION=$(npm list videojs-flvjs-es6 2>&1 | grep "videojs-flvjs-es6@" | head -1 | sed 's/.*videojs-flvjs-es6@//' | sed 's/ .*//')
    echo "✓ videojs-flvjs-es6 已安装 (版本: $FLV_VERSION)"
else
    echo "✗ videojs-flvjs-es6 未安装"
fi
echo ""

# 提供测试URL
echo "========================================="
echo "测试页面URL:"
echo "========================================="
echo "1. 文件详情页 (带详细调试信息):"
echo "   http://localhost:3000/files/71"
echo ""
echo "2. VideoPlayer独立测试页:"
echo "   http://localhost:3000/test-video"
echo ""
echo "3. 如果您的EC2实例有公网IP，将localhost替换为:"
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null)
if [ -n "$PUBLIC_IP" ]; then
    echo "   http://$PUBLIC_IP:3000/files/71"
    echo "   http://$PUBLIC_IP:3000/test-video"
else
    echo "   http://<您的EC2公网IP>:3000/files/71"
fi
echo ""

echo "========================================="
echo "期望在页面上看到的调试信息:"
echo "========================================="
echo "调试信息:"
echo "fileInfo.type = 1"
echo "previewUrl = /api/v1/files/71/preview"
echo "条件1 (type === 1): true"
echo "条件2 (type === 2): false"
echo "条件3 (previewUrl存在): true"
echo "最终条件: true"
echo ""
echo "如果最终条件为 true，应该看到VideoPlayer组件"
echo "如果最终条件为 false，请截图告诉我哪个条件失败了"
echo ""

echo "========================================="
echo "如果遇到问题，请运行:"
echo "========================================="
echo "# 查看前端日志"
echo "tail -f /tmp/vite-reload.log"
echo ""
echo "# 查看后端日志"
echo "screen -r openwan"
echo ""
echo "# 重启服务"
echo "./restart-services.sh"
echo ""

echo "诊断完成！"
