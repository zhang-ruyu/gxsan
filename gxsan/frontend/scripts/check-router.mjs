#!/usr/bin/env node
/**
 * 构建前置校验：确保 router 中引用的每个组件都已 import。
 *
 * 背景：vite/rollup 遇到未定义的标识符时，会当作「全局变量」原样打包，
 * 构建阶段不报错也不警告，但运行时抛 ReferenceError，
 * 导致 router 创建失败 → 整个 Vue 应用挂载失败 → 白屏。
 * 这类问题在无显示器的环境下无法通过「进程是否存活」发现，必须静态拦截。
 */
import { readFileSync, existsSync } from 'node:fs'
import { resolve, dirname } from 'node:path'
import { fileURLToPath } from 'node:url'

const root = resolve(dirname(fileURLToPath(import.meta.url)), '..')
const routerFile = resolve(root, 'src/router/index.js')

if (!existsSync(routerFile)) {
  console.error(`[check-router] 找不到 ${routerFile}`)
  process.exit(1)
}

const src = readFileSync(routerFile, 'utf-8')

// 收集所有 import 进来的标识符（默认导入 + 具名导入）
const imported = new Set()
for (const m of src.matchAll(/import\s+([\w$]+)\s*(?:,\s*\{([^}]*)\})?\s*from/g)) {
  imported.add(m[1])
  if (m[2]) m[2].split(',').forEach((n) => {
    const name = n.split(/\s+as\s+/).pop().trim()
    if (name) imported.add(name)
  })
}
for (const m of src.matchAll(/import\s*\{([^}]*)\}\s*from/g)) {
  m[1].split(',').forEach((n) => {
    const name = n.split(/\s+as\s+/).pop().trim()
    if (name) imported.add(name)
  })
}

// 收集本文件内声明的标识符（支持懒加载写法 const X = () => import(...)）
for (const m of src.matchAll(/(?:const|let|var|function|class)\s+([\w$]+)/g)) {
  imported.add(m[1])
}

// 收集 routes 中引用的 component
const missing = []
for (const m of src.matchAll(/component\s*:\s*([\w$]+)/g)) {
  const name = m[1]
  if (!imported.has(name)) missing.push(name)
}

// 校验每个 import 的 .vue 文件确实存在
const brokenPaths = []
for (const m of src.matchAll(/import\s+[\w$]+\s+from\s+['"](\.[^'"]+\.vue)['"]/g)) {
  const p = resolve(dirname(routerFile), m[1])
  if (!existsSync(p)) brokenPaths.push(m[1])
}

let failed = false

if (missing.length) {
  failed = true
  console.error('\n[check-router] ✗ 以下路由组件被引用但未 import，运行时会白屏：')
  for (const n of [...new Set(missing)]) console.error(`    - ${n}`)
  console.error('  修复：在 src/router/index.js 顶部补上对应的 import 语句。')
}

if (brokenPaths.length) {
  failed = true
  console.error('\n[check-router] ✗ 以下 import 指向的文件不存在：')
  for (const p of brokenPaths) console.error(`    - ${p}`)
}

if (failed) {
  console.error('')
  process.exit(1)
}

console.log('[check-router] ✓ 路由组件引用完整')
