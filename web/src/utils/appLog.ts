/**
 * 统一前端日志系统：始终输出到浏览器控制台，
 * 在 Android WebView 中通过 AndroidNative.log() 桥接回传到服务端。
 * 在 iOS WKWebView 中静默忽略（只检查 AndroidNative）。
 *
 * Tag 约定：使用简短 PascalCase 模块名，如
 *   'ClawBench', 'ChatStream', 'PortForward', 'Store', 'useDialog'
 *
 * Level 映射：
 *   appLog.d → console.log  / AppLog.d (D)
 *   appLog.i → console.info / AppLog.i (I)
 *   appLog.w → console.warn / AppLog.w (W)
 *   appLog.e → console.error/ AppLog.e (E)
 */

function safeStringify(a: unknown): string {
  if (typeof a === 'string') return a
  if (typeof a === 'number' || typeof a === 'boolean') return String(a)
  try { return JSON.stringify(a) } catch { return String(a) }
}

function relay(level: string, tag: string, args: unknown[]): void {
  try {
    const native = (window as any).AndroidNative
    if (!native || !native.log) return
    if (native.isNativeApp?.() !== true) return
    if (window !== window.top) return
    const msg = args.map(safeStringify).join(' ')
    native.log(level, tag, msg)
  } catch {
    // bridge 不可用时静默忽略
  }
}

export const appLog = {
  d(tag: string, ...args: unknown[]) { console.log(`[${tag}]`, ...args); relay('D', tag, args) },
  i(tag: string, ...args: unknown[]) { console.info(`[${tag}]`, ...args); relay('I', tag, args) },
  w(tag: string, ...args: unknown[]) { console.warn(`[${tag}]`, ...args); relay('W', tag, args) },
  e(tag: string, ...args: unknown[]) { console.error(`[${tag}]`, ...args); relay('E', tag, args) },
}
