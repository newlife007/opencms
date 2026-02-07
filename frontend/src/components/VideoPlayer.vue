<template>
  <div class="video-player-wrapper">
    <video
      ref="videoElement"
      class="video-js vjs-default-skin vjs-big-play-centered"
      controls
      preload="auto"
    />
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount, watch } from 'vue'
import videojs from 'video.js'
import 'video.js/dist/video-js.css'
import flvjs from 'flv.js'

const props = defineProps({
  src: {
    type: String,
    required: true,
  },
  type: {
    type: String,
    default: 'video/mp4',
  },
  poster: {
    type: String,
    default: '',
  },
  width: {
    type: [String, Number],
    default: '100%',
  },
  height: {
    type: [String, Number],
    default: 'auto',
  },
  autoplay: {
    type: Boolean,
    default: false,
  },
})

const videoElement = ref(null)
let player = null
let flvPlayer = null // FLV.js播放器实例

const initPlayer = () => {
  if (!videoElement.value) {
    console.error('Video element not found')
    return
  }

  console.log('Initializing player for type:', props.type)

  // 如果是FLV格式，直接使用flv.js
  if (props.type === 'video/x-flv') {
    initFlvPlayer()
  } else {
    // 其他格式使用Video.js
    initVideoJsPlayer()
  }
}

const initFlvPlayer = () => {
  if (!flvjs.isSupported()) {
    console.error('FLV.js is not supported in this browser')
    // 降级到Video.js尝试
    initVideoJsPlayer()
    return
  }

  console.log('Initializing FLV.js player')

  // 创建Video.js播放器（仅用于UI控制，不指定源）
  const options = {
    autoplay: false, // 禁用自动播放，等FLV加载完成
    controls: true,
    preload: 'none', // 重要：不让Video.js预加载
    fluid: true,
    responsive: true,
    poster: props.poster,
    techOrder: ['html5'],
    // 不设置sources - 由FLV.js接管
    controlBar: {
      playToggle: true,
      volumePanel: { inline: false, vertical: true },
      currentTimeDisplay: true,
      timeDivider: true,
      durationDisplay: true,
      progressControl: { seekBar: true },
      liveDisplay: false,
      remainingTimeDisplay: false,
      customControlSpacer: true,
      playbackRateMenuButton: true,
      fullscreenToggle: true,
      pictureInPictureToggle: true
    },
    userActions: {
      hotkeys: true,
      click: true,
      doubleClick: true
    }
  }

  try {
    player = videojs(videoElement.value, options, function onPlayerReady() {
      console.log('Video.js UI ready')
      
      console.log('═══ Creating FLV Player ═══')
      console.log('URL:', props.src)
      console.log('CORS:', true)
      console.log('WithCredentials:', true)
      console.log('═══════════════════════════')
      
      // 创建FLV.js播放器实例
      flvPlayer = flvjs.createPlayer({
        type: 'flv',
        url: props.src,
        isLive: false,
        cors: true,
        withCredentials: true, // 发送认证cookies
      }, {
        enableWorker: false,
        enableStashBuffer: true,
        stashInitialSize: 128,
        autoCleanupSourceBuffer: true,
      })

      // 将FLV播放器附加到video元素
      const videoEl = this.el().querySelector('video')
      flvPlayer.attachMediaElement(videoEl)
      
      // 同步FLV事件到Video.js
      flvPlayer.on(flvjs.Events.MEDIA_INFO, (info) => {
        console.log('FLV media info:', info)
      })

      flvPlayer.on(flvjs.Events.ERROR, (errorType, errorDetail, errorInfo) => {
        console.error('═══ FLV ERROR ═══')
        console.error('Error Type:', errorType)
        console.error('Error Detail:', errorDetail)
        console.error('Error Info:', errorInfo)
        console.error('═══════════════')
        
        // 详细的错误类型说明
        if (errorType === flvjs.ErrorTypes.NETWORK_ERROR) {
          console.error('→ Network Error: Check authentication, CORS, or file availability')
          if (errorDetail === flvjs.ErrorDetails.NETWORK_STATUS_CODE_INVALID) {
            console.error('  → HTTP status code error (401/403/404/etc)')
          }
        } else if (errorType === flvjs.ErrorTypes.MEDIA_ERROR) {
          console.error('→ Media Error: FLV format issue or codec not supported')
        }
        
        this.error({ code: 4, message: `FLV Error: ${errorType} - ${errorDetail}` })
      })
      
      flvPlayer.on(flvjs.Events.LOADING_COMPLETE, () => {
        console.log('✓ FLV loading complete')
      })
      
      flvPlayer.on(flvjs.Events.RECOVERED_EARLY_EOF, () => {
        console.log('✓ FLV recovered from early EOF')
      })
      
      flvPlayer.on(flvjs.Events.STATISTICS_INFO, (stats) => {
        console.log('FLV statistics:', {
          speed: stats.speed + ' KB/s',
          decodedFrames: stats.decodedFrames,
          droppedFrames: stats.droppedFrames
        })
      })

      // 加载FLV数据
      flvPlayer.load()
      console.log('FLV.js player attached and loaded')
      
      // 如果设置了autoplay，在元数据加载后自动播放
      if (props.autoplay) {
        videoEl.addEventListener('loadedmetadata', () => {
          videoEl.play().catch(err => {
            console.warn('Autoplay failed:', err)
          })
        }, { once: true })
      }
    })

    player.on('error', () => {
      const err = player.error()
      console.error('Video player error:', err)
    })

  } catch (error) {
    console.error('Failed to initialize FLV player:', error)
  }
}

