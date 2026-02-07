<template>
  <div class="login-container">
    <!-- 左侧背景区域 -->
    <div class="login-left">
      <div class="left-content">
        <div class="logo-section">
          <el-icon class="logo-icon" :size="80">
            <VideoPlay />
          </el-icon>
          <h1 class="system-title">OpenWan</h1>
          <p class="system-subtitle">{{ t('auth.systemTitle') }}</p>
        </div>
        
        <div class="feature-list">
          <div class="feature-item">
            <el-icon class="feature-icon" :size="24">
              <VideoCamera />
            </el-icon>
            <div class="feature-text">
              <h3>{{ t('auth.feature1Title') }}</h3>
              <p>{{ t('auth.feature1Desc') }}</p>
            </div>
          </div>
          
          <div class="feature-item">
            <el-icon class="feature-icon" :size="24">
              <Search />
            </el-icon>
            <div class="feature-text">
              <h3>{{ t('auth.feature2Title') }}</h3>
              <p>{{ t('auth.feature2Desc') }}</p>
            </div>
          </div>
          
          <div class="feature-item">
            <el-icon class="feature-icon" :size="24">
              <Lock />
            </el-icon>
            <div class="feature-text">
              <h3>{{ t('auth.feature3Title') }}</h3>
              <p>{{ t('auth.feature3Desc') }}</p>
            </div>
          </div>
          
          <div class="feature-item">
            <el-icon class="feature-icon" :size="24">
              <Platform />
            </el-icon>
            <div class="feature-text">
              <h3>{{ t('auth.feature4Title') }}</h3>
              <p>{{ t('auth.feature4Desc') }}</p>
            </div>
          </div>
        </div>
        
        <div class="copyright">
          <p>{{ t('auth.copyright1') }}</p>
          <p>{{ t('auth.copyright2') }}</p>
        </div>
      </div>
    </div>
    
    <!-- 右侧登录区域 -->
    <div class="login-right">
      <div class="login-box">
        <div class="login-header">
          <h2>{{ t('auth.welcomeLogin') }}</h2>
          <p>{{ t('auth.welcomeTo') }}</p>
        </div>
        
        <el-form
          ref="loginFormRef"
          :model="loginForm"
          :rules="loginRules"
          class="login-form"
          @keyup.enter="handleLogin"
        >
          <el-form-item prop="username">
            <el-input
              v-model="loginForm.username"
              :placeholder="t('auth.pleaseEnterUsername')"
              size="large"
              prefix-icon="User"
              clearable
            >
              <template #prefix>
                <el-icon><User /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          
          <el-form-item prop="password">
            <el-input
              v-model="loginForm.password"
              type="password"
              :placeholder="t('auth.pleaseEnterPassword')"
              size="large"
              show-password
              clearable
            >
              <template #prefix>
                <el-icon><Lock /></el-icon>
              </template>
            </el-input>
          </el-form-item>
          
          <el-form-item>
            <el-button
              :loading="loading"
              type="primary"
              size="large"
              class="login-button"
              @click="handleLogin"
            >
              <span v-if="!loading">{{ t('auth.loginButton') }}</span>
              <span v-else>{{ t('auth.loggingIn') }}</span>
            </el-button>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, reactive, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useUserStore } from '@/stores/user'
import { ElMessage } from 'element-plus'
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
const router = useRouter()
const route = useRoute()
const userStore = useUserStore()

const loginFormRef = ref(null)
const loading = ref(false)

const loginForm = reactive({
  username: '',
  password: '',
})

const loginRules = computed(() => ({
  username: [
    { required: true, message: t('auth.pleaseEnterUsername'), trigger: 'blur' },
  ],
  password: [
    { required: true, message: t('auth.pleaseEnterPassword'), trigger: 'blur' },
  ],
}))

const handleLogin = async () => {
  if (!loginFormRef.value) return

  await loginFormRef.value.validate(async (valid) => {
    if (!valid) return

    loading.value = true
    try {
      const success = await userStore.login(loginForm)
      if (success) {
        ElMessage.success('登录成功')
        const redirect = route.query.redirect || '/dashboard'
        router.push(redirect)
      } else {
        // Don't show error here, the error is already shown by request interceptor
        // ElMessage.error('登录失败，请检查用户名和密码')
      }
    } catch (error) {
      // Error message is already shown by request interceptor
      // Just log it for debugging
      console.error('Login error:', error)
    } finally {
      loading.value = false
    }
  })
}
</script>

<style scoped>
.login-container {
  display: flex;
  height: 100vh;
  overflow: hidden;
}

/* 左侧背景区域 */
.login-left {
  flex: 7;
  min-width: 0;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  position: relative;
  overflow: hidden;
}

.login-left::before {
  content: '';
  position: absolute;
  top: -50%;
  left: -50%;
  width: 200%;
  height: 200%;
  background: radial-gradient(circle, rgba(255, 255, 255, 0.1) 1px, transparent 1px);
  background-size: 50px 50px;
  animation: backgroundMove 20s linear infinite;
}

