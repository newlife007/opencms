<template>
  <el-dialog
    v-model="visible"
    :title="isEdit ? '编辑用户' : '添加用户'"
    width="600px"
    :close-on-click-modal="false"
    @close="handleClose"
  >
    <el-form
      ref="formRef"
      :model="formData"
      :rules="formRules"
      label-width="100px"
    >
      <el-form-item label="用户名" prop="username">
        <el-input
          v-model="formData.username"
          placeholder="请输入用户名"
          :disabled="isEdit"
          clearable
        />
        <div class="form-tip">用户名一旦创建不可修改</div>
      </el-form-item>

      <el-form-item v-if="!isEdit" label="密码" prop="password">
        <el-input
          v-model="formData.password"
          type="password"
          placeholder="请输入密码"
          show-password
          clearable
        />
        <div class="form-tip">密码长度至少6位</div>
      </el-form-item>

      <el-form-item label="昵称" prop="nickname">
        <el-input
          v-model="formData.nickname"
          placeholder="请输入昵称（选填）"
          clearable
        />
      </el-form-item>

      <el-form-item label="邮箱" prop="email">
        <el-input
          v-model="formData.email"
          placeholder="请输入邮箱"
          clearable
        />
      </el-form-item>

      <el-form-item label="所属组" prop="group_id">
        <el-select
          v-model="formData.group_id"
          placeholder="请选择所属组"
          style="width: 100%"
          clearable
        >
          <el-option
            v-for="group in groups"
            :key="group.id"
            :label="group.name"
            :value="group.id"
          />
        </el-select>
      </el-form-item>

      <el-form-item label="用户等级" prop="level_id">
        <el-select
          v-model="formData.level_id"
          placeholder="请选择用户等级"
          style="width: 100%"
          clearable
        >
          <el-option
            v-for="level in levels"
            :key="level.id"
            :label="`Level ${level.id} - ${level.name}`"
            :value="level.id"
          />
        </el-select>
        <div class="form-tip">等级越高，可访问的资源越多</div>
      </el-form-item>

      <el-form-item label="状态" prop="status">
        <el-radio-group v-model="formData.status">
          <el-radio :label="1">启用</el-radio>
          <el-radio :label="0">禁用</el-radio>
        </el-radio-group>
      </el-form-item>

      <el-form-item label="备注" prop="remark">
        <el-input
          v-model="formData.remark"
          type="textarea"
          :rows="3"
          placeholder="请输入备注（选填）"
        />
      </el-form-item>
    </el-form>

    <template #footer>
      <el-button @click="handleClose">取消</el-button>
      <el-button type="primary" :loading="loading" @click="handleSubmit">
        确定
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup>
import { ref, computed, watch } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps({
  modelValue: {
    type: Boolean,
    default: false,
  },
  userData: {
    type: Object,
    default: null,
  },
  groups: {
    type: Array,
    default: () => [],
  },
  levels: {
    type: Array,
    default: () => [],
  },
})

const emit = defineEmits(['update:modelValue', 'success'])

const visible = computed({
  get: () => props.modelValue,
  set: (val) => emit('update:modelValue', val),
})

const isEdit = computed(() => !!props.userData?.id)

const formRef = ref(null)
const loading = ref(false)

const defaultFormData = {
  username: '',
  password: '',
  nickname: '',
  email: '',
  group_id: null,
  level_id: null,
  status: 1,
  remark: '',
}

const formData = ref({ ...defaultFormData })

// 验证用户名（字母、数字、下划线，4-20位）
const validateUsername = (rule, value, callback) => {
  if (!value) {
    callback(new Error('请输入用户名'))
  } else if (!/^[a-zA-Z0-9_]{4,20}$/.test(value)) {
    callback(new Error('用户名只能包含字母、数字、下划线，长度4-20位'))
  } else {
    callback()
  }
}

// 验证密码强度
const validatePassword = (rule, value, callback) => {
  if (!isEdit.value && !value) {
    callback(new Error('请输入密码'))
  } else if (value && value.length < 6) {
    callback(new Error('密码长度至少6位'))
  } else {
    callback()
  }
}

// 验证邮箱
const validateEmail = (rule, value, callback) => {
  if (value && !/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(value)) {
    callback(new Error('请输入有效的邮箱地址'))
  } else {
    callback()
  }
}

const formRules = {
  username: [{ validator: validateUsername, trigger: 'blur' }],
  password: [{ validator: validatePassword, trigger: 'blur' }],
  email: [{ validator: validateEmail, trigger: 'blur' }],
  group_id: [{ required: true, message: '请选择所属组', trigger: 'change' }],
  level_id: [{ required: true, message: '请选择用户等级', trigger: 'change' }],
}

// Watch userData changes
watch(
  () => props.userData,
  (newVal) => {
    if (newVal) {
      formData.value = {
        username: newVal.username || '',
        nickname: newVal.nickname || '',
        email: newVal.email || '',
        group_id: newVal.group_id || null,
        level_id: newVal.level_id || null,
        status: newVal.status ?? 1,
        remark: newVal.remark || '',
      }
    } else {
      formData.value = { ...defaultFormData }
    }
  },
  { immediate: true }
)

const handleClose = () => {
  visible.value = false
  formRef.value?.resetFields()
  formData.value = { ...defaultFormData }
}

const handleSubmit = async () => {
  if (!formRef.value) return

  try {
    await formRef.value.validate()
    loading.value = true

    // 提交数据（父组件处理实际的API调用）
    emit('success', { ...formData.value, id: props.userData?.id })

    ElMessage.success(isEdit.value ? '更新成功' : '创建成功')
    handleClose()
  } catch (error) {
    if (error.message) {
      ElMessage.error(error.message)
    }
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.form-tip {
  font-size: 12px;
  color: #999;
  line-height: 1.5;
  margin-top: 4px;
}
</style>
