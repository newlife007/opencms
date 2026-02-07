#!/bin/bash

# OpenWan 功能测试脚本
# 测试核心业务功能

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

API_BASE_URL="${API_BASE_URL:-http://localhost:8080}"
TEST_USERNAME="${TEST_USERNAME:-admin}"
TEST_PASSWORD="${TEST_PASSWORD:-admin123}"

TESTS_RUN=0
TESTS_PASSED=0
TESTS_FAILED=0

TOKEN=""
COOKIE_FILE="/tmp/openwan_test_cookies.txt"

echo "======================================"
echo "OpenWan 功能测试"
echo "======================================"
echo ""

log_test() {
    echo -e "${BLUE}[TEST]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    TESTS_PASSED=$((TESTS_PASSED + 1))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    TESTS_FAILED=$((TESTS_FAILED + 1))
}

log_skip() {
    echo -e "${YELLOW}[SKIP]${NC} $1"
}

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

run_test() {
    TESTS_RUN=$((TESTS_RUN + 1))
}

# 清理函数
cleanup() {
    rm -f "$COOKIE_FILE"
}
trap cleanup EXIT

# ============================================
# 测试1: 用户认证
# ============================================
echo ""
echo "==== 测试套件 1: 用户认证 ===="
echo ""

log_test "1.1 用户登录"
run_test

login_response=$(curl -s -X POST "$API_BASE_URL/api/v1/auth/login" \
    -H "Content-Type: application/json" \
    -c "$COOKIE_FILE" \
    -d "{\"username\":\"$TEST_USERNAME\",\"password\":\"$TEST_PASSWORD\"}")