const initVideoJsPlayer = () => {
  console.log('Initializing standard Video.js player')
  
  const options = {
    autoplay: props.autoplay,
    controls: true,
    preload: 'auto',
    fluid: true,
    responsive: true,
    poster: props.poster,
    techOrder: ['html5'],
    html5: {
      vhs: { overrideNative: true },
      nativeVideoTracks: false,
      nativeAudioTracks: false,
      nativeTextTracks: false
    },
    // 控制栏配置，确保进度条功能完整
    controlBar: {
      playToggle: true,
      volumePanel: {
        inline: false,
        vertical: true
      },
      currentTimeDisplay: true,
      timeDivider: true,
      durationDisplay: true,
      progressControl: {
        seekBar: true
      },
      liveDisplay: false,
      remainingTimeDisplay: false,
      customControlSpacer: true,
      playbackRateMenuButton: true,
      chaptersButton: false,
      descriptionsButton: false,
      subsCapsButton: false,
      audioTrackButton: false,
      fullscreenToggle: true,
      pictureInPictureToggle: true
    },
    // 用户操作配置，确保进度条可交互
    userActions: {
      hotkeys: true,
      click: true,
      doubleClick: true
    },
    sources: [{
      src: props.src,
      type: props.type,
    }],
  }

  console.log('Initializing with source:', { src: props.src, type: props.type })

  try {
    player = videojs(videoElement.value, options, function onPlayerReady() {
      console.log('Video player ready')
      
      // 确保进度条可拖拽
      const progressControl = this.controlBar.progressControl
      if (progressControl) {
        const seekBar = progressControl.seekBar
        if (seekBar) {
          seekBar.enable()
          console.log('SeekBar enabled for interaction')
        }
      }
    })

    player.on('error', () => {
      const err = player.error()
      console.error('Video player error:', err?.message || 'Unknown error', err)
    })

    player.on('loadedmetadata', () => {
      console.log('Video metadata loaded, duration:', player.duration())
    })

    player.on('seeking', () => {
      console.log('Seeking to:', player.currentTime())
    })

    player.on('seeked', () => {
      console.log('Seeked to:', player.currentTime())
    })

  } catch (error) {
    console.error('Failed to initialize video player:', error)
  }
}

const destroyPlayer = () => {
  // 销毁FLV播放器
  if (flvPlayer) {
    try {
      flvPlayer.pause()
      flvPlayer.unload()
      flvPlayer.detachMediaElement()
      flvPlayer.destroy()
      console.log('FLV player destroyed')
    } catch (e) {
      console.warn('Error destroying FLV player:', e)
    }
    flvPlayer = null
  }
  
  // 销毁Video.js播放器
  if (player) {
    try {
      player.dispose()
      console.log('Video.js player disposed')
    } catch (e) {
      console.warn('Error disposing player:', e)
    }
    player = null
  }
}

watch(() => props.src, (newSrc) => {
  if (!newSrc) return
  
  if (props.type === 'video/x-flv' && flvPlayer) {
    // FLV播放器：卸载并重新加载
    try {
      flvPlayer.unload()
      flvPlayer.load()
      console.log('FLV source reloaded:', newSrc)
    } catch (e) {
      console.error('Error reloading FLV source:', e)
    }
  } else if (player) {
    // Video.js播放器：更新源
    player.src({
      src: newSrc,
      type: props.type,
    })
    console.log('Video.js source updated:', newSrc)
  }
})

onMounted(() => {
  // 延迟初始化，确保DOM完全挂载
  setTimeout(() => {
    initPlayer()
  }, 100)
})

onBeforeUnmount(() => {
  destroyPlayer()
})

defineExpose({
  player,
})
</script>

<style scoped>
.video-player-wrapper {
  width: 100%;
  background: #000;
}

.video-js {
  width: 100%!important;
  height: 100%!important;
}

/* 确保进度条可见且可交互 */
:deep(.vjs-progress-control) {
  position: absolute;
  width: 100%;
  height: 30px; /* 增加点击区域 */
  bottom: 30px;
  cursor: pointer;
}

:deep(.vjs-progress-holder) {
  height: 6px; /* 进度条高度 */
  margin: 0;
  cursor: pointer;
}

/* 进度条悬停时增加高度 */
:deep(.vjs-progress-control:hover .vjs-progress-holder) {
  height: 10px;
  font-size: 1.5em; /* 增加进度球大小 */
}

/* 播放进度条 */
:deep(.vjs-play-progress) {
  background-color: #ff0000; /* 红色进度 */
  cursor: pointer;
}

/* 进度球 */
:deep(.vjs-play-progress .vjs-time-tooltip),
:deep(.vjs-play-progress::before) {
  display: block;
  cursor: pointer;
  pointer-events: auto;
}

/* 缓冲进度条 */
:deep(.vjs-load-progress) {
  background: rgba(255, 255, 255, 0.3);
}

/* Seek条 */
:deep(.vjs-seek-to-live-control),
:deep(.vjs-progress-control .vjs-mouse-display) {
  cursor: pointer;
  pointer-events: auto;
}

/* 确保所有控制元素可交互 */
:deep(.vjs-control-bar) {
  display: flex !important;
  pointer-events: auto !important;
}

:deep(.vjs-control) {
  pointer-events: auto !important;
}

/* 修复拖拽时的光标 */
:deep(.vjs-progress-control .vjs-play-progress),
:deep(.vjs-progress-control .vjs-progress-holder) {
  cursor: pointer !important;
  pointer-events: auto !important;
}

/* 响应式设计 */
@media (max-width: 768px) {
  :deep(.vjs-progress-holder) {
    height: 8px;
  }
  
  :deep(.vjs-progress-control) {
    height: 40px; /* 移动设备上增大触摸区域 */
  }
}
</style>
