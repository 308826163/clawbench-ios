# ClawBench iOS 移植项目 — Claude Code 行为规范

> **补充资料：** 详细架构文档见 `C:\Users\23611\.claude\projects\C--Users-23611\memory\clawbench-ios-architecture.md`（含完整文件清单、构建部署流程、GitHub Actions 细节）。

> **本项目禁止 Claude Code 自主设计架构。任何未在方案中明确要求的重构、替换、依赖新增、目录调整，一律视为违规操作并立即停止执行。**

> **⚠️ 默认平台规则：** 所有 UI 修改默认为 **iOS 专属**，必须用 `isIOSApp()` / `.ios-app` / `__BUILD_IOS__` 包裹隔离，不得影响桌面端。用户明确说"两个平台都改"时才可共用。

---

## 第零原则（最高优先级）

Claude Code 是执行者，不是架构设计者。

### Claude 禁止

- ❌ 自行修改架构
- ❌ 自行新增依赖
- ❌ 自行移动目录结构
- ❌ 自行重命名模块
- ❌ 自行替换技术方案
- ❌ 自行推断需求

### Claude 允许

- ✅ 严格按照方案执行
- ✅ 提出风险提醒
- ✅ 提出备选方案
- ✅ 等待人工确认

---

## 执行优先级

1. **第一优先级** — 遵守方案
2. **第二优先级** — 遵守现有代码结构
3. **第三优先级** — 最小修改原则
4. **第四优先级** — 保证可回滚

---

## 架构冻结原则

以下目录视为冻结：

```
internal/
cmd/
config/
web/src/
web/css/
assets/
```

### Claude 禁止

- 重命名
- 迁移目录
- 拆分模块
- 改动业务逻辑

### Claude 允许

- 新增文件
- 新增桥接接口
- 新增适配代码

---

## Vue 层约束

### Vue 负责

- 页面渲染
- 状态管理
- 业务逻辑
- API 调用

### Vue 禁止

- ❌ 重写页面
- ❌ 重构 Store
- ❌ 重写 Router
- ❌ 修改 Agent 实现

### 原则

原页面能运行，就不要重写。

---

## SwiftUI 层约束

### Swift 只允许接管

- Navigation
- TabBar
- Sheet
- Haptic
- Share
- FilePicker
- Keyboard
- ContextMenu

### Swift 禁止

- ❌ 承担业务逻辑
- ❌ 调用业务 API
- ❌ 保存业务状态

### 原则

SwiftUI 是交互容器，不是业务容器。

---

## 插件约束

每个插件必须：

- 一个插件
- 一个能力
- 一个职责

### 正确示例

```
HapticPlugin
NavigationPlugin
KeyboardPlugin
```

### 禁止示例

```
SystemPlugin
UtilityPlugin
AppManager
```

禁止万能插件。

---

## 修改文件数量限制

- **普通任务：** ≤ 3 个文件
- **复杂任务：** ≤ 8 个文件

### 超过限制时必须停止执行，输出：

```
预计修改文件：12个
超出限制。
原因：xxx
建议：xxx
等待确认。
```

---

## 禁止隐式依赖

禁止以下操作，除非方案明确要求：

```
npm install xxx
pod install xxx
swift package add xxx
```

### 任何新增依赖必须说明

- 目的
- 收益
- 风险
- 替代方案

等待确认。

---

## 禁止猜测

如果信息不足，必须输出：

```
缺少信息：
1. xxx
2. xxx
3. xxx

存在多个实现方案：
方案A
方案B

请确认后执行。
```

### 禁止以下词汇

- Assuming...
- Probably...
- I think...

禁止直接开始编码。

---

## 修改流程（强制）

### Phase 1 — 分析

- 阅读代码
- 输出影响范围

### Phase 2 — 设计

- 输出执行计划
- 输出修改文件清单

### Phase 3 — 等待确认

用户输入：`EXECUTE`

### Phase 4 — 开始修改

### Phase 5 — 输出报告

- 新增文件
- 修改文件
- 删除文件
- 风险
- 回滚方法

---

## Git 规范

每完成一个功能必须：

```
git add .
git commit -m "feat(ios): xxx"
```

禁止一次提交多个功能。

---

## 回滚规范

任何任务必须提供回滚命令：

```
git reset --hard HEAD~1
```

或者：

```
git revert commit-id
```

---

## 验收规范

每完成一个任务必须输出：

| 验收项 | 状态 |
|--------|------|
| 编译 | ✅ Pass |
| 功能 | ✅ Pass |
| 深色模式 | ✅ Pass |
| 横竖屏 | ✅ Pass |
| 回滚验证 | ✅ Pass |
| 修改文件 | （列出清单） |

---

## 项目信息

