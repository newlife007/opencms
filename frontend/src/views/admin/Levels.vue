<template>
  <div class="levels-management">
    <el-card>
      <template #header>
        <div class="card-header">
          <span>{{ t("admin.levels.title") }}</span>
          <el-button type="primary" icon="Plus" @click="handleAdd">添加等级</el-button>
        </div>
      </template>

      <!-- Levels Table -->
      <el-table :data="levels" v-loading="loading" style="width: 100%" border>
        <el-table-column prop="id" label="ID" width="80" />
        <el-table-column prop="name" label="等级名称" width="150">
          <template #default="{ row }">
            <el-tag type="success">{{ row.name }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="description" :label="t('common.description')" min-width="250" />
        <el-table-column prop="level" label="级别" width="100">
          <template #default="{ row }">
            <el-tag type="warning">{{ row.level }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="enabled" :label="t('files.status')" width="100">
          <template #default="{ row }">
            <el-switch
              v-model="row.enabled"
              @change="handleStatusChange(row)"
            />
          </template>
        </el-table-column>
        <el-table-column :label="t('common.actions')" width="200" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" size="small" icon="Edit" @click="handleEdit(row)">{{ t("common.edit") }}</el-button>
            <el-button type="danger" size="small" icon="Delete" @click="handleDelete(row)">{{ t("common.delete") }}</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <!-- Add/Edit Dialog -->
    <el-dialog 
      v-model="dialogVisible" 
      :title="dialogTitle"
      width="600px"
      @close="resetForm"
    >
      <el-form :model="form" :rules="rules" ref="formRef" label-width="100px">
        <el-form-item label="等级名称" prop="name">
          <el-input v-model="form.name" placeholder="请输入等级名称" />
        </el-form-item>
        <el-form-item :label="t('common.description')" prop="description">
          <el-input 
            v-model="form.description" 
            type="textarea" 
            :rows="3"
            placeholder="请输入描述"
          />
        </el-form-item>
        <el-form-item label="级别" prop="level">
          <el-input-number 
            v-model="form.level" 
            :min="1" 
            :max="10"
            placeholder="级别数值（1-10）"
          />
          <span style="margin-left: 10px; color: #999;">级别越高，可访问的文件越多</span>
        </el-form-item>
        <el-form-item :label="t('files.status')" prop="enabled">
          <el-switch v-model="form.enabled" />
          <span style="margin-left: 10px; color: #999;">
            {{ form.enabled ? '启用' : '禁用' }}
          </span>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">{{ t("common.cancel") }}</el-button>
        <el-button type="primary" @click="handleSubmit" :loading="submitting">{{ t("common.confirm") }}</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup>
import { ref, computed, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { ElMessage, ElMessageBox } from 'element-plus'
import levelsApi from '@/api/levels'


const { t } = useI18n()

const loading = ref(false)
const submitting = ref(false)
const levels = ref([])
const dialogVisible = ref(false)
const isEdit = ref(false)
const formRef = ref(null)

const form = ref({
  id: null,
  name: '',
  description: '',
  level: 1,
  enabled: true
})

const rules = {
  name: [
    { required: true, message: '请输入等级名称', trigger: 'blur' },
    { min: 2, max: 64, message: '长度在 2 到 64 个字符', trigger: 'blur' }
  ],
  description: [
    { max: 255, message: '长度不能超过 255 个字符', trigger: 'blur' }
  ],
  level: [
    { required: true, message: '请输入级别', trigger: 'blur' },
    { type: 'number', min: 1, max: 10, message: '级别范围为 1-10', trigger: 'blur' }
  ]
}

const dialogTitle = computed(() => isEdit.value ? '编辑等级' : '添加等级')

const loadLevels = async () => {
  loading.value = true
  try {
    const res = await levelsApi.getList()
    levels.value = res.data || []
  } catch (error) {
    console.error('加载等级列表失败:', error)
    ElMessage.error('加载等级列表失败: ' + (error.message || '未知错误'))
  } finally {
    loading.value = false
  }
}

const handleAdd = () => {
  isEdit.value = false
  resetForm()
  dialogVisible.value = true
}

const handleEdit = (row) => {
  isEdit.value = true
  form.value = { ...row }
  dialogVisible.value = true
}

const handleDelete = async (row) => {
  try {
    await ElMessageBox.confirm(
      `确定要删除等级 "${row.name}" 吗？删除后将无法恢复。`,
      '警告',
      {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }
    )
    
    await levelsApi.delete(row.id)
    ElMessage.success(t('message.deleteSuccess'))
    loadLevels()
  } catch (error) {
    if (error !== 'cancel') {
      console.error('删除等级失败:', error)
      ElMessage.error('删除失败: ' + (error.message || '未知错误'))
    }
  }
}

const handleStatusChange = async (row) => {
  try {
    await levelsApi.update(row.id, { enabled: row.enabled })
    ElMessage.success('状态更新成功')
  } catch (error) {
    console.error('更新状态失败:', error)
    ElMessage.error('更新状态失败: ' + (error.message || '未知错误'))
    // 恢复原状态
    row.enabled = !row.enabled
  }
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    
    submitting.value = true
    try {
      if (isEdit.value) {
        await levelsApi.update(form.value.id, form.value)
        ElMessage.success('更新成功')
      } else {
        await levelsApi.create(form.value)
        ElMessage.success('添加成功')
      }
      dialogVisible.value = false
      loadLevels()
    } catch (error) {
      console.error('提交失败:', error)
      ElMessage.error('操作失败: ' + (error.message || '未知错误'))
    } finally {
      submitting.value = false
    }
  })
}

const resetForm = () => {
  form.value = {
    id: null,
    name: '',
    description: '',
    level: 1,
    enabled: true
  }
  if (formRef.value) {
    formRef.value.resetFields()
  }
}

onMounted(() => {
  loadLevels()
})
</script>

<style scoped>
.levels-management {
  padding: 20px;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