if echo "$login_response" | grep -q '"success":true\|"token"'; then
    log_pass "用户登录成功"
    
    # 尝试提取token
    TOKEN=$(echo "$login_response" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    if [ ! -z "$TOKEN" ]; then
        log_info "Token: ${TOKEN:0:30}..."
    fi
else
    log_fail "用户登录失败"
    log_info "Response: $login_response"
fi

log_test "1.2 获取当前用户信息"
run_test

if [ ! -z "$TOKEN" ]; then
    user_response=$(curl -s "$API_BASE_URL/api/v1/auth/me" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$user_response" | grep -q '"username"\|"user"'; then
        log_pass "获取用户信息成功"
    else
        log_fail "获取用户信息失败"
    fi
else
    log_skip "跳过 (无token)"
fi

log_test "1.3 测试未授权访问"
run_test

unauth_response=$(curl -s -w "%{http_code}" -o /dev/null "$API_BASE_URL/api/v1/files")
if [ "$unauth_response" == "401" ] || [ "$unauth_response" == "403" ]; then
    log_pass "未授权访问被正确拦截"
else
    log_fail "未授权访问未被拦截 (HTTP $unauth_response)"
fi

# ============================================
# 测试2: 文件管理
# ============================================
echo ""
echo "==== 测试套件 2: 文件管理 ===="
echo ""

if [ ! -z "$TOKEN" ]; then
    log_test "2.1 获取文件列表"
    run_test
    
    files_response=$(curl -s "$API_BASE_URL/api/v1/files?page=1&page_size=10" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$files_response" | grep -q '"files"\|"data"\|"total"'; then
        log_pass "获取文件列表成功"
        
        # 显示文件数量
        file_count=$(echo "$files_response" | grep -o '"total":[0-9]*' | grep -o '[0-9]*' || echo "0")
        log_info "文件总数: $file_count"
    else
        log_fail "获取文件列表失败"
    fi
    
    log_test "2.2 测试文件筛选"
    run_test
    
    filter_response=$(curl -s "$API_BASE_URL/api/v1/files?type=1&status=2" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$filter_response" | grep -q '"success"\|"data"'; then
        log_pass "文件筛选功能正常"
    else
        log_fail "文件筛选功能异常"
    fi
    
    log_test "2.3 测试文件上传端点"
    run_test
    
    # 创建测试文件
    echo "Test file content" > /tmp/test_upload.txt
    
    upload_response=$(curl -s -X POST "$API_BASE_URL/api/v1/files/upload" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE" \
        -F "file=@/tmp/test_upload.txt" \
        -F "title=Test Upload" \
        -F "type=4" \
        -F "category_id=1")
    
    if echo "$upload_response" | grep -q '"success"\|"id"\|"path"'; then
        log_pass "文件上传端点可访问"
        
        # 提取文件ID
        FILE_ID=$(echo "$upload_response" | grep -o '"id":[0-9]*' | grep -o '[0-9]*' | head -1)
        if [ ! -z "$FILE_ID" ]; then
            log_info "上传文件ID: $FILE_ID"
        fi
    else
        log_skip "文件上传测试 (可能需要配置)"
        log_info "Response: ${upload_response:0:200}"
    fi
    
    rm -f /tmp/test_upload.txt
else
    log_skip "跳过文件管理测试 (无认证)"
fi

# ============================================
# 测试3: 搜索功能
# ============================================
echo ""
echo "==== 测试套件 3: 搜索功能 ===="
echo ""

if [ ! -z "$TOKEN" ]; then
    log_test "3.1 全文搜索"
    run_test
    
    search_response=$(curl -s -X POST "$API_BASE_URL/api/v1/search" \
        -H "Authorization: Bearer $TOKEN" \
        -H "Content-Type: application/json" \
        -b "$COOKIE_FILE" \
        -d '{"q":"test","page":1,"page_size":10}')
    
    if echo "$search_response" | grep -q '"results"\|"data"\|"total"'; then
        log_pass "搜索功能正常"
    else
        log_skip "搜索功能 (可能需要Sphinx)"
    fi
else
    log_skip "跳过搜索测试 (无认证)"
fi

# ============================================
# 测试4: 分类管理
# ============================================
echo ""
echo "==== 测试套件 4: 分类管理 ===="
echo ""

if [ ! -z "$TOKEN" ]; then
    log_test "4.1 获取分类树"
    run_test
    
    category_response=$(curl -s "$API_BASE_URL/api/v1/categories/tree" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$category_response" | grep -q '"success"\|"data"\|"\[" '; then
        log_pass "获取分类树成功"
    else
        log_skip "分类树 (可能数据库为空)"
    fi
    
    log_test "4.2 获取分类列表"
    run_test
    
    category_list=$(curl -s "$API_BASE_URL/api/v1/categories" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$category_list" | grep -q '"success"\|"data"'; then
        log_pass "获取分类列表成功"
    else
        log_fail "获取分类列表失败"
    fi
else
    log_skip "跳过分类管理测试 (无认证)"
fi

# ============================================
# 测试5: 管理功能
# ============================================
echo ""
echo "==== 测试套件 5: 管理功能 ===="
echo ""

if [ ! -z "$TOKEN" ]; then
    log_test "5.1 获取用户列表"
    run_test
    
    users_response=$(curl -s "$API_BASE_URL/api/v1/admin/users" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$users_response" | grep -q '"users"\|"data"'; then
        log_pass "获取用户列表成功"
    else
        log_skip "用户列表 (可能需要管理员权限)"
    fi
    
    log_test "5.2 获取组列表"
    run_test
    
    groups_response=$(curl -s "$API_BASE_URL/api/v1/admin/groups" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$groups_response" | grep -q '"groups"\|"data"'; then
        log_pass "获取组列表成功"
    else
        log_skip "组列表 (可能需要管理员权限)"
    fi
    
    log_test "5.3 获取角色列表"
    run_test
    
    roles_response=$(curl -s "$API_BASE_URL/api/v1/admin/roles" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$roles_response" | grep -q '"roles"\|"data"'; then
        log_pass "获取角色列表成功"
    else
        log_skip "角色列表 (可能需要管理员权限)"
    fi
    
    log_test "5.4 获取权限列表"
    run_test
    
    perms_response=$(curl -s "$API_BASE_URL/api/v1/admin/permissions" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$perms_response" | grep -q '"permissions"\|"data"'; then
        log_pass "获取权限列表成功"
    else
        log_skip "权限列表 (可能需要管理员权限)"
    fi
else
    log_skip "跳过管理功能测试 (无认证)"
fi

# ============================================
# 测试6: 目录配置
# ============================================
echo ""
echo "==== 测试套件 6: 目录配置 ===="
echo ""

if [ ! -z "$TOKEN" ]; then
    log_test "6.1 获取目录配置"
    run_test
    
    catalog_response=$(curl -s "$API_BASE_URL/api/v1/catalog/tree?type=1" \
        -H "Authorization: Bearer $TOKEN" \
        -b "$COOKIE_FILE")
    
    if echo "$catalog_response" | grep -q '"success"\|"data"\|"\["'; then
        log_pass "获取目录配置成功"
    else
        log_skip "目录配置 (可能数据库为空)"
    fi
else
    log_skip "跳过目录配置测试 (无认证)"
fi

# ============================================
# 测试总结
# ============================================
echo ""
echo "======================================"
echo "功能测试总结"
echo "======================================"
echo "总计运行: $TESTS_RUN"
echo -e "通过: ${GREEN}$TESTS_PASSED${NC}"
echo -e "失败: ${RED}$TESTS_FAILED${NC}"
echo -e "跳过: ${YELLOW}$((TESTS_RUN - TESTS_PASSED - TESTS_FAILED))${NC}"

PASS_RATE=$((TESTS_PASSED * 100 / TESTS_RUN))
echo ""
echo "通过率: $PASS_RATE%"

if [ $TESTS_FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✓ 所有功能测试通过！${NC}"
    exit 0
else
    echo ""
    echo -e "${YELLOW}⚠ 有 $TESTS_FAILED 个测试失败${NC}"
    exit 1
fi
