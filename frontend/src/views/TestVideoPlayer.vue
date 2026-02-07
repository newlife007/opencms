<template>
  <div style="padding: 20px;">
    <h1>VideoPlayer组件测试</h1>
    
    <div style="background: #fff3cd; padding: 10px; margin: 20px 0;">
      <h3>测试说明</h3>
      <p>这个页面直接测试VideoPlayer组件是否能正常渲染和工作</p>
    </div>

    <div style="background: white; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
      <h2>测试1: 使用硬编码的视频URL</h2>
      <div style="background: #e3f2fd; padding: 10px; margin: 10px 0;">
        <p>预览URL: {{ testUrl }}</p>
        <p>VideoPlayer组件应该显示在下方:</p>
      </div>
      <div style="background: #000; min-height: 400px;">
        <VideoPlayer
          v-if="testUrl"
          :src="testUrl"
          type="video/mp4"
        />
      </div>
    </div>

    <div style="background: white; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
      <h2>测试2: 登录后测试真实API</h2>
      <button @click="login" :disabled="isLoggedIn" style="padding: 10px 20px; margin: 5px;">
        {{ isLoggedIn ? '✓ 已登录' : '点击登录' }}
      </button>
      <button @click="loadPreview" :disabled="!isLoggedIn" style="padding: 10px 20px; margin: 5px;">
        加载预览
      </button>
      
      <div v-if="previewUrl" style="margin-top: 20px;">
        <div style="background: #d4edda; padding: 10px; margin: 10px 0;">
          <p>✓ 预览URL已加载: {{ previewUrl }}</p>
          <p>VideoPlayer组件应该显示在下方:</p>
        </div>
        <div style="background: #000; min-height: 400px;">
          <VideoPlayer
            :src="previewUrl"
            type="video/x-flv"
          />
        </div>
      </div>
    </div>

    <div style="background: white; padding: 20px; margin: 20px 0; box-shadow: 0 2px 4px rgba(0,0,0,0.1);">
      <h2>控制台日志</h2>
      <div style="background: #f5f5f5; padding: 10px; font-family: monospace; white-space: pre-wrap; max-height: 300px; overflow-y: auto;">
        {{ logs }}
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import VideoPlayer from '@/components/VideoPlayer.vue'

const testUrl = ref('http://localhost:8080/api/v1/files/71/preview')
const isLoggedIn = ref(false)
const previewUrl = ref('')
const logs = ref('')

function addLog(message) {
  const timestamp = new Date().toLocaleTimeString()
  logs.value += `[${timestamp}] ${message}\n`
  console.log(message)
}

addLog('页面加载完成')
addLog('测试URL: ' + testUrl.value)

async function login() {
  addLog('开始登录...')
  try {
    const response = await fetch('http://localhost:8080/api/v1/auth/login', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        username: 'testuser',
        password: 'testpass123'
      }),
      credentials: 'include'
    })

    if (response.ok) {
      isLoggedIn.value = true
      addLog('✓ 登录成功')
    } else {
      addLog('✗ 登录失败: ' + response.statusText)
    }
  } catch (error) {
    addLog('✗ 登录错误: ' + error.message)
  }
}

async function loadPreview() {
  addLog('加载预览...')
  previewUrl.value = 'http://localhost:8080/api/v1/files/71/preview'
  addLog('✓ 预览URL设置完成: ' + previewUrl.value)
}
</script>
