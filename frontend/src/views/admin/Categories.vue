<template>
  <div class="categories-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>
            <el-icon><FolderOpened /></el-icon>{{ t("admin.categories.title") }}</span>
          <div class="header-actions">
            <el-input
              v-model="filterText"
              placeholder="搜索分类..."
              clearable
              style="width: 250px; margin-right: 10px"
              prefix-icon="Search"
            />
            <el-button type="primary" icon="Plus" @click="handleAdd(null)">
              添加根分类
            </el-button>
          </div>
        </div>
      </template>

      <el-tree
        ref="treeRef"
        :data="categoryTree"
        :props="treeProps"
        :filter-node-method="filterNode"
        node-key="id"
        default-expand-all
        draggable
        :allow-drop="allowDrop"
        @node-drop="handleDrop"
        v-loading="loading"
        class="category-tree"
      >
        <template #default="{ node, data }">
          <div class="tree-node-content">
            <span class="node-info">
              <el-icon class="node-icon" :class="{ 'has-children': data.children && data.children.length > 0 }">
                <Folder v-if="data.children && data.children.length > 0" />
                <Document v-else />
              </el-icon>
              <span class="node-name">{{ node.label }}</span>
              <el-tag v-if="!data.enabled" size="small" type="info" style="margin-left: 8px">禁用</el-tag>
              <el-tag v-if="data.children && data.children.length > 0" size="small" type="" style="margin-left: 8px">
                {{ data.children.length }} 个子分类
              </el-tag>
            </span>
            <span class="node-actions">
              <el-button
                type="primary"
                size="small"
                icon="Plus"
                text
                @click.stop="handleAdd(data)"
                title="添加子分类"
              >
                添加子分类
              </el-button>
              <el-button
                type="success"
                size="small"
                icon="Edit"
                text
                @click.stop="handleEdit(data)"
                title="编辑"
              >{{ t("common.edit") }}</el-button>
              <el-button
                type="danger"
                size="small"
                icon="Delete"
                text
                @click.stop="handleDelete(data)"
                title="删除"
              >{{ t("common.delete") }}</el-button>
            </span>
          </div>
        </template>
      </el-tree>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog
      v-model="dialogVisible"
      :title="formTitle"
      width="600px"
      :close-on-click-modal="false"
      @close="handleDialogClose"
    >
      <el-form
        ref="categoryFormRef"
        :model="categoryForm"
        :rules="categoryRules"
        label-width="120px"
      >
        <el-form-item label="上级分类">
          <el-tree-select
            v-model="categoryForm.parent_id"
            :data="categoryTreeForSelect"
            :props="treeSelectProps"
            placeholder="不选择则为根分类"
            clearable
            check-strictly
            :render-after-expand="false"
            style="width: 100%"
          />
          <div class="form-tip" v-if="categoryForm.parent_id">
            <el-icon><InfoFilled /></el-icon>
            当前上级：<strong>{{ getParentCategoryName(categoryForm.parent_id) }}</strong>
          </div>
          <div class="form-tip" v-else>
            <el-icon><InfoFilled /></el-icon>
            未选择上级分类，将创建为<strong>{{ t("admin.categories.rootCategory") }}</strong>
          </div>
        </el-form-item>

        <el-form-item label="分类名称" prop="name">
          <el-input
            v-model="categoryForm.name"
            placeholder="请输入分类名称，如：视频资源"
            clearable
          />
        </el-form-item>

        <el-form-item label="分类描述" prop="description">
          <el-input
            v-model="categoryForm.description"
            type="textarea"
            :rows="3"
            placeholder="请输入分类描述，方便其他用户理解此分类的用途"
            clearable
            maxlength="200"
            show-word-limit
          />
        </el-form-item>

        <el-form-item label="排序权重" prop="weight">
          <el-input-number
            v-model="categoryForm.weight"
            :min="0"
            :max="9999"
            :step="1"
            controls-position="right"
            style="width: 100%"
          />
          <div class="form-tip">
            <el-icon><InfoFilled /></el-icon>
            数字越大排序越靠前，默认为 0
          </div>
        </el-form-item>

        <el-form-item :label="t('files.status')">
          <el-switch
            v-model="categoryForm.enabled"
            :active-value="true"
            :inactive-value="false"
            active-text="启用"
            inactive-text="禁用"
            inline-prompt
          />
          <div class="form-tip">
            <el-icon><InfoFilled /></el-icon>
            禁用后，该分类将不会在文件上传时显示
          </div>
        </el-form-item>
      </el-form>

      <template #footer>
        <span class="dialog-footer">
          <el-button @click="dialogVisible = false">{{ t("common.cancel") }}</el-button>
          <el-button type="primary" @click="handleSubmit" :loading="submitting">
            {{ isEdit ? '保存修改' : '立即创建' }}
          </el-button>
        </span>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, watch, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import categoryApi from '@/api/category'