- **源码版本：** clawbench-0.48.0 魔改版
- **项目路径：** `D:\AI\BF\clawbench IOS\clawbench-ios\`
- **服务器地址：** `http://192.168.0.107:20003`
- **Bundle ID：** `com.gin.clawbench`
- **安装方式：** 巨魔助手（TrollStore）
- **GitHub 仓库：** https://github.com/308826163/clawbench-ios

---

## Capacitor 键盘适配方案（重要）

### 问题描述

iOS WKWebView 中，键盘弹出时默认会缩小 viewport，导致：
- `position: fixed` 的元素（如导航栏）跟随上移
- 输入栏和消息气泡位置异常

### 解决方案：Capacitor Keyboard Body 模式

在 `web/src/main.ts` 中调用：

```typescript
try {
    const cap = (window as any).Capacitor
    if (cap?.Plugins?.Keyboard) {
        cap.Plugins.Keyboard.setResizeMode({ mode: 'Body' })
    }
} catch (_) { /* 非 Capacitor 环境静默忽略 */ }
```

### 运行逻辑

```
键盘弹出
  ↓
viewport 不变（844px）
body 缩小（844px → 400px）
  ↓
导航栏：position: fixed → 相对 viewport → 位置不动 → 被键盘遮盖 ✓
聊天区域：在 body 内 → 随 body 缩小 → 内容上移 ✓
输入栏：在 body 底部 → 随 body 上移 → 在键盘上方 ✓
```

### 为什么不用其他方案

| 方案 | 问题 |
|------|------|
| `interactive-widget=overlays-content` | WKWebView 不支持 |
| Visual Viewport API 监听 | 补丁方案，非根源解决 |
| `position: absolute` + 固定高度 | 需要手动计算，复杂 |
| `Body` 模式 | ✅ 原生支持，根源解决 |

### 注意事项

1. 此代码只在 Capacitor 环境生效，浏览器中静默忽略
2. 需要 iOS 项目安装 `@capacitor/keyboard` 插件
3. 修改后需要重新打包 IPA 才能生效

---

## 开发工作流

### Web 端修改 → iOS 构建流程

1. 在本地修改 `web/src/` 下的代码
2. 构建前端：`cd web && npm run build`
3. 提交代码：`git add . && git commit -m "feat: xxx"`
4. 推送到 GitHub：`git push origin main`
5. 触发 iOS 构建：`gh workflow run "Build iOS IPA" --repo 308826163/clawbench-ios --ref main`
6. 等待构建完成，下载 IPA 安装

### 代码仓库关系

- **Web 源码仓库：** `308826163/apple`（修改 web 代码）
- **iOS 构建仓库：** `308826163/clawbench-ios`（自动拉取 web 代码打包 IPA）

---

## 平台隔离架构

### 两个独立端

```
iOS 端                              电脑端
├─ APP UI（原生壳）                  └─ 桌面浏览器访问 web UI
│  AppDelegate.swift                   （完全独立，与 iOS 端无关）
│  Info.plist
│  启动画面
│  需要重新打包 IPA 才能更新
│
└─ 服务器 UI（web/ 代码）
   从 192.168.0.107:20003 加载
   到 APP 的 WKWebView 里
   部署到服务器，APP 刷新即生效
```

**核心概念：**
- **iOS 端** = APP UI + 服务器 UI，两部分都属于 iOS 端
- **电脑端** = 桌面浏览器，完全独立，与 iOS 端无关
- **web/ 代码是共享的**，两个端都加载同一份，通过隔离机制做分支
- 改 web/ 不加隔离 → 两个端都变
- 改 web/ 加 `isIOSApp()` / `.ios-app` 隔离 → 只有 iOS 端变，电脑端不变

**两个 Capacitor 配置文件：**
- `capacitor.config.ts`（项目根目录）— 开发用，含 `server.url` 远程加载
- `ios/App/App/capacitor.config.json`（iOS 目录内）— 构建用，无 `server.url`

---

### 部署方式

| 改动位置 | 部署方式 | 生效方式 |
|----------|----------|----------|
| `web/` 代码（服务器 UI） | 编译后部署到 Go 后端服务器 | iOS APP 刷新即生效，电脑端刷新即生效 |
| `ios/` 原生层（APP UI） | 重新打包 IPA，安装到手机 | 需要重新安装 |
| `internal/` 后端 | 重新编译 clawbench.exe，重启服务 | 两个端都受影响 |

---

### 平台检测机制（三层兜底）

文件：`web/src/utils/platformDetect.ts`

```
层1: window.Capacitor.getPlatform() === 'ios'   ← Capacitor 全局对象
层2: window.webkit?.messageHandlers              ← WKWebView 原生桥
层3: User-Agent 含 iPhone/iPad + __capacitor__   ← 兜底
```

导出函数：
- `isIOSApp()` — 是否 iOS 原生 APP
- `isDesktopWeb()` — 是否电脑浏览器

