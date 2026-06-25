/// <reference types="vite/client" />

declare module '*.vue' {
  import type { DefineComponent } from 'vue'
  const component: DefineComponent<{}, {}, any>
  export default component
}

// Vite define 编译期平台标识
declare const __BUILD_IOS__: boolean
declare const __BUILD_DESKTOP__: boolean
