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

const initPlayer = () => {
  if (!videoElement.value) {
    console.error('Video element not found')
    return
  }

  // 简化配置：只使用html5 tech，避免FLV插件问题
  // 对于FLV文件，让浏览器尝试原生播放或转码
  const options = {
    autoplay: props.autoplay,
    controls: true,
    preload: 'auto',
    fluid: true,
    poster: props.poster,
    techOrder: ['html5'],
    html5: {
      vhs: {
        overrideNative: true
      },
      nativeVideoTracks: false,
      nativeAudioTracks: false,
      nativeTextTracks: false
    },
    sources: [
      {
        src: props.src,
        type: props.type === 'video/x-flv' ? 'video/mp4' : props.type, // FLV fallback to MP4
      },
    ],
  }

  try {
    player = videojs(videoElement.value, options, function onPlayerReady() {
      console.log('Video player ready')
    })

    player.on('error', () => {
      const err = player.error()
      console.error('Video player error:', err?.message || 'Unknown error')
    })

    player.on('loadedmetadata', () => {
      console.log('Video metadata loaded')
    })

  } catch (error) {
    console.error('Failed to initialize video player:', error)
  }
}

const destroyPlayer = () => {
  if (player) {
    try {
      player.dispose()
    } catch (e) {
      console.warn('Error disposing player:', e)
    }
    player = null
  }
}

watch(() => props.src, (newSrc) => {
  if (player && newSrc) {
    player.src({
      src: newSrc,
      type: props.type === 'video/x-flv' ? 'video/mp4' : props.type,
    })
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
  width: 100%;
  height: 100%;
}
</style>