编译期标识（Vite define，tree-shake 死代码消除）：
- `__BUILD_IOS__` / `__BUILD_DESKTOP__` — 声明在 `web/src/env.d.ts`

---

### 编译期隔离

| 环境 | .env.ios | .env.desktop | dev 模式 |
|------|----------|--------------|----------|
| `VITE_BUILD_PLATFORM` | `ios` | `desktop` | 无（默认 all） |
| `__BUILD_IOS__` | true | false | true |
| `__BUILD_DESKTOP__` | false | true | true |

构建命令：
- `npm run build:ios` → 只打包 iOS 代码
- `npm run build:desktop` → 只打包桌面代码
- `npm run build` → 全部打包（dev 用）

---

### 运行时隔离（App.vue）

#### iOS 专属逻辑（`isIOSApp()` 为 true）

| 位置 | 代码 | 作用 |
|------|------|------|
| 模板 | `v-if="isIOSApp()"` SVG | Gooey Filter 液态玻璃滤镜 |
| onMounted | `document.body.classList.add('ios-app')` | CSS 类名 |
| updateStatusBarStyle | `import('@capacitor/status-bar')` | 主题切换同步状态栏 |
| useChatKeyboard | visualViewport 监听 | WKWebView 无 adjustResize，JS 检测键盘高度 |

#### 桌面专属逻辑（`isDesktopWeb()` 为 true）

| 位置 | 代码 | 作用 |
|------|------|------|
| onMounted | `document.body.classList.add('desktop-web')` | CSS 类名 |
| handleOpenSession | `clawbench-open-session` 事件 | Android 推送跳转会话 |
| handleOpenTask | `clawbench-open-task` 事件 | Android 推送跳转任务 |
| onMounted | `window.AndroidNative?.getPassword()` | Android 密码自动登录 |
| onMounted | `window.AndroidNative?.getPendingNavigation()` | Android 冷启动导航恢复 |

---

### iOS 原生层（AppDelegate.swift）

```
1. 缓存 WKWebView 引用
2. 禁用弹性滚动（bounces = false）
3. 禁用自动内容插入（contentInsetAdjustmentBehavior = .never）
4. 监听键盘弹出/收起，手动调整 scrollView.contentInset（0.28s 动画）
```

---

### 文件分类清单

#### iOS 专属文件

| 文件 | 作用 |
|------|------|
| `ios/App/App/AppDelegate.swift` | 原生壳入口 |
| `ios/App/App/Info.plist` | Bundle ID、HTTP 允许、屏幕方向 |
| `ios/App/App/Base.lproj/LaunchScreen.storyboard` | 启动画面 |
| `ios/App/App/Assets.xcassets/` | App 图标、Splash 图片 |
| `ios/App/App/capacitor.config.json` | iOS 构建时 Capacitor 配置 |
| `ios/App/CapApp-SPM/` | Swift Package Manager 依赖 |
| `.env.ios` | iOS 构建环境变量 |
| `.github/workflows/build-ios.yml` | GitHub Actions iOS 构建 |

#### 共享文件（含平台分支逻辑）

| 文件 | 平台分支点 |
|------|-----------|
| `web/src/utils/platformDetect.ts` | 检测函数本身 |
| `web/src/App.vue` | `isIOSApp()` → Gooey / StatusBar / 键盘；`isDesktopWeb()` → Android 桥接 |
| `web/src/components/chat/ChatInputBar.vue` | `@capacitor/keyboard` dismissKeyboard |
| `web/src/composables/useChatKeyboard.ts` | visualViewport 键盘检测 |
| `web/src/composables/useTerminalKeyboard.ts` | 终端键盘高度检测 |
| `web/src/env.d.ts` | `__BUILD_IOS__` / `__BUILD_DESKTOP__` 类型声明 |
| `capacitor.config.ts` | Capacitor 插件配置 |
| `vite.config.ts` | 编译期 define 配置 |

---

### 修改影响范围判断

| 改动类型 | 影响范围 | 测试要求 | 部署方式 |
|----------|----------|----------|----------|
| 改 web/ 不加隔离 | iOS 端 + 电脑端 | 两个端都要测 | 部署到服务器 |
| 改 web/ 加 `isIOSApp()` 隔离 | 只 iOS 端 | 测 iOS 端 | 部署到服务器 |
| 改 web/ 加 `isDesktopWeb()` 隔离 | 只电脑端 | 测电脑端 | 部署到服务器 |
| 改 `.swift` / `.plist` | 只 iOS APP | 测 iOS APP | 重新打包 IPA |
| 改 `internal/` / `cmd/` | 两个端 | 两个端都要测 | 重启服务 |

**默认规则：所有 UI 修改默认 iOS 专属，用 `isIOSApp()` / `.ios-app` / `__BUILD_IOS__` 隔离，不影响电脑端。**

---

