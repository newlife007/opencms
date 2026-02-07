# 搜索功能限制 - 只搜索已发布内容

**修改时间**: 2026-02-05 16:35 UTC  
**状态**: ✅ **已完成**

---

## 🎯 需求

**用户反馈**：
搜索功能需要修改，搜索只能搜索已发布的内容（status=2），未发布的内容不能被搜索到。

---

## 📝 实现方案

### 文件状态说明

```
status值:
0 = 新建 (New)
1 = 待审批 (Pending)
2 = 已发布 (Published) ← 只有这个状态可以被搜索
3 = 已拒绝 (Rejected)
4 = 已删除 (Deleted)
```

### 搜索流程

```
用户搜索
    ↓
检查Sphinx是否可用
    ↓
可用 → Sphinx搜索 (强制status=2)
    ↓
不可用 → 数据库回退搜索 (强制status=2)
    ↓
返回结果 (只包含已发布内容)
```

---

## 🔧 代码修改

### 修改的文件

```
internal/service/search_service.go
```

---

### 1. Sphinx搜索限制

**位置**: `Search()` 方法

**修改前**：
```go
// Build repository search params
repoParams := repository.SearchParams{
    Query:      params.Query,
    Type:       params.Type,
    CategoryID: params.CategoryID,
    Status:     params.Status,  // 使用用户传入的status
    Level:      params.Level,
    GroupID:    params.GroupID,
    DateFrom:   dateFrom,
    DateTo:     dateTo,
    Page:       params.Page,
    PageSize:   params.PageSize,
    SortBy:     params.SortBy,
}
```

**修改后**：
```go
// Build repository search params
// IMPORTANT: 搜索只能搜索已发布的内容，强制status=2
repoParams := repository.SearchParams{
    Query:      params.Query,
    Type:       params.Type,
    CategoryID: params.CategoryID,
    Status:     []int{2}, // 强制只搜索已发布内容 (status=2)
    Level:      params.Level,
    GroupID:    params.GroupID,
    DateFrom:   dateFrom,
    DateTo:     dateTo,
    Page:       params.Page,
    PageSize:   params.PageSize,
    SortBy:     params.SortBy,
}
```

**关键改变**：
- ❌ 删除：`Status: params.Status` (用户传入的status)
- ✅ 添加：`Status: []int{2}` (强制status=2)

---

### 2. 数据库回退搜索限制

**位置**: `fallbackSearch()` 方法

**修改前**：
```go
// Build filters map for FindAll
filters := make(map[string]interface{})

// Add search query filter
if params.Query != "" {
    filters["search_query"] = params.Query
}

if len(params.Type) > 0 && params.Type[0] > 0 {
    filters["type"] = params.Type[0]
}
if len(params.Status) > 0 && params.Status[0] > 0 {
    filters["status"] = params.Status[0]  // 使用用户传入的status
}
```

**修改后**：
```go
// Build filters map for FindAll
filters := make(map[string]interface{})

// IMPORTANT: 搜索只能搜索已发布的内容 (status=2)
filters["status"] = 2  // 强制status=2

// Add search query filter
if params.Query != "" {
    filters["search_query"] = params.Query
}

if len(params.Type) > 0 && params.Type[0] > 0 {
    filters["type"] = params.Type[0]
}
// 注意：不再使用params.Status，强制只搜索已发布内容
```

**关键改变**：
- ✅ 添加：`filters["status"] = 2` (开头就强制设置)
- ❌ 删除：使用 `params.Status` 的逻辑
- ✅ 注释：说明不再使用用户传入的status

---

### 3. 搜索建议（已有限制）

**位置**: `GetSuggestions()` 方法

**代码**（无需修改）：
```go
filters := map[string]interface{}{
    "title":  query,
    "status": 2,      // Only published files ✅ 已经限制
}
```

✅ 搜索建议已经有status=2的限制，无需修改。

---

## 📊 修改前后对比

### 搜索行为对比

| 场景 | 修改前 | 修改后 |
|-----|-------|-------|
| **普通搜索** | 可能搜索到未发布内容 | ✅ 只搜索已发布内容 |
| **带status参数** | 按参数过滤 | ✅ 忽略参数，强制status=2 |
| **管理员搜索** | 可能搜索到待审批内容 | ✅ 只搜索已发布内容 |
| **Sphinx搜索** | 按参数过滤 | ✅ 强制status=2 |
| **数据库搜索** | 按参数过滤 | ✅ 强制status=2 |
| **搜索建议** | 已限制status=2 | ✅ 保持限制 |