const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const dialogVisible = ref(false)
const categoryTree = ref([])
const selectedCategory = ref(null)
const filterText = ref('')

const treeRef = ref(null)
const categoryFormRef = ref(null)

const treeProps = {
  label: 'name',
  children: 'children',
  value: 'id',
}

const treeSelectProps = {
  label: 'name',
  children: 'children',
  value: 'id',
  disabled: 'disabled',
}

const categoryForm = reactive({
  id: null,
  parent_id: null,
  name: '',
  description: '',
  weight: 0,
  enabled: true,
})

const categoryRules = {
  name: [
    { required: true, message: '请输入分类名称', trigger: 'blur' },
    { min: 2, max: 50, message: '分类名称长度在 2 到 50 个字符', trigger: 'blur' },
  ],
  weight: [
    { type: 'number', message: '排序权重必须是数字', trigger: 'blur' },
  ],
}

const formTitle = computed(() => {
  if (!isEdit.value) {
    if (categoryForm.parent_id) {
      return `添加子分类 - 上级：${getParentCategoryName(categoryForm.parent_id)}`
    }
    return '添加根分类'
  }
  return `编辑分类 - ${categoryForm.name || ''}`
})

// Build tree for select (exclude current editing category to prevent circular reference)
const categoryTreeForSelect = computed(() => {
  if (!isEdit.value) {
    return categoryTree.value
  }
  // When editing, exclude current category and its descendants
  return filterCategoryTree(categoryTree.value, categoryForm.id)
})

const filterCategoryTree = (tree, excludeId) => {
  return tree.filter(node => node.id !== excludeId).map(node => ({
    ...node,
    children: node.children ? filterCategoryTree(node.children, excludeId) : []
  }))
}

const filterNode = (value, data) => {
  if (!value) return true
  return data.name.includes(value)
}

const allowDrop = (draggingNode, dropNode, type) => {
  // Allow drop as inner child or before/after sibling
  return true
}

const formatDate = (timestamp) => {
  if (!timestamp) return '-'
  // Convert Unix timestamp to date
  return new Date(timestamp * 1000).toLocaleString('zh-CN')
}

// Get category name by ID (for display)
const getParentCategoryName = (parentId) => {
  if (!parentId) return '无'
  const findCategory = (tree, id) => {
    for (const node of tree) {
      if (node.id === id) {
        return node.name
      }
      if (node.children) {
        const found = findCategory(node.children, id)
        if (found) return found
      }
    }
    return null
  }
  return findCategory(categoryTree.value, parentId) || `分类 #${parentId}`
}

