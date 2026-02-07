#!/bin/bash

echo "============================================"
echo "文件详情页修复 - 快速测试"  
echo "============================================"
echo ""

echo "测试地址:"
echo "  简化版本: http://localhost:3000/files/71"
echo "  测试页面: http://localhost:3000/test-video"
echo ""

# 获取公网IP
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null)
if [ -n "$PUBLIC_IP" ]; then
    echo "如果您的EC2实例有公网IP:"
    echo "  简化版本: http://$PUBLIC_IP:3000/files/71"
    echo "  测试页面: http://$PUBLIC_IP:3000/test-video"
    echo ""
fi

echo "============================================"
echo "自动化测试开始..."
echo "============================================"
echo ""

# 测试前端服务
echo "1. 测试前端服务..."
if curl -s http://localhost:3000 > /dev/null; then
    echo "   ✓ 前端服务正常"
else
    echo "   ✗ 前端服务异常"
    exit 1
fi

# 测试后端API
echo "2. 测试后端API..."
if curl -s http://localhost:8080/health > /dev/null; then
    echo "   ✓ 后端API正常"
else
    echo "   ✗ 后端API异常"
    exit 1
fi

# 登录并测试文件API
echo "3. 测试文件详情API..."
curl -s -c /tmp/test.cookie -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}' > /dev/null

FILE_DATA=$(curl -s -b /tmp/test.cookie http://localhost:8080/api/v1/files/71)
if echo "$FILE_DATA" | jq -e '.data.id or .id' > /dev/null 2>&1; then
    FILE_TYPE=$(echo "$FILE_DATA" | jq -r '.data.type // .type')
    FILE_TITLE=$(echo "$FILE_DATA" | jq -r '.data.title // .title // "无标题"' | head -c 30)
    echo "   ✓ 文件详情获取成功"
    echo "     文件ID: 71"
    echo "     文件类型: $FILE_TYPE (1=视频 2=音频 3=图片 4=富媒体)"
    echo "     文件标题: $FILE_TITLE"
else
    echo "   ✗ 文件详情获取失败"
fi

# 测试预览API
echo "4. 测试预览API..."
PREVIEW_STATUS=$(curl -s -o /dev/null -w "%{http_code}" -b /tmp/test.cookie \
  http://localhost:8080/api/v1/files/71/preview)
if [ "$PREVIEW_STATUS" == "200" ]; then
    echo "   ✓ 预览API正常 (HTTP $PREVIEW_STATUS)"
else
    echo "   ✗ 预览API异常 (HTTP $PREVIEW_STATUS)"
fi

echo ""
echo "============================================"
echo "期望在浏览器中看到的内容:"
echo "============================================"
echo ""
echo "访问 http://localhost:3000/files/71 应该看到："
echo ""
echo "【卡片1: 页面状态】"
echo "  ✓ 页面已加载"
echo "  文件ID: 71"
echo "  Loading: false"
echo "  fileInfo存在: true"
echo ""
echo "【卡片2: 文件信息】"
echo "  ID: 71"
echo "  标题: $FILE_TITLE"
echo "  类型: $FILE_TYPE"
echo "  扩展名: mp4"
echo ""
echo "【卡片3: 预览URL】"
echo "  计算结果: /api/v1/files/71/preview"
echo "  条件检查: type=$FILE_TYPE, isVideo=$([ "$FILE_TYPE" == "1" ] && echo "true" || echo "false")"
echo ""
echo "【卡片4: VideoPlayer组件测试】"
if [ "$FILE_TYPE" == "1" ] || [ "$FILE_TYPE" == "2" ]; then
    echo "  ✓ 条件满足，VideoPlayer应该在下方显示"
    echo "  [应该看到黑色区域内有视频播放器]"
else
    echo "  ✗ 条件不满足 (文件类型不是视频或音频)"
fi
echo ""
echo "============================================"
echo "浏览器控制台应该显示的日志:"
echo "============================================"
echo "[Simple] Component mounted, route: 71"
echo "[Simple] Loading file: 71"
echo "[Simple] Computing previewUrl: {hasId: true, type: $FILE_TYPE, ...}"
echo "[Simple] Preview URL: /api/v1/files/71/preview"
echo "[Simple] File loaded: {id: 71, ...}"
if [ "$FILE_TYPE" == "1" ] || [ "$FILE_TYPE" == "2" ]; then
    echo "[Simple] showVideoPlayer: true"
    echo "Initializing video player with src: /api/v1/files/71/preview"
else
    echo "[Simple] showVideoPlayer: false"
fi
echo ""

echo "============================================"
echo "下一步操作:"
echo "============================================"
echo "1. 在浏览器中打开: http://localhost:3000/files/71"
echo "2. 按 Ctrl+F5 强制刷新"
echo "3. 按 F12 打开开发者工具"
echo "4. 查看 Console 标签页的日志"
echo "5. 告诉我您看到的内容和VideoPlayer是否显示"
echo ""
echo "如果页面能正常打开但VideoPlayer不显示,"
echo "请将浏览器Console的日志复制给我。"
echo ""
