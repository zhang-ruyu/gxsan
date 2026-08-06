import { createApp } from 'vue'
import App from './App.vue'
import router from './router'
import './styles/main.css'

// 全局错误兜底：任何渲染/运行时异常不再静默白屏，而是在页面上显示可读错误浮层，
// 便于定位真实报错（白屏根因往往就是某个未捕获异常把整页渲染搞崩）。
function showErrorOverlay(err) {
  const msg = (err && (err.stack || err.message)) || String(err)
  let el = document.getElementById('global-error-overlay')
  if (!el) {
    el = document.createElement('div')
    el.id = 'global-error-overlay'
    el.style.cssText =
      'position:fixed;inset:0;z-index:99999;background:#1e1e1e;color:#ff6b6b;' +
      'font-family:monospace;font-size:13px;line-height:1.6;padding:24px;' +
      'overflow:auto;white-space:pre-wrap;'
    document.body.appendChild(el)
  }
  el.textContent = '⚠️ 应用发生错误（请把这段截图发给开发者）：\n\n' + msg
}

const app = createApp(App)

app.config.errorHandler = (err, instance, info) => {
  console.error('[Vue error]', err, info)
  showErrorOverlay(err)
}

app.use(router)
app.mount('#app')

// 兜底：未被 Vue 捕获的 Promise 异常 / 全局错误也显示出来，避免直接白屏
window.addEventListener('unhandledrejection', (e) => {
  console.error('[unhandledrejection]', e.reason)
  showErrorOverlay(e.reason)
})
window.onerror = (msg, src, line, col, err) => {
  console.error('[window.onerror]', msg, src, line, col, err)
  showErrorOverlay(err || msg)
  return false
}
