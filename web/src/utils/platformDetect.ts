// ==========================================================================
// 平台识别工具：运行时判断当前环境是否 iOS APP
//
// 检测策略（多层兜底，适配 Capacitor 远程加载模式）：
// 1. window.Capacitor.getPlatform() → Capacitor 全局对象
// 2. window.webkit?.messageHandlers → WKWebView 原生桥
// 3. User-Agent 含 iPhone/iPad + __capacitor__ → iOS 设备兜底
// ==========================================================================

// iOS 设备 User-Agent 检测（排除 Mac 桌面 Safari）
const isIOSUserAgent = (): boolean => {
  if (typeof navigator === 'undefined') return false
  const ua = navigator.userAgent
  return /iPhone|iPad|iPod/.test(ua) && !/Macintosh/.test(ua)
}

// 判断环境：是否 iOS Capacitor APP
export const isIOSApp = (): boolean => {
  const cap = (window as any).Capacitor
  if (cap?.getPlatform?.() === 'ios') return true
  if ((window as any).webkit?.messageHandlers) return true
  if (isIOSUserAgent() && (window as any).__capacitor__) return true
  return false
}
