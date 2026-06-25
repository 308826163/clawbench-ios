import type { CapacitorConfig } from '@capacitor/cli';

const config: CapacitorConfig = {
  appId: 'com.gin.clawbench',
  appName: 'ClawBench',
  webDir: 'public',
  server: {
    // 开发时连接到本地服务器
    url: 'http://192.168.0.107:20003',
    cleartext: true, // 允许 HTTP 连接
  },
  ios: {
    // iOS 特定配置
    contentInset: 'never', // 根除系统自动底部安全留白，解决上滑黑边
    backgroundColor: '#000000', // 背景色
    preferredContentMode: 'mobile', // 移动端内容模式
    scrollEnabled: true, // 启用滚动
  },
  plugins: {
    // 插件配置
    StatusBar: {
      style: 'AUTO', // 自动根据主题切换深浅模式
      backgroundColor: '#000000', // 状态栏背景色
    },
    Keyboard: {
      resize: 'body', // 只挤压页面内部内容，网页整体高度不变，底部导航永久钉死底部
      style: 'DARK', // 深色键盘
      animationDuration: 280, // 匹配iOS原生键盘250~300ms动画时长，消除闪烁断层
      resizeDelay: 0, // 无延迟触发视口resize
    },
    // 新增：开启JS与原生Swift通信，实现Tab双向联动
    WebView: {
      allowWebViewInspection: true
    },
    KeyboardPlugin: {},
  },
};

export default config;
// trigger build