@keyframes backgroundMove {
  0% {
    transform: translate(0, 0);
  }
  100% {
    transform: translate(50px, 50px);
  }
}

.left-content {
  position: relative;
  z-index: 1;
  max-width: 600px;
  padding: 60px;
}

.logo-section {
  text-align: center;
  margin-bottom: 60px;
  animation: fadeInDown 1s ease-out;
}

@keyframes fadeInDown {
  from {
    opacity: 0;
    transform: translateY(-30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.logo-icon {
  color: #fff;
  margin-bottom: 20px;
  filter: drop-shadow(0 5px 15px rgba(0, 0, 0, 0.3));
}

.system-title {
  font-size: 48px;
  font-weight: 700;
  margin: 0 0 10px 0;
  text-shadow: 0 2px 10px rgba(0, 0, 0, 0.3);
  letter-spacing: 2px;
}

.system-subtitle {
  font-size: 20px;
  font-weight: 300;
  margin: 0;
  opacity: 0.95;
  letter-spacing: 4px;
}

.feature-list {
  margin-top: 40px;
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.feature-item {
  display: flex;
  align-items: flex-start;
  padding: 20px;
  background: rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  backdrop-filter: blur(10px);
  transition: all 0.3s ease;
  animation: fadeInUp 1s ease-out;
  animation-fill-mode: both;
}

.feature-item:nth-child(1) { animation-delay: 0.2s; }
.feature-item:nth-child(2) { animation-delay: 0.4s; }
.feature-item:nth-child(3) { animation-delay: 0.6s; }
.feature-item:nth-child(4) { animation-delay: 0.8s; }

@keyframes fadeInUp {
  from {
    opacity: 0;
    transform: translateY(30px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.feature-item:hover {
  background: rgba(255, 255, 255, 0.15);
  transform: scale(1.02);
}

.feature-icon {
  flex-shrink: 0;
  margin-right: 15px;
  color: #fff;
}

.feature-text h3 {
  font-size: 18px;
  font-weight: 600;
  margin: 0 0 8px 0;
  color: #fff;
}

.feature-text p {
  font-size: 14px;
  margin: 0;
  opacity: 0.9;
  line-height: 1.6;
}

.copyright {
  text-align: center;
  margin-top: 60px;
  opacity: 0.8;
  font-size: 13px;
  animation: fadeIn 1s ease-out 1s both;
}

@keyframes fadeIn {
  from { opacity: 0; }
  to { opacity: 0.8; }
}

.copyright p {
  margin: 5px 0;
}

/* 右侧登录区域 */
.login-right {
  flex: 3;
  min-width: 0;
  background: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  box-shadow: -5px 0 20px rgba(0, 0, 0, 0.1);
}

.login-box {
  width: 100%;
  max-width: 360px;
  padding: 40px 20px;
  animation: fadeInRight 0.8s ease-out;
}

@keyframes fadeInRight {
  from {
    opacity: 0;
    transform: translateX(30px);
  }
  to {
    opacity: 1;
    transform: translateX(0);
  }
}

.login-header {
  text-align: center;
  margin-bottom: 40px;
}

.login-header h2 {
  font-size: 28px;
  color: #303133;
  margin: 0 0 10px 0;
  font-weight: 600;
}

.login-header p {
  font-size: 14px;
  color: #909399;
  margin: 0;
  font-weight: 300;
}

.login-form {
  margin-top: 30px;
}

.login-form :deep(.el-form-item) {
  margin-bottom: 24px;
}

.login-form :deep(.el-input__wrapper) {
  padding: 12px 15px;
  border-radius: 8px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.05);
  transition: all 0.3s ease;
}

.login-form :deep(.el-input__wrapper:hover) {
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.1);
}

.login-form :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 2px rgba(102, 126, 234, 0.2);
}

.login-button {
  width: 100%;
  height: 44px;
  font-size: 16px;
  font-weight: 500;
  border-radius: 8px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  border: none;
  transition: all 0.3s ease;
}

.login-button:hover {
  transform: translateY(-2px);
  box-shadow: 0 5px 15px rgba(102, 126, 234, 0.4);
}

.login-button:active {
  transform: translateY(0);
}

/* 响应式设计 */
@media (max-width: 1200px) {
  .left-content {
    padding: 40px;
    max-width: 500px;
  }
  
  .feature-list {
    margin-top: 30px;
    gap: 15px;
  }
  
  .feature-item {
    padding: 15px;
  }
}

@media (max-width: 900px) {
  .feature-list {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 768px) {
  .login-container {
    flex-direction: column;
  }
  
  .login-left {
    display: none;
  }
  
  .login-right {
    width: 100%;
    padding: 20px;
  }
  
  .login-box {
    max-width: 100%;
  }
}
</style>