## 本次修改记录（2026-06-25）

### 修改内容

| 文件 | 改动 | 说明 |
|------|------|------|
| `App.vue` | 导航栏 padding-bottom: 12px → 27px | 导航栏上移 15px |
| `App.vue` | FloatingInputBar 禁用，恢复原版 ChatInputBar | 去掉收起/拖拽功能 |
| `ChatInputBar.vue` | `position: fixed` → `position: relative` | 输入栏改为文档流，配合键盘联动 |
| `ChatInputBar.vue` | 键盘联动用 Capacitor 原生事件 | `keyboardWillShow` + `keyboardWillHide` |
| `ChatInputBar.vue` | transform 动画参数 | 60ms 延迟 + 0.32s cubic-bezier(0.33,1,0.68,1) |
| `ChatPanelContent.vue` | 新增 `.chat-messages-scroll` 滚动容器 | 消息列表独立滚动，输入栏钉在底部 |
| `ChatPanelContent.vue` | 输入栏 `position: relative; bottom: 90px` | 拉回导航栏上方 |
| `BottomSheet.vue` | `.bs-panel` 加 `bottom: env(safe-area-inset-bottom)` | 弹窗底部避开 home indicator |
| `BottomSheet.vue` | 全屏模式加 `top: env(safe-area-inset-top)` | 弹窗顶部避开状态栏 |
| `LiquidTabBar.vue` | 普通图标颜色 `rgba(255,255,255,0.8)` + 透明度 0.65 | 深色模式图标调亮 |
| `LiquidTabBar.vue` | 指示器图标保持 `#0A84FF` | Apple 蓝不变 |
| `base.css` | `#app` 去掉 `padding-top: env(safe-area-inset-top)` | 消除重复安全区域间距 |

### 经验教训

#### 1. fixed 定位组件不能用 margin/padding 移动

输入栏用 `position: fixed; bottom: 95px;`。改 `margin-bottom` 或 `padding-bottom` 对它无效。
- ✅ 正确：改 `bottom` 值，或用 `transform: translateY()`
- ❌ 错误：改 `margin-bottom`、`padding-bottom`

#### 2. Capacitor 键盘事件比 visualViewport 更可靠

`visualViewport` 在 Capacitor 环境下计算键盘高度可能不准。用 Capacitor 原生事件最可靠：
```js
import { Keyboard } from '@capacitor/keyboard'
Keyboard.addListener('keyboardWillShow', (info) => { info.keyboardHeight })
Keyboard.addListener('keyboardWillHide', () => { ... })
```

#### 3. 键盘弹出/收起用不同事件

- 弹出：`keyboardWillShow`（键盘开始出现前触发，输入栏提前到位）
- 收起：`keyboardWillHide`（键盘开始收起时触发，跟手回落）
- ❌ 用 `keyboardDidHide` 会有延迟（键盘收完才触发）

#### 4. 桥接延迟需要补偿

Capacitor 桥接从原生到 JS 有 ~50-60ms 延迟。键盘弹出时需要 `setTimeout(60ms)` 补偿，否则输入栏比键盘先到。

#### 5. CSS transition 时长要匹配 iOS 键盘

- iOS 键盘动画：250ms，cubic-bezier(0.33, 1, 0.68, 1)（ease-out）
- 输入栏动画：320ms（比键盘长，覆盖桥接延迟）

#### 6. CSS transform 不触发布局，光标会滞后

`transform: translateY()` 移动的是视觉位置，浏览器光标基于布局位置渲染。动画过程中光标会短暂滞后。这是 CSS transform 的固有限制。

#### 7. 三层 overflow: hidden 裁剪链

布局中有三层裁剪，去掉任何一层都不够：
```
.app-container (overflow: hidden)
  .tab-panel (overflow: hidden)
    .chat-panel-content (overflow: hidden)
```
要让内容溢出到导航栏区域，需要同时去掉三层。但去掉后滚动容器的内容仍然被自身 `overflow-y: auto` 裁剪，无法真正溢出。

#### 8. 滚动容器的 padding 不会溢出自身边界

`overflow-y: auto` 的滚动容器，其 `padding-bottom` 在容器内部，不会超出容器边界。即使外层没有 `overflow: hidden`，滚动内容也不会溢出到导航栏区域。

#### 9. 导航栏后面的"背景墙"来源

导航栏（LiquidTabBar）背景是透明玻璃效果。透出的是 **TabPanel 的 `background: var(--bg-secondary)` 实体色**，不是 body 背景。

### 服务器部署步骤

```bash
cd "D:\AI\BF\clawbench IOS\clawbench-ios"
npm run build                    # 编译 web 代码到 public/
taskkill /F /IM clawbench.exe    # 杀掉旧进程
./clawbench.exe &                # 后台启动服务器
tasklist | grep -i claw          # 确认进程存在
```