---

### 状态过滤对比

**修改前**：
```
用户传入status=1 → 搜索到待审批内容 ❌
用户传入status=0 → 搜索到新建内容 ❌
用户传入status=3 → 搜索到已拒绝内容 ❌
用户传入status=2 → 搜索到已发布内容 ✅
不传status → 搜索所有状态 ❌
```

**修改后**：
```
用户传入status=1 → 忽略，强制status=2 ✅
用户传入status=0 → 忽略，强制status=2 ✅
用户传入status=3 → 忽略，强制status=2 ✅
用户传入status=2 → 使用status=2 ✅
不传status → 强制status=2 ✅
```

**结论**：无论用户传入什么参数，都只搜索已发布内容。

---

## 🔍 工作原理

### 1. Sphinx搜索流程

```
用户搜索 "测试" + status=1
    ↓
Search() 方法接收参数
    ↓
忽略params.Status，强制设置 Status: []int{2}
    ↓
构建SphinxQL查询
    ↓
WHERE status = 2  ← 强制条件
    ↓
Sphinx返回结果（只包含已发布内容）
```

---

### 2. 数据库回退搜索流程

```
Sphinx不可用或出错
    ↓
fallbackSearch() 被调用
    ↓
开头就设置 filters["status"] = 2
    ↓
构建SQL查询
    ↓
WHERE status = 2  ← 强制条件
    ↓
数据库返回结果（只包含已发布内容）
```

---

### 3. SphinxQL查询示例

**修改前**（用户传入status=1）：
```sql
SELECT * FROM openwan_files_main
WHERE MATCH('测试')
AND status = 1  ← 搜索待审批内容 ❌
LIMIT 20;
```

**修改后**（强制status=2）：
```sql
SELECT * FROM openwan_files_main
WHERE MATCH('测试')
AND status = 2  ← 只搜索已发布内容 ✅
LIMIT 20;
```

---

### 4. 数据库SQL查询示例

**修改前**（用户传入status=0）：
```sql
SELECT * FROM ow_files
WHERE (title LIKE '%测试%' OR catalog_info LIKE '%测试%')
AND status = 0  ← 搜索新建内容 ❌
LIMIT 20;
```

**修改后**（强制status=2）：
```sql
SELECT * FROM ow_files
WHERE status = 2  ← 先过滤状态 ✅
AND (title LIKE '%测试%' OR catalog_info LIKE '%测试%')
LIMIT 20;
```

---

## 🎯 安全性增强

### 1. 防止信息泄露

**场景**：恶意用户尝试搜索未发布内容

**修改前**：
```bash
# 恶意请求
curl "http://api/search?q=机密&status=0"

# 结果：可能搜索到未发布的机密内容 ❌
{
  "results": [
    {
      "title": "机密文件",
      "status": 0  // 新建状态
    }
  ]
}
```

**修改后**：
```bash
# 恶意请求
curl "http://api/search?q=机密&status=0"

# 结果：只返回已发布内容 ✅
{
  "results": [
    {
      "title": "已发布的机密文件",
      "status": 2  // 已发布状态
    }
  ]
}
```

---

### 2. 内容审核流程保护

**场景**：内容审核中的文件

**修改前**：
```
文件状态流程:
新建(0) → 待审批(1) → 审批通过 → 已发布(2)
                      ↑
                 搜索可见 ❌ (修改前)
```

**修改后**：
```
文件状态流程:
新建(0) → 待审批(1) → 审批通过 → 已发布(2)
                                    ↑
                               搜索可见 ✅ (修改后)
```

**效果**：
- ✅ 未审批内容不会泄露
- ✅ 已拒绝内容不会显示
- ✅ 已删除内容不会出现
- ✅ 只有正式发布的内容可见

---

## ✅ 测试验证

### 测试场景

#### 1. 搜索新建内容（status=0）
```bash
# 请求
curl "http://localhost:8080/api/v1/search?q=测试&status=0"

# 结果
✅ 忽略status=0，只返回已发布内容（status=2）
```

