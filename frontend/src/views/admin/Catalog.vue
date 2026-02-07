<template>
  <div class="catalog-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("admin.catalog.title") }}</span>
          <div>
            <el-radio-group v-model="currentFileType" @change="loadCatalogTree">
              <el-radio-button :label="1">视频</el-radio-button>
              <el-radio-button :label="2">音频</el-radio-button>
              <el-radio-button :label="3">图片</el-radio-button>
              <el-radio-button :label="4">富媒体</el-radio-button>
            </el-radio-group>
          </div>
        </div>
      </template>

      <el-row :gutter="20">
        <!-- Left: Catalog Tree -->
        <el-col :span="12">
          <div class="tree-container">
            <div class="tree-header">
              <span>属性结构</span>
              <el-button type="primary" icon="Plus" size="small" @click="handleAdd(null)">
                添加根属性
              </el-button>
            </div>

            <el-tree
              ref="treeRef"
              :data="catalogTree"
              :props="treeProps"
              node-key="id"
              default-expand-all
              draggable
              @node-drop="handleDrop"
              v-loading="loading"
            >
              <template #default="{ node, data }">
                <div class="tree-node">
                  <span class="node-label">
                    <el-icon><List /></el-icon>
                    {{ node.label }}
                    <el-tag v-if="!data.enabled" size="small" type="info">禁用</el-tag>
                  </span>
                  <span class="node-actions">
                    <el-button
                      type="primary"
                      size="small"
                      icon="Plus"
                      circle
                      @click.stop="handleAdd(data)"
                    />
                    <el-button
                      type="success"
                      size="small"
                      icon="Edit"
                      circle
                      @click.stop="handleEdit(data)"
                    />
                    <el-button
                      type="danger"
                      size="small"
                      icon="Delete"
                      circle
                      @click.stop="handleDelete(data)"
                    />
                  </span>
                </div>
              </template>
            </el-tree>
          </div>
        </el-col>

        <!-- Right: Catalog Form -->
        <el-col :span="12">
          <div class="form-container">
            <div class="form-header">
              <span>{{ formTitle }}</span>
            </div>

            <el-form
              ref="catalogFormRef"
              :model="catalogForm"
              :rules="catalogRules"
              label-width="120px"
            >
              <el-form-item label="上级属性">
                <el-tree-select
                  v-model="catalogForm.parent_id"
                  :data="catalogTree"
                  :props="treeProps"
                  node-key="id"
                  value-key="id"
                  placeholder="请选择上级属性（不选则为根属性）"
                  clearable
                  check-strictly
                />
              </el-form-item>

              <el-form-item label="属性名称" prop="name">
                <el-input
                  v-model="catalogForm.name"
                  placeholder="请输入属性英文名称，如: director、duration、title"
                />
                <span class="form-tip">英文字段名，用于数据存储（如: director, title, duration）</span>
              </el-form-item>

              <el-form-item label="显示标签" prop="label">
                <el-input
                  v-model="catalogForm.label"
                  placeholder="请输入中文显示名称，如: 导演、时长、标题"
                />
                <span class="form-tip">中文名称，用于界面显示（如: 导演、标题、时长）</span>
              </el-form-item>

              <el-form-item label="字段类型" prop="field_type">
                <el-select
                  v-model="catalogForm.field_type"
                  placeholder="请选择字段类型"
                  style="width: 100%"
                >
                  <el-option label="分组（用于组织子字段）" value="group" />
                  <el-option label="文本（单行输入）" value="text" />
                  <el-option label="数字" value="number" />
                  <el-option label="日期" value="date" />
                  <el-option label="下拉选择" value="select" />
                  <el-option label="多行文本" value="textarea" />
                </el-select>
                <span class="form-tip">选择字段在编目表单中的展示方式</span>
              </el-form-item>

              <el-form-item label="是否必填">
                <el-switch
                  v-model="catalogForm.required"
                  active-text="必填"
                  inactive-text="可选"
                />
              </el-form-item>

              <el-form-item 
                label="选项配置" 
                v-if="catalogForm.field_type === 'select'"
              >
                <el-input
                  v-model="catalogForm.options"
                  type="textarea"
                  :rows="3"
                  placeholder='JSON格式，如: ["选项1", "选项2", "选项3"]'
                />
                <span class="form-tip">下拉选项（JSON数组格式）</span>
              </el-form-item>

              <el-form-item label="属性描述">
                <el-input
                  v-model="catalogForm.description"
                  type="textarea"
                  :rows="2"
                  placeholder="请输入属性的说明，如: 视频标题、导演姓名"
                />
              </el-form-item>

              <el-form-item label="排序权重">
                <el-input-number
                  v-model="catalogForm.weight"
                  :min="0"
                  :max="999"
                  placeholder="数字越大越靠前"
                />
                <span class="form-tip">数字越大排序越靠前</span>
              </el-form-item>

              <el-form-item :label="t('files.status')">
                <el-switch
                  v-model="catalogForm.enabled"
                  :active-value="true"
                  :inactive-value="false"
                  active-text="启用"
                  inactive-text="禁用"
                />
              </el-form-item>

              <el-form-item>
                <el-button type="primary" @click="handleSubmit" :loading="submitting">
                  {{ isEdit ? '更新' : '创建' }}
                </el-button>
                <el-button @click="handleReset">{{ t("common.reset") }}</el-button>
                <el-button type="success" @click="handlePreview">{{ t("admin.catalog.previewForm") }}</el-button>
              </el-form-item>
            </el-form>
          </div>
        </el-col>
      </el-row>
    </el-card>

    <!-- Form Preview Dialog -->
    <el-dialog v-model="previewDialogVisible" title="表单预览" width="800px">
      <div class="form-preview">
        <el-alert
          title="这是根据当前属性配置生成的编目表单预览"
          type="info"
          :closable="false"
          style="margin-bottom: 20px"
        >
          <template #default>
            <div style="margin-top: 8px; font-size: 13px; color: #606266;">
              当用户上传{{ fileTypeNames[currentFileType] }}文件时，会看到以下编目表单。
              每个属性对应一个输入框，用于填写文件的详细信息。
            </div>
          </template>
        </el-alert>
        
        <el-form label-width="120px" label-position="right">
          <template v-for="category in catalogTree" :key="category.id">
            <!-- 分类标题 -->
            <el-divider content-position="left">
              <el-icon><Folder /></el-icon>
              {{ category.label || category.name }}
            </el-divider>
            
            <!-- 分类下的属性 -->
            <template v-if="category.children && category.children.length">
              <el-form-item 
                v-for="field in category.children" 
                :key="field.id"
                :label="field.label || field.name"
              >
                <el-input 
                  :placeholder="`请输入${field.name}`"
                  disabled
                  style="width: 100%"
                >
                  <template #suffix>
                    <el-tooltip :content="field.description || '暂无说明'" placement="top">
                      <el-icon style="cursor: help;"><InfoFilled /></el-icon>
                    </el-tooltip>
                  </template>
                </el-input>
                <div v-if="field.description" class="field-desc">
                  {{ field.description }}
                </div>
              </el-form-item>
            </template>
            
            <!-- 如果分类没有子属性 -->
            <el-empty 
              v-else 
              description="该分类暂无子属性" 
              :image-size="60"
              style="padding: 10px 0;"
            />
          </template>
          
          <!-- 如果没有任何属性 -->
          <el-empty 
            v-if="!catalogTree || catalogTree.length === 0"
            description="当前文件类型暂无属性配置"
            :image-size="100"
          />
        </el-form>
      </div>
      
      <template #footer>
        <el-button @click="previewDialogVisible = false">{{ t("common.close") }}</el-button>
        <el-button type="primary" @click="previewDialogVisible = false">
          了解
        </el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, reactive, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import catalogApi from '@/api/catalog'


