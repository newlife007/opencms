#!/bin/bash

echo "================================================================"
echo "  紧急修复：清除Vite缓存并重启服务"
echo "================================================================"
echo ""
echo "问题: 即使修复了VideoPlayer.vue，仍然报getTech错误"
echo "原因: Vite开发服务器使用了旧的缓存模块"
echo ""
echo "✅ 已完成的操作:"
echo "   1. 清除 node_modules/.vite 缓存"
echo "   2. 重启 Vite 开发服务器"
echo "   3. 重新构建前端（清除dist和缓存）"
echo ""
echo "================================================================"
echo "  现在请测试"
echo "================================================================"
echo ""
echo "重要：必须完全刷新浏览器！"
echo ""
echo "方法1: 强制刷新（推荐）"
echo "-------"
echo "按 Ctrl+Shift+Delete 打开清除浏览器数据"
echo "选择："
echo "  ☑ 缓存的图片和文件"
echo "  ☑ Cookie和其他网站数据（可选）"
echo "时间范围: 过去1小时"
echo "然后点击 '清除数据'"
echo ""
echo "方法2: 硬刷新"
echo "-------"
echo "按住 Ctrl+Shift+R (Windows/Linux)"
echo "或 Cmd+Shift+R (Mac)"
echo ""
echo "方法3: 无痕模式"
echo "-------"
echo "打开新的无痕/隐私窗口"
echo "Ctrl+Shift+N (Chrome) 或 Ctrl+Shift+P (Firefox)"
echo ""
echo "================================================================"
echo "  测试步骤"
echo "================================================================"
echo ""
echo "1. 清除浏览器缓存（见上方）"
echo ""
echo "2. 访问文件列表"
echo "   http://localhost:3000/files"
echo ""

PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4 2>/dev/null)
if [ -n "$PUBLIC_IP" ]; then
    echo "   或: http://$PUBLIC_IP:3000/files"
    echo ""
fi

echo "3. 按 F12 打开开发者工具 → Console"
echo ""
echo "4. 点击任意文件的'详情'按钮"
echo ""
echo "================================================================"
echo "  期望结果"
echo "================================================================"
echo ""
echo "✅ 控制台应该显示:"
echo "   [FileList] viewDetail clicked, id: XX"
echo "   [FileList] Navigating to: /files/XX"
echo "   Computing previewUrl: ..."
echo "   Preview URL generated: ..."
echo "   Initializing video player with src: ..."
echo "   Video.js player is ready"
echo "   Tech in use: html5 (或 flvjs)"
echo ""
echo "❌ 不应该有的错误:"
echo "   ✗ getTech is not a function"
echo "   ✗ 任何红色的TypeError"
echo ""
echo "================================================================"
echo "  验证服务状态"
echo "================================================================"
echo ""

# 检查Vite进程
if ps aux | grep -v grep | grep -q "vite"; then
    echo "✓ Vite开发服务器运行中"
    ViteProcessTime=$(ps aux | grep vite | grep -v grep | awk '{print $9}')
    echo "  启动时间: $ViteProcessTime"
else
    echo "✗ Vite服务器未运行"
    echo "  手动启动: cd /home/ec2-user/openwan/frontend && npm run dev"
fi

echo ""

# 检查日志
if [ -f /tmp/vite-restart.log ]; then
    ERRORS=$(grep -i "error" /tmp/vite-restart.log | wc -l)
    if [ "$ERRORS" -eq 0 ]; then
        echo "✓ Vite启动无错误"
    else
        echo "⚠ Vite日志中有 $ERRORS 个错误"
        echo "  查看: tail -20 /tmp/vite-restart.log"
    fi
fi

echo ""

# 测试前端访问
if curl -s http://localhost:3000 > /dev/null; then
    echo "✓ 前端服务可访问 (http://localhost:3000)"
else
    echo "✗ 前端服务不可访问"
fi

echo ""

# 测试后端访问
if curl -s http://localhost:8080/health > /dev/null; then
    echo "✓ 后端服务可访问 (http://localhost:8080)"
else
    echo "✗ 后端服务不可访问"
fi

echo ""
echo "================================================================"
echo "  如果问题依然存在"
echo "================================================================"
echo ""
echo "如果清除缓存后仍然报 getTech 错误:"
echo ""
echo "1. 完全关闭浏览器"
echo "   - 不是关闭标签页，是退出整个浏览器程序"
echo "   - 等待5秒"
echo "   - 重新打开浏览器"
echo ""
echo "2. 尝试不同的浏览器"
echo "   - Chrome"
echo "   - Firefox"
echo "   - Edge"
echo ""
echo "3. 检查Network标签页"
echo "   - 按F12 → Network标签"
echo "   - 刷新页面"
echo "   - 找到 'videojs-plugins-*.js' 文件"
echo "   - 查看文件名中的hash是否变化"
echo "   - 应该不再是 'videojs-plugins-a3625071.js'"
echo ""
echo "4. 提供完整错误信息"
echo "   - 错误的完整堆栈"
echo "   - videojs-plugins文件的完整名称"
echo "   - 浏览器类型和版本"
echo ""
echo "================================================================"
echo "  检查VideoPlayer.vue修复"
echo "================================================================"
echo ""

if grep -q "this.tech({ IWillNotUseThisInPlugins: true })" /home/ec2-user/openwan/frontend/src/components/VideoPlayer.vue; then
    echo "✗ 警告: VideoPlayer.vue仍有旧代码！"
    echo "  需要重新修复"
else
    echo "✓ VideoPlayer.vue已正确修复"
    if grep -q "this.techName_" /home/ec2-user/openwan/frontend/src/components/VideoPlayer.vue; then
        echo "  使用 techName_ 属性（安全方式）"
    fi
fi

echo ""
echo "================================================================"
echo ""
echo "🎯 关键步骤: 清除浏览器缓存！"
echo ""
echo "   1. 按 Ctrl+Shift+Delete"
echo "   2. 清除'缓存的图片和文件'"
echo "   3. 访问 http://localhost:3000/files"
echo "   4. 点击详情按钮"
echo "   5. 查看控制台是否还有 getTech 错误"
echo ""
echo "   然后告诉我结果！"
echo ""
echo "================================================================"