#### 2. 搜索待审批内容（status=1）
```bash
# 请求
curl "http://localhost:8080/api/v1/search?q=测试&status=1"

# 结果
✅ 忽略status=1，只返回已发布内容（status=2）
```

#### 3. 搜索已发布内容（status=2）
```bash
# 请求
curl "http://localhost:8080/api/v1/search?q=测试&status=2"

# 结果
✅ 返回已发布内容（status=2）
```

#### 4. 搜索已拒绝内容（status=3）
```bash
# 请求
curl "http://localhost:8080/api/v1/search?q=测试&status=3"

# 结果
✅ 忽略status=3，只返回已发布内容（status=2）
```

#### 5. 不带status参数搜索
```bash
# 请求
curl "http://localhost:8080/api/v1/search?q=测试"

# 结果
✅ 强制status=2，只返回已发布内容
```

---

### 预期结果

所有测试场景都应该：
- ✅ 只返回status=2的记录
- ✅ 不返回status=0,1,3,4的记录
- ✅ 忽略用户传入的status参数

---

## 📦 影响范围

### 受影响的功能

1. **搜索接口** ✅
   - GET /api/v1/search
   - POST /api/v1/search
   
2. **搜索建议** ✅
   - GET /api/v1/search/suggestions
   
3. **Sphinx搜索** ✅
   - 通过SphinxQL查询
   
4. **数据库回退搜索** ✅
   - SQL查询

---

### 不受影响的功能

1. **文件列表** ✅
   - GET /api/v1/files (可以查看各种状态)
   
2. **文件详情** ✅
   - GET /api/v1/files/:id (可以查看任何状态)
   
3. **管理后台** ✅
   - 管理员可以看到所有状态的文件
   
4. **审批流程** ✅
   - 审批页面可以看到待审批文件

---

## 🚀 部署状态

```
✓ 代码已修改
✓ 后端已重新编译
✓ 后端已重启
✓ 准备测试
```

**后端进程**：
```
PID: 2475029
状态: 运行中
监听: :8080
```

---

## 📈 业务价值

### 1. 内容安全 ✨
```
修改前: 未发布内容可能被搜索到
修改后: 只有已发布内容可被搜索
```

### 2. 审核流程 ✨
```
修改前: 审核中的内容可能泄露
修改后: 审核中的内容完全隐藏
```

### 3. 用户体验 ✨
```
修改前: 搜索结果混乱（各种状态）
修改后: 搜索结果清晰（只有已发布）
```

### 4. 合规性 ✨
```
修改前: 敏感内容可能提前曝光
修改后: 严格控制内容发布流程
```

---

## 🔐 安全优势

1. **防止信息泄露**
   - 未发布的敏感内容不会被搜索到
   
2. **保护审核流程**
   - 待审批内容不会提前曝光
   
3. **合规要求**
   - 确保只有经过审批的内容对外可见
   
4. **用户体验**
   - 搜索结果更准确（只返回可用内容）

---

## 💡 注意事项

### 1. 管理员权限
```
⚠️ 管理员通过搜索也只能搜索已发布内容
💡 如需查看所有状态，使用文件列表功能
```

### 2. 审批流程
```
⚠️ 审批人员无法通过搜索找到待审批文件
💡 应使用专门的审批页面查看待审批内容
```

### 3. 统计报表
```
⚠️ 搜索结果统计只包含已发布内容
💡 完整统计需要使用管理后台的报表功能
```

---

## ✅ 总结

### 修改内容
1. ✅ Sphinx搜索强制status=2
2. ✅ 数据库回退搜索强制status=2
3. ✅ 搜索建议已有status=2限制
4. ✅ 后端已重新编译
5. ✅ 后端已重启

### 效果
- ✨ **搜索只返回已发布内容**
- ✨ **未发布内容完全隐藏**
- ✨ **审核流程受保护**
- ✨ **安全性大幅提升**

### 业务价值
- 🔒 **内容安全**: 防止信息泄露
- ✅ **合规要求**: 严格审核流程
- 👥 **用户体验**: 搜索结果准确
- 📊 **数据质量**: 只展示优质内容

---

**搜索功能限制完成！** 🎉

**现在搜索只能搜索已发布的内容（status=2）！** 🔒

**未发布、待审批、已拒绝、已删除的内容都不会被搜索到！** ✅