const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const isEdit = ref(false)
const currentFileType = ref(1)
const catalogTree = ref([])
const previewDialogVisible = ref(false)

// File type names for display
const fileTypeNames = {
  1: '视频',
  2: '音频',
  3: '图片',
  4: '富媒体'
}


const treeRef = ref(null)
const catalogFormRef = ref(null)

const treeProps = {
  label: 'label',     // 显示字段：使用 label 作为显示标签（中文名称）
  children: 'children',
  value: 'id',        // 值字段：使用 id 作为选中的值
}

const catalogForm = reactive({
  id: null,
  parent_id: null,
  name: '',
  label: '',        // 中文显示名称
  description: '',
  field_type: 'text',  // ✅ 字段类型：text, number, date, select, textarea, group
  required: false,     // ✅ 是否必填
  options: '',         // ✅ 选项（JSON字符串，用于select类型）
  weight: 0,
  enabled: true,       // ✅ 修复：使用布尔值
})

const catalogRules = {
  name: [
    { required: true, message: '请输入属性名称（英文字段名）', trigger: 'blur' },
    { 
      pattern: /^[a-zA-Z][a-zA-Z0-9_]*$/, 
      message: '属性名称只能包含英文字母、数字和下划线，且必须以字母开头', 
      trigger: 'blur' 
    },
    { 
      min: 2, 
      max: 64, 
      message: '属性名称长度必须在 2-64 个字符之间', 
      trigger: 'blur' 
    },
  ],
  label: [
    { required: true, message: '请输入显示标签（中文名称）', trigger: 'blur' },
    { 
      min: 1, 
      max: 64, 
      message: '显示标签长度必须在 1-64 个字符之间', 
      trigger: 'blur' 
    },
  ],
}

