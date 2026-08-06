<template>
  <div class="title-bar">
    <div class="title-bar__controls">
      <button class="win-btn" title="最小化" @click="minimize" aria-label="最小化">
        <svg viewBox="0 0 24 24" width="13" height="13">
          <line x1="5" y1="12" x2="19" y2="12" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        </svg>
      </button>
      <button class="win-btn" :title="isFullscreen ? '退出全屏' : '全屏'" @click="toggleFullscreen" aria-label="全屏">
        <svg viewBox="0 0 24 24" width="13" height="13" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
          <template v-if="!isFullscreen">
            <path d="M4 9V4h5" />
            <path d="M20 9V4h-5" />
            <path d="M4 15v5h5" />
            <path d="M20 15v5h-5" />
          </template>
          <template v-else>
            <path d="M9 4v5H4" />
            <path d="M15 4v5h5" />
            <path d="M9 20v-5H4" />
            <path d="M15 20v-5h5" />
          </template>
        </svg>
      </button>
      <button class="win-btn win-btn--close" title="关闭" @click="close" aria-label="关闭">
        <svg viewBox="0 0 24 24" width="13" height="13">
          <line x1="6" y1="6" x2="18" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <line x1="18" y1="6" x2="6" y2="18" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
        </svg>
      </button>
    </div>
    <div class="title-bar__title">股息三</div>
  </div>
</template>

<script>
import { WindowMinimise, WindowToggleFullscreen, WindowClose } from '../../wailsjs/go/main/App'

export default {
  name: 'TitleBar',
  data() {
    return {
      isFullscreen: false
    }
  },
  methods: {
    async minimize() {
      try {
        await WindowMinimise()
      } catch (e) {
        console.error('最小化失败', e)
      }
    },
    async toggleFullscreen() {
      try {
        await WindowToggleFullscreen()
        this.isFullscreen = !this.isFullscreen
      } catch (e) {
        console.error('切换全屏失败', e)
      }
    },
    async close() {
      try {
        await WindowClose()
      } catch (e) {
        console.error('关闭失败', e)
      }
    }
  }
}
</script>

<style scoped>
.title-bar {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  height: var(--titlebar-height, 38px);
  z-index: 1000;
  display: flex;
  align-items: center;
  padding: 0 8px;
  background: var(--card-bg, #ffffff);
  border-bottom: 1px solid var(--border-color, #e0e0e0);
  user-select: none;
  /* Wails v2 拖拽区域：整个标题栏可拖动窗口 */
  --wails-draggable: drag;
  -webkit-app-region: drag;
}

.title-bar__controls {
  display: flex;
  gap: 2px;
  /* 按钮区不可触发拖拽，保证点击生效 */
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

.title-bar__title {
  position: absolute;
  left: 50%;
  transform: translateX(-50%);
  font-size: 13px;
  font-weight: 600;
  color: var(--text-secondary, #5f6368);
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
  pointer-events: none;
}

.win-btn {
  width: 34px;
  height: 28px;
  border: none;
  background: transparent;
  color: var(--text-secondary, #5f6368);
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: background 0.15s, color 0.15s;
  --wails-draggable: no-drag;
  -webkit-app-region: no-drag;
}

.win-btn:hover {
  background: rgba(0, 0, 0, 0.07);
  color: var(--text-primary, #202124);
}

.win-btn--close:hover {
  background: #ea4335;
  color: #ffffff;
}
</style>
