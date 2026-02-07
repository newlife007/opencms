<template>
  <el-dropdown trigger="click" @command="handleLanguageChange">
    <span class="language-switcher">
      <el-icon><Iphone /></el-icon>
      <span class="language-label">{{ currentLanguageLabel }}</span>
      <el-icon class="el-icon--right"><arrow-down /></el-icon>
    </span>
    <template #dropdown>
      <el-dropdown-menu>
        <el-dropdown-item 
          v-for="locale in availableLocales" 
          :key="locale.value"
          :command="locale.value"
          :class="{ 'is-active': currentLocale === locale.value }"
        >
          <el-icon v-if="currentLocale === locale.value"><Check /></el-icon>
          <span :style="{ marginLeft: currentLocale === locale.value ? '8px' : '28px' }">
            {{ locale.label }}
          </span>
        </el-dropdown-item>
      </el-dropdown-menu>
    </template>
  </el-dropdown>
</template>

<script setup>
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { setLocale, getLocale, availableLocales } from '@/i18n'
import { Iphone, ArrowDown, Check } from '@element-plus/icons-vue'
import { ElMessage } from 'element-plus'

const { locale } = useI18n()

const currentLocale = computed(() => getLocale())

const currentLanguageLabel = computed(() => {
  const current = availableLocales.find(l => l.value === currentLocale.value)
  return current ? current.label : 'Language'
})

const handleLanguageChange = (localeValue) => {
  if (localeValue === currentLocale.value) return
  
  setLocale(localeValue)
  
  // 刷新页面以应用新语言（Element Plus 组件需要重新加载）
  window.location.reload()
}
</script>

<style scoped>
.language-switcher {
  display: flex;
  align-items: center;
  cursor: pointer;
  padding: 5px 12px;
  border-radius: 4px;
  transition: background-color 0.3s;
}

.language-switcher:hover {
  background-color: var(--el-fill-color-light);
}

.language-label {
  margin: 0 8px;
  font-size: 14px;
}

.el-dropdown-menu__item.is-active {
  color: var(--el-color-primary);
  font-weight: 500;
}
</style>
