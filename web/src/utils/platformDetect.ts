// ==========================================================================
// 平台识别工具：运行时判断当前环境是 iOS APP 还是电脑浏览器
//
// 检测策略（多层兜底，适配 Capacitor 远程加载模式）：
// 1. window.Capacitor.getPlatform() → Capacitor 全局对象
// 2. window.webkit?.messageHandlers → WKWebView 原生桥
// 3. User-Agent 含 iPhone/iPad + 非桌面浏览器 → iOS 设备兜底
// ==========================================================================

// iOS 设备 User-Agent 检测（排除 Mac 桌面 Safari）
const isIOSUserAgent = (): boolean => {
  if (typeof navigator === 'undefined') return false
  const ua = navigator.userAgent
  // 必须含 iPhone/iPad/iPod，且不含 Macintosh（排除桌面 Safari 的 iPad 模式）
  return /iPhone|iPad|iPod/.test(ua) && !/Macintosh/.test(ua)
}

// 判断环境：是否 iOS Capacitor APP
export const isIOSApp = (): boolean => {
  // 层1：Capacitor 全局对象
  const cap = (window as any).Capacitor
  if (cap?.getPlatform?.() === 'ios') return true
  // 层2：WKWebView 原生桥（iOS 独有）
  if ((window as any).webkit?.messageHandlers) return true
  // 层3：User-Agent 兜底（iOS 设备 + 非普通 Safari = WebView/APP）
  if (isIOSUserAgent() && (window as any).__capacitor__) return true
  // 层4：纯 iOS 设备检测（最后兜底，可能误判普通 Safari）
  // 注意：此层会把 iOS Safari 也识别为 APP，按需启用
  // if (isIOSUserAgent()) return true
  return false
}

// 判断环境：是否电脑浏览器网页（非原生平台 = 电脑浏览器）
export const isDesktopWeb = (): boolean => {
  if (isIOSApp()) return false
  const cap = (window as any).Capacitor
  if (cap?.isNativePlatform?.()) return false
  return true
}

// ── 编译期平台标识（Vite define 替换，用于 tree-shake 死代码消除） ──
declare const __BUILD_IOS__: boolean
declare const __BUILD_DESKTOP__: boolean

export const BUILD_IOS: boolean = typeof __BUILD_IOS__ !== 'undefined' ? __BUILD_IOS__ : true
export const BUILD_DESKTOP: boolean = typeof __BUILD_DESKTOP__ !== 'undefined' ? __BUILD_DESKTOP__ : true