const loadCategoryTree = async () => {
  loading.value = true
  try {
    const res = await categoryApi.getTree()
    if (res.success) {
      categoryTree.value = res.data || []
    }
  } catch (error) {
    ElMessage.error('加载分类树失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = (parentCategory) => {
  isEdit.value = false
  selectedCategory.value = null
  Object.assign(categoryForm, {
    id: null,
    parent_id: parentCategory?.id || null,
    name: '',
    description: '',
    weight: 0,
    enabled: true,
  })
  dialogVisible.value = true
}

const handleEdit = async (category) => {
  isEdit.value = true
  selectedCategory.value = category
  
  // Load full category details
  try {
    const res = await categoryApi.getDetail(category.id)
    if (res.success) {
      const data = res.data
      Object.assign(categoryForm, {
        id: data.id,
        parent_id: data.parent_id || null,
        name: data.name,
        description: data.description || '',
        weight: data.weight,
        enabled: data.enabled,
      })
      dialogVisible.value = true
    }
  } catch (error) {
    ElMessage.error('加载分类详情失败')
  }
}

const handleDialogClose = () => {
  categoryFormRef.value?.resetFields()
  isEdit.value = false
  selectedCategory.value = null
}

const handleSubmit = async () => {
  if (!categoryFormRef.value) return

  await categoryFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      const data = { ...categoryForm }
      // Convert null parent_id to 0 for root category
      if (data.parent_id === null) {
        data.parent_id = 0
      }
      
      if (isEdit.value) {
        await categoryApi.update(data.id, data)
        ElMessage.success('更新成功')
      } else {
        await categoryApi.create(data)
        ElMessage.success('创建成功')
      }
      
      dialogVisible.value = false
      await loadCategoryTree()
      handleDialogClose()
    } catch (error) {
      const errMsg = error?.response?.data?.error || error?.message || '操作失败'
      ElMessage.error(isEdit.value ? `更新失败: ${errMsg}` : `创建失败: ${errMsg}`)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (category) => {
  if (category.children && category.children.length > 0) {
    ElMessage.warning('该分类下还有子分类，请先删除子分类')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除分类"${category.name}"吗？删除后无法恢复！`,
      '删除确认',
      {
        confirmButtonText: '确定删除',
        cancelButtonText: '取消',
        type: 'warning',
      }
    )
    
    await categoryApi.delete(category.id)
    ElMessage.success(t('message.deleteSuccess'))
    await loadCategoryTree()
    
    if (selectedCategory.value?.id === category.id) {
      handleDialogClose()
    }
  } catch (error) {
    if (error !== 'cancel') {
      const errMsg = error?.response?.data?.error || error?.message || '删除失败'
      ElMessage.error(`删除失败: ${errMsg}`)
    }
  }
}

const handleDrop = async (draggingNode, dropNode, dropType) => {
  try {
    // Calculate new parent_id based on drop position
    let newParentId = 0

    if (dropType === 'inner') {
      newParentId = dropNode.data.id
    } else {
      newParentId = dropNode.data.parent_id
    }

    // Update category
    await categoryApi.update(draggingNode.data.id, {
      parent_id: newParentId,
    })
    
    ElMessage.success('移动成功')
    await loadCategoryTree()
  } catch (error) {
    const errMsg = error?.response?.data?.error || error?.message || '移动失败'
    ElMessage.error(`移动失败: ${errMsg}`)
    await loadCategoryTree() // Reload to restore original structure
  }
}

watch(filterText, (val) => {
  treeRef.value?.filter(val)
})

onMounted(() => {
  loadCategoryTree()
})
</script>

<style scoped>
.categories-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.card-header > span {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  font-weight: 500;
}

.header-actions {
  display: flex;
  align-items: center;
}

.category-tree {
  margin-top: 16px;
}

.tree-node-content {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 4px 8px;
  border-radius: 4px;
  transition: background-color 0.2s;
}

.tree-node-content:hover {
  background-color: #f5f7fa;
}

.node-info {
  display: flex;
  align-items: center;
  gap: 8px;
  flex: 1;
}

.node-icon {
  font-size: 16px;
  color: #909399;
}

.node-icon.has-children {
  color: #409eff;
}

.node-name {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
}

.node-actions {
  display: none;
  gap: 8px;
  margin-left: 16px;
}

.tree-node-content:hover .node-actions {
  display: flex;
}

.form-tip {
  display: flex;
  align-items: center;
  gap: 4px;
  margin-top: 6px;
  font-size: 12px;
  color: #909399;
}

.form-tip strong {
  color: #409eff;
  font-weight: 500;
}

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

:deep(.el-tree-node__content) {
  height: auto;
  min-height: 36px;
  padding: 4px 0;
}

:deep(.el-tree-node__expand-icon) {
  font-size: 14px;
}
</style>