const formTitle = computed(() => {
  if (!isEdit.value) {
    return catalogForm.parent_id ? '添加子字段' : '添加根字段'
  }
  return '编辑属性'
})

const loadCatalogTree = async () => {
  loading.value = true
  try {
    console.log('Loading catalog tree for file type:', currentFileType.value)
    const res = await catalogApi.getTreeByType(currentFileType.value)
    console.log('Catalog tree API response:', res)
    
    if (res.success) {
      catalogTree.value = res.data || []
      console.log('Catalog tree loaded:', catalogTree.value.length, 'root nodes')
      console.log('Tree data:', JSON.stringify(catalogTree.value, null, 2))
    } else {
      console.warn('API returned success=false:', res)
    }
  } catch (error) {
    console.error('Load catalog tree error:', error)
    ElMessage.error('加载属性配置失败')
  } finally {
    loading.value = false
  }
}

const handleAdd = (parentCatalog) => {
  isEdit.value = false
  Object.assign(catalogForm, {
    id: null,
    parent_id: parentCatalog?.id || null,
    name: '',
    label: '',           // 中文显示名称
    description: '',
    field_type: 'text',  // ✅ 默认文本类型
    required: false,     // ✅ 默认非必填
    options: '',         // ✅ 默认无选项
    weight: 0,
    enabled: true,       // ✅ 修复：使用布尔值
  })
}

const handleEdit = async (catalog) => {
  isEdit.value = true
  
  // 直接使用树节点数据，不需要再次请求API
  Object.assign(catalogForm, {
    id: catalog.id,
    parent_id: catalog.parent_id,
    name: catalog.name,
    label: catalog.label || '',             // 中文显示名称
    description: catalog.description || '',
    field_type: catalog.field_type || 'text',  // ✅ 字段类型
    required: catalog.required || false,       // ✅ 是否必填
    options: catalog.options || '',            // ✅ 选项
    weight: catalog.weight || 0,
    enabled: catalog.enabled ? true : false,  // ✅ 修复：布尔值
  })
}

