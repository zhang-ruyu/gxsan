// 统一的后端 JSON 响应解析。
//
// 错误契约（见 backend app.go）：
//   查询类接口已统一为「成功返回 JSON 字符串、失败 reject promise」，
//   所以正常路径下本函数只负责 JSON.parse；
//   但为防止任何残留的 {"error": "..."} 字符串混入，解析后若含 error 字段
//   则主动抛错，交由调用方的 catch 统一处理，避免脏数据进入视图。
export function parseJSON(str) {
  if (str && typeof str === 'object') return str
  const data = JSON.parse(str)
  if (data && typeof data === 'object' && Object.prototype.hasOwnProperty.call(data, 'error')) {
    throw new Error(typeof data.error === 'string' ? data.error : '未知错误')
  }
  return data
}
