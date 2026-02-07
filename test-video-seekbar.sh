#!/bin/bash
#
# 视频播放器拖拽测试脚本
#

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  视频播放器进度条拖拽测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

echo -e "${YELLOW}修复内容:${NC}"
echo "  ✓ 启用Video.js SeekBar交互"
echo "  ✓ 添加完整的控制栏配置"
echo "  ✓ 增强CSS样式确保可拖拽"
echo "  ✓ 添加键盘快捷键支持"
echo "  ✓ 优化移动端触摸体验"
echo ""

echo -e "${YELLOW}测试步骤:${NC}"
echo ""
echo -e "${BLUE}1. 清除浏览器缓存${NC}"
echo "   - Chrome/Edge: 按 Ctrl+Shift+R"
echo "   - Firefox: 按 Ctrl+Shift+R"
echo "   - Mac: 按 Cmd+Shift+R"
echo ""

echo -e "${BLUE}2. 访问应用${NC}"
echo "   URL: ${GREEN}http://localhost${NC}"
echo "   用户名: ${GREEN}admin${NC}"
echo "   密码: ${GREEN}admin123${NC}"
echo ""

echo -e "${BLUE}3. 测试进度条功能${NC}"
echo ""
echo "   ${YELLOW}测试A: 点击跳转${NC}"
echo "   - 在进度条任意位置点击"
echo "   - 视频应立即跳转到该位置"
echo "   - 检查播放位置是否正确"
echo ""

echo "   ${YELLOW}测试B: 拖拽进度${NC}"
echo "   - 按住进度条上的进度球"
echo "   - 左右拖动"
echo "   - 观察时间提示是否显示"
echo "   - 释放后视频从新位置播放"
echo ""

echo "   ${YELLOW}测试C: 悬停效果${NC}"
echo "   - 鼠标悬停在进度条上"
echo "   - 进度条应该变粗"
echo "   - 显示当前位置的时间"
echo ""

echo "   ${YELLOW}测试D: 键盘控制${NC}"
echo "   - 左箭头: 后退5秒"
echo "   - 右箭头: 前进5秒"
echo "   - 空格键: 播放/暂停"
echo "   - 上下箭头: 调节音量"
echo ""

echo -e "${BLUE}4. 检查浏览器控制台${NC}"
echo "   - 打开开发者工具 (F12)"
echo "   - 切换到 Console 标签"
echo "   - 查找以下日志:"
echo "     • 'Video player ready'"
echo "     • 'SeekBar enabled for interaction'"
echo "     • 'Video metadata loaded, duration: X'"
echo "     • 'Seeking to: X' (拖拽时)"
echo "     • 'Seeked to: X' (拖拽完成)"
echo ""

echo -e "${YELLOW}故障排查:${NC}"
echo ""
echo "  ${BLUE}问题: 进度条还是不能拖拽${NC}"
echo "  解决: 1. 确保已硬刷新 (Ctrl+Shift+R)"
echo "       2. 检查Console是否有错误"
echo "       3. 验证Video.js已加载: typeof videojs"
echo ""

echo "  ${BLUE}问题: 点击没有反应${NC}"
echo "  解决: 在Console执行:"
echo "       const player = document.querySelector('.video-js').player"
echo "       player.controlBar.progressControl.seekBar.enable()"
echo ""

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  准备测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

echo -e "前端已更新到最新版本:"
ls -lh /home/ec2-user/openwan/frontend/dist/index.html | awk '{print "  构建时间: " $6, $7, $8}'
echo ""

echo -e "按 ${GREEN}Enter${NC} 继续..."
read -p ""

# 检查是否有浏览器可用
if command -v xdg-open > /dev/null; then
    echo "正在打开浏览器..."
    xdg-open http://localhost 2>/dev/null || true
elif command -v open > /dev/null; then
    echo "正在打开浏览器..."
    open http://localhost 2>/dev/null || true
else
    echo "请手动在浏览器中打开: http://localhost"
fi

echo ""
echo -e "${GREEN}✓ 测试准备完成${NC}"
echo ""
echo "提示: 如果进度条功能正常，你应该能够:"
echo "  • 点击进度条任意位置跳转"
echo "  • 拖拽进度球调整播放位置"
echo "  • 看到悬停时的时间提示"
echo "  • 使用键盘箭头键快进/快退"
echo ""
echo -e "${YELLOW}完成测试后，按 Ctrl+C 退出${NC}"
echo ""

# 实时显示API日志中的seek相关信息
echo "监控播放器事件（按Ctrl+C停止）:"
echo "----------------------------------------"
tail -f /home/ec2-user/openwan/logs/api.log 2>/dev/null | grep -i "seek\|video\|player" --line-buffered || {
    echo "无法读取日志文件，请手动检查浏览器控制台"
    sleep infinity
}
