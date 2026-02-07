<template>
  <el-form
    ref="formRef"
    :model="formData"
    :rules="formRules"
    :label-width="labelWidth"
    :label-position="labelPosition"
  >
    <template v-for="field in fieldSchema" :key="field.name">
      <el-form-item
        v-if="isFieldVisible(field)"
        :label="field.label"
        :prop="field.name"
        :required="field.required"
      >
        <!-- Text Input -->
        <el-input
          v-if="field.type === 'text' || field.type === 'string'"
          v-model="formData[field.name]"
          :placeholder="field.placeholder || `请输入${field.label}`"
          :disabled="field.disabled"
          clearable
        />

        <!-- Number Input -->
        <el-input-number
          v-else-if="field.type === 'number'"
          v-model="formData[field.name]"
          :min="field.min"
          :max="field.max"
          :step="field.step"
          :disabled="field.disabled"
          style="width: 100%"
        />

        <!-- Textarea -->
        <el-input
          v-else-if="field.type === 'textarea'"
          v-model="formData[field.name]"
          type="textarea"
          :rows="field.rows || 3"
          :placeholder="field.placeholder || `请输入${field.label}`"
          :disabled="field.disabled"
        />

        <!-- Select -->
        <el-select
          v-else-if="field.type === 'select'"
          v-model="formData[field.name]"
          :placeholder="field.placeholder || `请选择${field.label}`"
          :disabled="field.disabled"
          :multiple="field.multiple"
          clearable
          style="width: 100%"
        >
          <el-option
            v-for="option in field.options"
            :key="option.value"
            :label="option.label"
            :value="option.value"
          />
        </el-select>

        <!-- Radio Group -->
        <el-radio-group
          v-else-if="field.type === 'radio'"
          v-model="formData[field.name]"
          :disabled="field.disabled"
        >
          <el-radio
            v-for="option in field.options"
            :key="option.value"
            :label="option.value"
          >
            {{ option.label }}
          </el-radio>
        </el-radio-group>

        <!-- Checkbox Group -->
        <el-checkbox-group
          v-else-if="field.type === 'checkbox'"
          v-model="formData[field.name]"
          :disabled="field.disabled"
        >
          <el-checkbox
            v-for="option in field.options"
            :key="option.value"
            :label="option.value"
          >
            {{ option.label }}
          </el-checkbox>
        </el-checkbox-group>

        <!-- Date Picker -->
        <el-date-picker
          v-else-if="field.type === 'date'"
          v-model="formData[field.name]"
          type="date"
          :placeholder="field.placeholder || `请选择${field.label}`"
          :disabled="field.disabled"
          style="width: 100%"
        />

        <!-- DateTime Picker -->
        <el-date-picker
          v-else-if="field.type === 'datetime'"
          v-model="formData[field.name]"
          type="datetime"
          :placeholder="field.placeholder || `请选择${field.label}`"
          :disabled="field.disabled"
          style="width: 100%"
        />

        <!-- Date Range Picker -->
        <el-date-picker
          v-else-if="field.type === 'daterange'"
          v-model="formData[field.name]"
          type="daterange"
          range-separator="至"
          start-placeholder="开始日期"
          end-placeholder="结束日期"
          :disabled="field.disabled"
          style="width: 100%"
        />

        <!-- Switch -->
        <el-switch
          v-else-if="field.type === 'switch'"
          v-model="formData[field.name]"
          :disabled="field.disabled"
          :active-text="field.activeText"
          :inactive-text="field.inactiveText"
        />

        <!-- Slider -->
        <el-slider
          v-else-if="field.type === 'slider'"
          v-model="formData[field.name]"
          :min="field.min || 0"
          :max="field.max || 100"
          :step="field.step || 1"
          :disabled="field.disabled"
        />

        <!-- File Upload -->
        <el-upload
          v-else-if="field.type === 'upload'"
          :action="field.action"
          :accept="field.accept"
          :disabled="field.disabled"
          :limit="field.limit"
        >
          <el-button size="small" type="primary">点击上传</el-button>
        </el-upload>

        <!-- Custom Slot -->
        <slot
          v-else-if="field.type === 'slot'"
          :name="field.name"
          :field="field"
          :value="formData[field.name]"
        />

        <!-- Help Text -->
        <div v-if="field.help" class="field-help">
          {{ field.help }}
        </div>
      </el-form-item>
    </template>

    <!-- Form Actions -->
    <el-form-item v-if="showActions">
      <slot name="actions">
        <el-button type="primary" @click="handleSubmit">
          {{ submitText }}
        </el-button>
        <el-button @click="handleReset">
          {{ resetText }}
        </el-button>
      </slot>
    </el-form-item>
  </el-form>
</template>

<script setup>
import { ref, computed, watch } from 'vue'

const props = defineProps({
  // Form model data
  modelValue: {
    type: Object,
    default: () => ({}),
  },
  // Field schema array
  schema: {
    type: Array,
    required: true,
  },
  // Label width
  labelWidth: {
    type: String,
    default: '120px',
  },
  // Label position
  labelPosition: {
    type: String,
    default: 'right',
  },
  // Show action buttons
  showActions: {
    type: Boolean,
    default: true,
  },
  // Submit button text
  submitText: {
    type: String,
    default: '提交',
  },
  // Reset button text
  resetText: {
    type: String,
    default: '重置',
  },
})

const emit = defineEmits(['update:modelValue', 'submit', 'reset'])

const formRef = ref(null)
const formData = ref({ ...props.modelValue })

// Watch for external changes
watch(
  () => props.modelValue,
  (newVal) => {
    formData.value = { ...newVal }
  },
  { deep: true }
)

// Watch for internal changes
watch(
  formData,
  (newVal) => {
    emit('update:modelValue', newVal)
  },
  { deep: true }
)

// Build field schema with defaults
const fieldSchema = computed(() => {
  return props.schema.map(field => ({
    type: 'text',
    required: false,
    disabled: false,
    ...field,
  }))
})

// Build validation rules
const formRules = computed(() => {
  const rules = {}
  fieldSchema.value.forEach(field => {
    if (field.required || field.rules) {
      rules[field.name] = []
      
      // Required rule
      if (field.required) {
        rules[field.name].push({
          required: true,
          message: `${field.label}不能为空`,
          trigger: field.type === 'select' ? 'change' : 'blur',
        })
      }
      
      // Custom rules
      if (field.rules && Array.isArray(field.rules)) {
        rules[field.name].push(...field.rules)
      }
    }
  })
  return rules
})

// Check if field should be visible based on conditions
const isFieldVisible = (field) => {
  if (!field.condition) return true
  
  const { field: conditionField, value: conditionValue } = field.condition
  return formData.value[conditionField] === conditionValue
}

const handleSubmit = async () => {
  if (!formRef.value) return
  
  try {
    await formRef.value.validate()
    emit('submit', formData.value)
  } catch (error) {
    console.error('Form validation failed:', error)
  }
}

const handleReset = () => {
  formRef.value?.resetFields()
  emit('reset')
}

const validate = () => {
  return formRef.value?.validate()
}

const resetFields = () => {
  formRef.value?.resetFields()
}

const clearValidate = () => {
  formRef.value?.clearValidate()
}

// Expose methods
defineExpose({
  validate,
  resetFields,
  clearValidate,
})
</script>

<style scoped>
.field-help {
  font-size: 12px;
  color: #909399;
  line-height: 1.5;
  margin-top: 4px;
}
</style>
