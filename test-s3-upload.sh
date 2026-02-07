#!/bin/bash
#
# S3 存储功能测试脚本
#

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}  OpenWan S3 存储功能测试${NC}"
echo -e "${GREEN}========================================${NC}"
echo ""

# 1. 检查服务
echo -e "${YELLOW}[1/5] 检查服务状态...${NC}"
if curl -s http://localhost:8080/api/v1/ping | grep -q "pong"; then
    echo -e "  ✓ 后端API服务正常"
else
    echo -e "  ${RED}✗ 后端API服务异常${NC}"
    exit 1
fi

# 2. 检查S3配置
echo -e "\n${YELLOW}[2/5] 验证S3配置...${NC}"
S3_BUCKET=$(grep s3_bucket /home/ec2-user/openwan/configs/config.yaml | awk '{print $2}' | tr -d '"')
S3_REGION=$(grep s3_region /home/ec2-user/openwan/configs/config.yaml | awk '{print $2}')
echo -e "  S3存储桶: $S3_BUCKET"
echo -e "  AWS区域: $S3_REGION"

if aws s3 ls "s3://$S3_BUCKET/" --region "$S3_REGION" > /dev/null 2>&1; then
    echo -e "  ✓ S3存储桶访问正常"
else
    echo -e "  ${RED}✗ S3存储桶访问失败${NC}"
    exit 1
fi

# 3. 记录上传前的S3文件数量
echo -e "\n${YELLOW}[3/5] 检查S3当前状态...${NC}"
BEFORE_COUNT=$(aws s3 ls "s3://$S3_BUCKET/openwan/" --recursive 2>/dev/null | wc -l || echo "0")
echo -e "  当前文件数: $BEFORE_COUNT"

# 4. 测试通过前端上传
echo -e "\n${YELLOW}[4/5] 准备测试文件...${NC}"
TEST_FILE="/tmp/openwan-s3-test-$(date +%s).txt"
echo "This is a test file for OpenWan S3 storage integration" > "$TEST_FILE"
echo "Uploaded at: $(date)" >> "$TEST_FILE"
echo "Test ID: $(uuidgen)" >> "$TEST_FILE"
echo -e "  ✓ 测试文件已创建: $TEST_FILE"

# 5. 显示上传指南
echo -e "\n${YELLOW}[5/5] 手动测试步骤:${NC}"
echo -e ""
echo -e "  ${GREEN}请按以下步骤测试S3上传：${NC}"
echo -e ""
echo -e "  1️⃣  在浏览器中打开: ${GREEN}http://localhost${NC}"
echo -e "  2️⃣  使用以下凭证登录:"
echo -e "     用户名: ${GREEN}admin${NC}"
echo -e "     密码:   ${GREEN}admin123${NC}"
echo -e "  3️⃣  点击左侧菜单 \"文件管理\" → \"文件上传\""
echo -e "  4️⃣  拖放测试文件或点击选择: ${GREEN}$TEST_FILE${NC}"
echo -e "  5️⃣  填写必填信息:"
echo -e "     - 选择分类"
echo -e "     - 文件类型: 富媒体 (文本文件)"
echo -e "     - 标题: S3 Upload Test"
echo -e "  6️⃣  点击 \"开始上传\""
echo -e "  7️⃣  等待上传完成，观察进度条"
echo -e ""

echo -e "${GREEN}验证上传结果:${NC}"
echo -e ""
echo -e "在上传完成后，运行以下命令验证文件已上传到S3："
echo -e ""
echo -e "  ${YELLOW}# 列出最近上传的文件${NC}"
echo -e "  aws s3 ls s3://$S3_BUCKET/openwan/ --recursive --human-readable | tail -10"
echo -e ""
echo -e "  ${YELLOW}# 查看文件总数${NC}"
echo -e "  aws s3 ls s3://$S3_BUCKET/openwan/ --recursive | wc -l"
echo -e ""
echo -e "  ${YELLOW}# 下载文件验证内容${NC}"
echo -e "  aws s3 cp s3://$S3_BUCKET/openwan/data1/{file_path} /tmp/downloaded.txt"
echo -e "  cat /tmp/downloaded.txt"
echo -e ""

echo -e "${GREEN}查看上传日志:${NC}"
echo -e "  tail -f /home/ec2-user/openwan/logs/api.log | grep -i upload"
echo -e ""

echo -e "${YELLOW}========================================${NC}"
echo -e "${YELLOW}  测试准备完成！${NC}"
echo -e "${YELLOW}========================================${NC}"
echo ""
echo -e "测试文件: $TEST_FILE"
echo -e "文件大小: $(ls -lh $TEST_FILE | awk '{print $5}')"
echo -e ""
echo -e "按 ${GREEN}Enter${NC} 键打开浏览器，或手动访问 ${GREEN}http://localhost${NC}"
read -p ""

# 尝试打开浏览器
if command -v xdg-open > /dev/null; then
    xdg-open http://localhost 2>/dev/null || true
elif command -v open > /dev/null; then
    open http://localhost 2>/dev/null || true
else
    echo "请手动在浏览器中打开: http://localhost"
fi

echo ""
echo -e "${GREEN}✓ 浏览器应已打开，请按照上述步骤测试上传${NC}"
echo ""