const handleReset = () => {
  catalogFormRef.value?.resetFields()
  isEdit.value = false
}

const handleSubmit = async () => {
  if (!catalogFormRef.value) return

  await catalogFormRef.value.validate(async (valid) => {
    if (!valid) return

    submitting.value = true
    try {
      // 构建提交数据，包含所有必需字段
      const data = {
        type: currentFileType.value,              // ✅ 添加文件类型（必需）
        parent_id: catalogForm.parent_id || 0,    // 根节点的parent_id为0
        name: catalogForm.name,                   // 英文字段名（必需）
        label: catalogForm.label || catalogForm.name,  // ✅ 显示标签（必需）
        description: catalogForm.description || '',
        field_type: catalogForm.field_type || 'text',  // ✅ 字段类型（必需）
        required: catalogForm.required ? true : false,
        options: catalogForm.options || '',
        weight: catalogForm.weight || 0,
        enabled: catalogForm.enabled ? true : false,  // ✅ 修复：布尔值
      }
      
      if (isEdit.value) {
        // 编辑时包含 id
        data.id = catalogForm.id
        await catalogApi.update(data.id, data)
        ElMessage.success('更新成功')
      } else {
        // 创建时不需要 id
        delete data.id
        await catalogApi.create(data)
        ElMessage.success('创建成功')
      }
      
      await loadCatalogTree()
      handleReset()
    } catch (error) {
      console.error('Submit error:', error)
      const errorMsg = error.response?.data?.error || error.message || '未知错误'
      ElMessage.error(`${isEdit.value ? '更新' : '创建'}失败: ${errorMsg}`)
    } finally {
      submitting.value = false
    }
  })
}

const handleDelete = async (catalog) => {
  if (catalog.children && catalog.children.length > 0) {
    ElMessage.warning('请先删除子字段')
    return
  }

  try {
    await ElMessageBox.confirm(
      `确定要删除字段"${catalog.label}"吗？删除后无法恢复！`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'error',
      }
    )
    
    await catalogApi.delete(catalog.id)
    ElMessage.success(t('message.deleteSuccess'))
    await loadCatalogTree()
  } catch (error) {
    if (error !== 'cancel') {
      ElMessage.error(t('message.deleteFailed'))
    }
  }
}

const handleDrop = async (draggingNode, dropNode, dropType) => {
  try {
    let newParentId = null
    if (dropType === 'inner') {
      newParentId = dropNode.data.id
    } else {
      newParentId = dropNode.data.parent_id
    }

    await catalogApi.update(draggingNode.data.id, {
      parent_id: newParentId,
      weight: draggingNode.data.weight,
    })
    
    ElMessage.success('移动成功')
    await loadCatalogTree()
  } catch (error) {
    ElMessage.error('移动失败')
    await loadCatalogTree()
  }
}

const handlePreview = () => {
  previewDialogVisible.value = true
}

onMounted(() => {
  loadCatalogTree()
})
</script>

<style scoped>
.catalog-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}

.tree-container,
.form-container {
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 15px;
  min-height: 600px;
}

.tree-header,
.form-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
  padding-bottom: 10px;
  border-bottom: 1px solid #ebeef5;
  font-weight: 500;
}

.tree-node {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding-right: 10px;
}

.node-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.node-actions {
  display: none;
}

.tree-node:hover .node-actions {
  display: flex;
  gap: 5px;
}

.form-preview {
  max-height: 600px;
  overflow-y: auto;
}

.field-desc {
  font-size: 12px;
  color: #909399;
  margin-top: 4px;
  line-height: 1.4;
}

.form-tip {
  font-size: 12px;
  color: #909399;
  margin-left: 10px;
  line-height: 1.4;
}

.form-preview :deep(.el-divider__text) {
  font-weight: 500;
  color: #303133;
}

.form-preview :deep(.el-form-item) {
  margin-bottom: 18px;
}
</style>
