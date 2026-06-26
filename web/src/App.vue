<template>
  <div>
    <!-- SVG Gooey Filter for Liquid Glass effect — iOS APP 专属 -->
    <svg v-if="isIOSApp()" style="position: absolute; width: 0; height: 0;" aria-hidden="true">
      <defs>
        <filter id="gooey">
          <feGaussianBlur in="SourceGraphic" stdDeviation="10" result="blur" />
          <feColorMatrix in="blur" mode="matrix" values="1 0 0 0 0  0 1 0 0 0  0 0 1 0 0  0 0 0 18 -7" result="gooey" />
          <feComposite in="SourceGraphic" in2="gooey" operator="atop" />
        </filter>
      </defs>
    </svg>

    <!-- Loading state: show nothing while checking auth -->
    <div v-if="isAuthenticated === null" style="display:none" />

    <!-- Login -->
    <LoginView v-else-if="!isAuthenticated" @login-success="handleLoginSuccess" />

    <!-- Setup wizard (after login, before main UI) -->
    <SetupWizard v-else-if="needsSetup" @complete="handleSetupComplete" />

    <!-- Main app -->
    <div v-else class="app-container" :class="{ 'chrome-hidden': terminalActive, 'chat-keyboard-open': chatKeyboardActive, 'project-switching': switchingProject }" :key="projectKey">
      <AppHeader
        :hidden="terminalActive"
        :project-root="projectRoot"
        :home-dir="homeDir"
        @open-project-dialog="handleOpenProjectDialog"
      />

      <main class="main-content">
        <div class="content-area" id="contentArea">
          <!-- Chat Tab -->
          <TabPanel tabId="chat" :activeTab="activeTab" :direction="tabDirection">
            <template #header>
              <span class="bs-header-title">{{ sessionIdentity.agentHeaderTitle.value }}</span>
              <div v-if="sessionIdentity.currentSessionTitle.value" class="bs-header-description">
                <HeaderMarquee :text="sessionIdentity.currentSessionTitle.value">{{ sessionIdentity.currentSessionTitle.value }}</HeaderMarquee>
              </div>
              <!-- ACP 按钮：从输入栏移至 header 行靠右 -->
              <button v-if="showAcpResume" class="header-acp-btn" @click.stop="chatPanelRef?.openAcpDrawer()" :title="t('chat.acpSession.title')">
                <svg xmlns="http://www.w3.org/2000/svg" width="14" height="14" viewBox="0 0 24 24" fill="currentColor" class="icon icon-tabler icons-tabler-filled icon-tabler-chevron-down"><path stroke="none" d="M0 0h24v24H0z" fill="none" /><path d="M18.707 8.293a1 1 0 0 1 0 1.414l-6 6a1 1 0 0 1 -1.414 0l-6 -6a1 1 0 0 1 1.414 -1.414l5.293 5.293l5.293 -5.293a1 1 0 0 1 1.414 0" /></svg>
              </button>
            </template>
            <ChatPanelContent
              ref="chatPanelRef"
              :active="activeTab === 'chat'"
              :current-file="currentFile"
              :current-dir="currentDir"
              :acpInHeader="true"
              @open="switchTab('chat')"
              @open-file="handleSelectFile"
              @task-card-click="onTaskCardClick"
            />
          </TabPanel>

          <!-- File Browse Tab (合一：目录浏览 + 文件覆盖预览) -->
          <TabPanel tabId="browse" :activeTab="activeTab" :noHeader="true" :direction="tabDirection">
            <div class="browse-panel">
              <FileManagerContent
                ref="fileManagerRef"
                :entries="dirEntries"
                :current-dir="currentDir"
                :current-file="currentFile"
                :show-hidden="showHidden"
                :sort-field="sortField"
                :sort-dir="sortDir"
                :dir-loading="store.state.dirLoading"
                @navigate-dir="handleNavigateDir"
                @select-file="handleBrowseSelectFile"
                @toggle-sort="handleToggleSort"
                @toggle-hidden="toggleHidden"
                @rename="handleRename"
                @delete="handleDelete"
                @batch-delete="handleBatchDelete"
                @refresh="handleRefresh"
                @open-terminal="handleOpenTerminal"
              />
              <FileOverlay
                ref="fileOverlayRef"
                :overlay-open="fileNav.overlayOpen.value"
                :current-file="currentFile"
                :toc-open="tocOpen"
                :search-open="searchOpen"
                :markdown-view-mode="markdownViewMode"
                :file-history-open="fileHistoryOpen"
                :toc-file="tocFile"
                :pdf-outline="pdfOutline"
                @delete="handleDelete($event)"
                @show-details="detailsOpen = true"
                @open-git-history="openFileHistory"
                @toggle-toc="tocOpen = !tocOpen"
                @toggle-search="currentFile?.content && (searchOpen = !searchOpen)"
                @toggle-view="markdownViewMode = markdownViewMode === 'rendered' ? 'raw' : 'rendered'"
                @refresh="handleRefresh"
                @jump="scrollToLine"
                @jump-page="handleJumpPdfPage"
                @close-git-history="fileHistoryOpen = false"
                @open-file="handleOverlayOpenFile"
                @overlay-close="handleOverlayClose"
                @overlay-go-back="handleOverlayGoBack"
              />
            </div>
          </TabPanel>

          <!-- History Tab -->
          <TabPanel tabId="history" :activeTab="activeTab" :noHeader="true" :direction="tabDirection">
            <GitHistoryContent
              mode="project"
              :active="activeTab === 'history'"
              @open-file="handleSelectFile"
            />
          </TabPanel>

          <!-- Proxy Tab -->
          <TabPanel tabId="proxy" :activeTab="activeTab" :noHeader="true" :direction="tabDirection">
            <ProxyPanelContent />
          </TabPanel>

          <!-- Terminal Tab -->
          <TabPanel tabId="terminal" :activeTab="activeTab" :noHeader="true" :direction="tabDirection">
            <TerminalPanelContent
              :requested-cwd="terminalRequestedCwd"
              :active="activeTab === 'terminal'"
              @cwd-handled="terminalRequestedCwd = null"
            />
          </TabPanel>

          <!-- Tasks Tab -->
          <TabPanel tabId="tasks" :activeTab="activeTab" :noHeader="true" :direction="tabDirection">
            <TaskTab :active="activeTab === 'tasks'" @open-file="handleTaskOpenFile" />
          </TabPanel>

          <!-- Settings Tab -->
          <TabPanel tabId="settings" :activeTab="activeTab" :noHeader="true" :direction="tabDirection">
            <SettingsPage :active="activeTab === 'settings'" />
          </TabPanel>
        </div>
      </main>

      <Lightbox />

      <ProjectDialog
        :open="projectDialogOpen"
        @close="projectDialogOpen = false"
      />

      <FileDetailsDialog
        :file="currentFile"
        :open="activeTab === 'browse' && fileNav.overlayOpen.value && detailsOpen"
        @close="detailsOpen = false"
      />

      <!-- Quote question floating bar -->
      <QuoteQuestionBar
        :visible="quoteQuestion.visible.value"
        :quoteData="quoteQuestion.quoteData.value"
        :sessionLabel="sessionIdentity.agentHeaderTitle.value"
        :sessionTitle="sessionIdentity.currentSessionTitle.value"
        :currentSessionId="sessionIdentity.currentSessionId.value"
        @send="quoteQuestion.sendMessage($event)"
        @close="quoteQuestion.closeSheet()"
        @pin="quoteQuestion.pinBar()"
        @unpin="quoteQuestion.unpinBar()"
        @open-sessions="handleQuoteOpenSessions"
      />

      <!-- Global session drawer — accessible from any tab -->
      <SessionDrawer
        ref="sessionDrawerRef"
        :open="sessionIdentity.sessionDrawerOpen.value"
        :currentSessionId="sessionIdentity.currentSessionId.value"
        :runningSessionIds="sessionIdentity.runningSessions.value"
        @close="sessionIdentity.sessionDrawerOpen.value = false"
        @select="handleSessionSelect"
        @create="handleSessionCreate"
        @delete="handleSessionDelete"
      />

      <!-- Bottom dock (tab bar) — LiquidTabBar 双层嵌套视口差位联动方案 -->
      <div v-if="isAuthenticated" v-show="!anyKeyboardActive" class="bottom-dock-wrapper">
        <LiquidTabBar
          v-model="activeIndex"
          :tabs="tabList"
          :active-tab-id="activeTab"
          @change="handleTabChange"
        >
          <template #icon="{ tab, index, layer }">
            <!-- 前 3 个普通 tab：chat / browse / history（底座 + 水滴层均渲染图标） -->
            <template v-if="index < 3">
              <!-- 会话 tab（index 0）：运行中时显示锥形渐变旋转光环 -->
              <div v-if="index === 0" class="radar-icon-wrapper" :class="{ 'has-running': isAgentRunning }">
                <component :is="tab.icon" :size="22" />
              </div>
              <!-- Git 历史 tab（index 2）：显示工作树变更计数徽标 -->
              <div v-else-if="index === 2" class="dock-btn-wrap">
                <component :is="tab.icon" :size="22" />
                <span v-if="store.state.gitWorkingTreeChangeCount > 0 && activeTab !== 'history'" class="dock-badge dock-badge-count" :class="{ 'dock-badge-pop': historyBadgeAnim }" @animationend="historyBadgeAnim = false">{{ formatBadgeCount(store.state.gitWorkingTreeChangeCount) }}</span>
              </div>
              <component v-else :is="tab.icon" :size="22" />
            </template>

            <!-- 第 4 个 tab：动态快捷按钮（仅底座层挂载 ref 坐标尺 + 真实点击事件） -->
            <template v-else-if="index === 3">
              <div
                :ref="(el) => { if (layer === 'base') taskBtnRef = el }"
                class="original-tab-proxy-node"
                @click.stop="handleDockSlot4Click()"
              >
                <component :is="dockSlot4Icon" :size="22" />
              </div>
            </template>

            <!-- 第 5 个 tab：更多（仅底座层挂载 ref 坐标尺 + 真实点击事件） -->
            <template v-else-if="index === 4">
              <div
                :ref="(el) => { if (layer === 'base') overflowBtnRef = el }"
                class="original-tab-proxy-node"
                @click.stop="toggleOverflowMenu()"
              >
                <component :is="tab.icon || 'MoreHorizontal'" :size="22" />
                <span v-if="hasUnreadNotification" class="dock-badge-dot"></span>
              </div>
            </template>
          </template>
        </LiquidTabBar>
        <div class="dock-safe-area"></div>
      </div>
    </div>

    <!-- 悬浮输入框 — 已禁用，使用原版 ChatInputBar -->
    <Teleport to="body">
      <FloatingInputBar
        v-if="false"
        :placeholder="t('chat.input.placeholder')"
        @send="handleFloatingSend"
      />
    </Teleport>

    <Teleport to="body">
      <Transition name="dock-popup">
        <div v-if="overflowMenuOpen" class="dock-overflow-popup" :style="overflowPopupStyle" @keydown.escape="overflowMenuOpen = false">
          <button class="dock-overflow-item" :class="{ active: activeTab === 'tasks' }" @click.stop="handleOverflowSelect('tasks')">
            <CalendarClock :size="16" />
            <span>{{ t('nav.tasks') }}</span>
            <span v-if="store.state.taskUnreadCount > 0" class="dock-overflow-count">{{ store.state.taskUnreadCount }}</span>
          </button>
          <button v-if="!isSSHDisabled" class="dock-overflow-item" :class="{ active: activeTab === 'proxy' }" @click.stop="handleOverflowSelect('proxy')">
            <EthernetPort :size="16" />
            <span>{{ t('nav.portForward') }}</span>
            <span v-if="store.state.portForwardActiveCount > 0" class="dock-overflow-count">{{ store.state.portForwardActiveCount }}</span>
          </button>
          <button v-if="!isTerminalDisabled" class="dock-overflow-item" :class="{ active: activeTab === 'terminal' }" @click.stop="handleOverflowSelect('terminal')">
            <TerminalIcon :size="16" />
            <span>{{ t('terminal.title') }}</span>
            <span v-if="store.state.terminalSessionCount > 0" class="dock-overflow-count">{{ store.state.terminalSessionCount }}</span>
          </button>
          <div class="dock-overflow-divider"></div>
          <button class="dock-overflow-item" @click.stop="handleOverflowSettings">
            <Settings :size="16" />
            <span>{{ t('nav.settings') }}</span>
          </button>
        </div>
      </Transition>
    </Teleport>

    <ToastNotification :toast="toast" />
    <DialogOverlay />
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, provide, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSettingsConfig } from '@/composables/useSettingsConfig'
import { formatBadgeCount } from './utils/format.ts'
import { isIOSApp } from '@/utils/platformDetect.ts'
import { MessageSquare, FolderOpen, GitBranch, EthernetPort, Terminal as TerminalIcon, CalendarClock, MoreHorizontal, Settings, RotateCcw } from 'lucide-vue-next'
import IconMessageCircle from './components/common/IconMessageCircle.vue'
import IconFolderOpen from './components/common/IconFolderOpen.vue'
import AppHeader from './components/common/AppHeader.vue'
import TabPanel from './components/common/TabPanel.vue'
import LiquidTabBar from './components/common/LiquidTabBar.vue'
import FileOverlay from './components/file/FileOverlay.vue'
import Lightbox from './components/media/Lightbox.vue'
import ChatPanelContent from './components/chat/ChatPanelContent.vue'
import FloatingInputBar from './components/chat/FloatingInputBar.vue'
import FileManagerContent from './components/file/FileManagerContent.vue'
import GitHistoryContent from './components/git/GitHistoryContent.vue'
import ProxyPanelContent from './components/proxy/ProxyPanelContent.vue'
import TerminalPanelContent from './components/terminal/TerminalPanelContent.vue'
import ProjectDialog from './components/ProjectDialog.vue'
import LoginView from './components/LoginView.vue'
import SetupWizard from './components/setup/SetupWizard.vue'
import TocDrawer from './components/TocDrawer.vue'
import FileDetailsDialog from './components/file/FileDetailsDialog.vue'
import GitHistoryDrawer from './components/git/GitHistoryDrawer.vue'
import SearchDrawer from './components/common/SearchDrawer.vue'
import ToastNotification from './components/common/ToastNotification.vue'
import DialogOverlay from './components/common/DialogOverlay.vue'
import SessionDrawer from './components/session/SessionDrawer.vue'
import QuoteQuestionBar from './components/common/QuoteQuestionBar.vue'
import HeaderMarquee from './components/common/HeaderMarquee.vue'
import SettingsPage from './components/settings/SettingsPage.vue'
import TaskTab from '@/components/task/TaskTab.vue'
import { useQuoteQuestion } from './composables/useQuoteQuestion.ts'
import { useTaskTab, registerSwitchTab, onTaskEvent } from '@/composables/useTaskTab.ts'
import { resetAgents, agentCanResume } from '@/composables/useAgents'
import { useSessionIdentity, registerSessionDrawerRef, resetIdentity } from './composables/useSessionIdentity.ts'
import { loadSessionsOnce, resetChatSessionState } from './composables/useChatSession.ts'
import { resetTaskTabState } from './composables/useTaskTab.ts'
import { clearPlanState } from './composables/usePlanProgress.ts'
import { useToast } from './composables/useToast.ts'
import { useAppMode } from './composables/useAppMode.ts'
import { useTerminalKeyboard } from './composables/useTerminalKeyboard.ts'
import { useChatKeyboard } from './composables/useChatKeyboard.ts'
import { usePortForward } from './composables/usePortForward.ts'
import { useTerminalStatus } from './composables/useTerminalStatus.ts'
import { useFileWatch } from './composables/useFileWatch.ts'
import { useFileNavStack } from './composables/useFileNavStack'
import { refreshCurrentFile } from './composables/useFileRefresh.ts'
import { useGlobalEvents } from './composables/useGlobalEvents'
import { useEdgeSwipeBack, useFeatureBackHandler } from './composables/useEdgeSwipeBack'
import { handleBackNavigation } from './composables/useBackHandler'
import { store } from './stores/app.ts'
import { setPendingCommitNavigation } from './composables/useCommitNavigation.ts'
import { initMermaid, reRenderMermaid } from './utils/mermaid.ts'
import { getFileType } from './utils/fileType.ts'
import 'highlight.js/styles/github.css'
import 'highlight.js/styles/github-dark.css'
import './assets/hljs-light-override.css'

const isAuthenticated = ref(null)
const needsSetup = ref(false)
const { t } = useI18n()

// SPA hot project switch: key forces Vue to destroy/rebuild the app-container subtree
const projectKey = ref('initial')
const switchingProject = ref(false)

async function hotSwitchProject(newProjectPath, pendingSessionId) {
  // ── Phase 1: Fade out ──
  switchingProject.value = true
  await nextTick()
  await new Promise(r => setTimeout(r, 150))

  // ── Phase 2: POST to backend to set new project cookie ──
  try {
    await store.setProject(newProjectPath)
  } catch (err) {
    // Project doesn't exist — revert fade-out and show error
    switchingProject.value = false
    const msgKey = err?.msgKey
    if (msgKey === 'NotADirectory') {
      toast.show(t('appHeader.projectPathNotFound'), { icon: '⚠️', type: 'error', duration: 3000 })
      // Remove stale project from recent list
      fetch('/api/recent-projects', {
        method: 'DELETE',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: newProjectPath })
      }).catch(() => {})
    } else {
      toast.show(t('appHeader.switchProjectFailed', { error: err.message }), { icon: '⚠️', type: 'error', duration: 3000 })
    }
    return
  }

  // ── Phase 3: Reset module-level singletons ──
  // (store state already reset by setProject, but identity/agents need explicit reset)
  resetIdentity()
  resetAgents()
  resetChatSessionState()
  clearPlanState()
  resetTaskTabState()
  fileNav.closeOverlay()

  // ── Phase 4: Change key → Vue destroys old component tree & builds new one ──
  projectKey.value = newProjectPath

  // ── Phase 5: Reload all project-scoped data (mirrors onMounted after auth) ──
  try { await store.loadProject() } catch (_) {
    toast.show(t('toast.projectLoadFailed'), { icon: '⚠️', type: 'error', duration: 0, onClick: () => location.reload() })
  }
  await sessionIdentity.initSessionFromAPI()
  loadSessionsOnce()
  try { await store.loadFiles('') } catch (_) {}
  store.loadGitBranch().catch(() => {})
  loadTasks()
  loadConfig()
  loadSSHInfo().catch(() => {})
  loadTerminalStatus().catch(() => {})
  if (isAppMode.value) syncToNative().catch(() => {})

  // ── Phase 6: Handle cross-project pending navigation ──
  if (pendingSessionId) {
    const checkReady = () => {
      if (sessionIdentity.currentSessionId.value) {
        switchTab('chat')
        sessionIdentity.switchSession(pendingSessionId)
      } else {
        setTimeout(checkReady, 100)
      }
    }
    checkReady()
  }

  // ── Phase 8: Fade in ──
  switchingProject.value = false
}

const activeTab = ref('chat')
const tabDirection = ref('right') // 'left' 或 'right'，用于控制界面切换动画方向
const dockRef = ref<HTMLElement | null>(null)

// ── Liquid Tab Bar ──
const mainTabs = ref([
  { id: 'chat', index: 0 },
  { id: 'browse', index: 1 },
  { id: 'history', index: 2 },
  { id: 'slot4', index: 3 },
])

// LiquidTabBar 切换事件处理（第4格跟随 dockSlot4Tab 动态变化）
const tabMapping = computed(() => ['chat', 'browse', 'history', dockSlot4Tab.value, 'more'])

// 导航栏标签列表（传给 LiquidTabBar 组件，5个 tab）
const tabList = computed(() => [
  { name: t('nav.chat'), icon: IconMessageCircle },
  { name: t('nav.fileManager'), icon: IconFolderOpen },
  { name: t('nav.history'), icon: GitBranch },
  { name: dockSlot4Title.value, icon: dockSlot4Icon.value },  // 动态第4格：跟随"更多"菜单选择
  { name: t('nav.more'), icon: MoreHorizontal },
])

// 当前激活标签的索引（双向绑定）
const activeIndex = computed({
  get: () => tabMapping.value.indexOf(activeTab.value),
  set: (index) => {
    const tabId = tabMapping.value[index]
    if (tabId && tabId !== 'more') switchTab(tabId)
  }
})

// LiquidTabBar 内部 ref（用于溢出菜单定位）
let taskBtnRef = null

// ChatPanelContent ref（用于 header ACP 按钮调用 openAcpDrawer）
const chatPanelRef = ref(null)

// ACP 恢复按钮显示条件
const showAcpResume = computed(() => sessionIdentity.currentAgentId.value && agentCanResume(sessionIdentity.currentAgentId.value))

// 任务运行状态（用于雷达动画）
const isAgentRunning = computed(() => store.state.chatRunning || store.state.taskRunning)

// 未读通知（用于更多按钮红点）- 包含溢出菜单子项的徽标状态
const hasUnreadNotification = computed(() => {
  return store.state.taskUnreadCount > 0 ||
    store.state.chatUnreadCount > 0 ||
    store.state.portForwardActiveCount > 0 ||
    store.state.terminalSessionCount > 0
})

// Tab 切换处理
function handleTabChange(index) {
  const tabId = tabMapping.value[index]
  if (tabId === 'more') {
    toggleOverflowMenu()
    return
  }
  if (tabId) switchTab(tabId)
}

function switchTab(tab) {
  if (activeTab.value === tab) return

  // 计算切换方向
  const currentIndex = tabMapping.value.indexOf(activeTab.value)
  const targetIndex = tabMapping.value.indexOf(tab)
  if (currentIndex !== -1 && targetIndex !== -1) {
    tabDirection.value = targetIndex > currentIndex ? 'left' : 'right'
  }

  activeTab.value = tab
  if (tab === 'browse') {
    store.loadFiles(store.state.currentDir)
  }
  if (tab === 'chat') {
    // Recalculate instead of blindly clearing — if the user switches to chat
    // but hasn't opened the unread session, the indicator should keep flashing.
    // loadSessionsOnce checks unreadCount per session (excluding current), so
    // it only clears when all sessions are actually read.
    loadSessionsOnce()
  }
  if (tab === 'tasks') {
    // Only stop dock button flash — don't clear per-task unread badges.
    // Per-task badges are cleared when the user enters that task's execution history.
    store.state.taskUnreadCount = 0
  }
  // Close overflow menu when switching to a main tab
  if (!overflowTabs.value.includes(tab)) {
    overflowMenuOpen.value = false
  }
}

// 手势事件直接委托给 composable 的 handlePointerDown/Move/Up

// 初始化液态导航栏容器引用
onMounted(() => {
  if (dockRef.value) {
    liquidSetContainerRef(dockRef.value)
  }
})

const detailsOpen = ref(false)
const tocOpen = ref(false)
const searchOpen = ref(false)
const fileHistoryOpen = ref(false)

function openFileHistory() {
  fileHistoryOpen.value = true
}

const markdownViewMode = ref('rendered')

const toast = useToast()
provide('toast', toast)

const sessionIdentity = useSessionIdentity()

const showHidden = ref(false)
const { localConfig, setLocalConfig: setSetting, loadConfig, getServerValueWithDefault } = useSettingsConfig()
// Initialize from settings config (which handles legacy key migration)
showHidden.value = !!localConfig.showHidden
const sortField = ref(localConfig.sortField || null)
const sortDir = ref(localConfig.sortDir || 'asc')

useFileWatch({
  fileManagerOpen: computed(() => activeTab.value === 'browse'),
  currentDir: computed(() => store.state.currentDir),
  currentFile: computed(() => store.state.currentFile),
})

const fileNav = useFileNavStack()

const { isAppMode } = useAppMode()
const { syncToNative, sshInfo, loadSSHInfo } = usePortForward()
const { terminalRuntimeEnabled, loadTerminalStatus } = useTerminalStatus()
const isSSHDisabled = computed(() => sshInfo.value?.enabled === false)
// Use runtime status (actual server state) not config value — mirrors SSH pattern.
// Config may say enabled=true before restart; the runtime API returns false until
// the terminal manager actually exists.  `null` means "not yet loaded" → treat as
// disabled to avoid a flash of the terminal button on first mount.
const isTerminalDisabled = computed(() => terminalRuntimeEnabled.value !== true)
watch(isSSHDisabled, (disabled) => {
  if (disabled && activeTab.value === 'proxy') {
    switchTab('chat')
  }
})
watch(isTerminalDisabled, (disabled) => {
  if (disabled && activeTab.value === 'terminal') {
    switchTab('chat')
  }
})
const { navigateToTaskSettings, navigateToTaskHistory, openExecDetail, loadTasks } = useTaskTab()
registerSwitchTab(switchTab)

// Wire up WS global events
const { onEvent, init: initGlobalEvents, destroy: destroyGlobalEvents } = useGlobalEvents()
const removeTaskHandler = onEvent((event, data) => {
    if (event === 'task_update') {
        onTaskEvent(data)
    }
})

const handleForeground = () => {
    // Only refresh after initialization is complete — during cold start
    // the onMounted handler loads fresh data; refreshing here with stale
    // state (e.g. old currentDir from WebView cache) would show wrong dir.
    if (!isAuthenticated.value) return
    // Full state pull — refresh everything that may have changed while backgrounded
    loadSessionsOnce()
    store.loadFiles(store.state.currentDir)
    store.loadGitBranch()
    loadTasks()
    loadTerminalStatus()
    if (store.state.currentFile?.path) {
        refreshCurrentFile()
    }
}

// Edge swipe back gesture detection (right-edge-left-swipe → go back)
useEdgeSwipeBack()

// 文件覆盖层的返回手势：文件栈优先
useFeatureBackHandler(
  'file-overlay',
  () => activeTab.value === 'browse' && fileNav.overlayOpen.value,
  () => {
    if (fileNav.canGoBack.value) {
      const prevPath = fileNav.goBack()
      if (prevPath) store.selectFile(prevPath)
    } else {
      fileNav.closeOverlay()
      tocOpen.value = false
      detailsOpen.value = false
      searchOpen.value = false
      fileHistoryOpen.value = false
    }
  },
)

// Android hardware back button / predictive back gesture → delegate to JS
window.addEventListener('clawbench-back-press', () => {
    // If any feature can handle back, do it and prevent the default Android behavior
    const handled = handleBackNavigation()
    // Set flag for the Android native code to check
    window.__clawbenchBackHandled = !!handled
})
window.addEventListener('clawbench-foreground', handleForeground)
const terminalRequestedCwd = ref(null)

// Hide AppHeader when terminal tab is active (always); remove padding-top too
// so terminal fills the full screen. Dock is hidden only when keyboard is open.
const terminalActive = computed(() => activeTab.value === 'terminal')
const { keyboardHeight: terminalKeyboardHeight } = useTerminalKeyboard()
const terminalKeyboardActive = computed(() => terminalActive.value && terminalKeyboardHeight.value > 0)

// Chat keyboard — on iOS WKWebView there's no adjustResize, so we detect
// keyboard via visualViewport and compensate in the web layer.
const { chatKeyboardHeight } = useChatKeyboard()
const chatKeyboardActive = computed(() => activeTab.value === 'chat' && chatKeyboardHeight.value > 0)

// Unified: any soft keyboard is open (terminal or chat)
const anyKeyboardActive = computed(() => terminalKeyboardActive.value || chatKeyboardActive.value)

const quoteQuestion = useQuoteQuestion()
const sessionDrawerRef = ref(null)

// Register SessionDrawer ref so identity.openAgentSelector() works
watch(sessionDrawerRef, (ref) => {
  if (ref) registerSessionDrawerRef(ref)
}, { immediate: true })

// Register identity actions (switchSession, createSession, etc.)
// These will be overwritten by ChatPanelContent when it mounts, but
// openAgentSelector is NOT registered here — it's handled via
// registerSessionDrawerRef above, which is independent.
function handleQuoteOpenSessions() {
  sessionIdentity.openSessionTab()
}

function handleSessionSelect(sessionId, backend) {
  sessionIdentity.switchSession(sessionId)
  sessionIdentity.sessionDrawerOpen.value = false
}

async function handleSessionCreate(agentId) {
  await sessionIdentity.createSession(agentId)
  // If drawer is still open, add the new session to the local list
  if (sessionDrawerRef.value && sessionIdentity.sessionDrawerOpen.value) {
    const id = sessionIdentity.currentSessionId.value
    if (id) {
      sessionDrawerRef.value.addSessionLocally({
        id,
        title: sessionIdentity.currentSessionTitle.value || '',
        backend: sessionIdentity.currentBackend.value || '',
        agentId: sessionIdentity.currentAgentId.value || '',
        model: sessionIdentity.currentModelName.value || '',
        updatedAt: new Date().toISOString(),
        unreadCount: 0,
      })
    }
  }
  sessionIdentity.sessionDrawerOpen.value = false
}

function handleSessionDelete(sessionId, backend) {
  sessionIdentity.deleteSession(sessionId, backend)
}

async function handleLoginSuccess() {
    // Load project BEFORE setting isAuthenticated so the backend sets the
    // clawbench_project cookie first. Without this, ChatPanelContent mounts
    // and calls loadHistory() which fails with NoProjectSelected (no cookie).
    try { await store.loadProject() } catch (_) { /* loadProject has its own error handling */ }
    // Check if setup wizard is needed (no agents + embedded Pi binary)
    try {
      const resp = await fetch('/api/setup/status')
      if (resp.ok) {
        const data = await resp.json()
        if (data.needs_setup) {
          isAuthenticated.value = true
          needsSetup.value = true
          return
        }
      }
    } catch { /* proceed to normal app if check fails */ }
    isAuthenticated.value = true
    initMermaid()
    await store.loadFiles('')
}

async function handleSetupComplete() {
    // Reset cached agents so fresh data is loaded
    resetAgents()

    // Register event listeners that were skipped during wizard (onMounted skipped them)
    window.addEventListener('open-file-manager', handleOpenFileManager)
    window.addEventListener('open-file-overlay', handleOpenFileOverlay)
    window.addEventListener('navigate-to-commit', handleNavigateToCommit)
    window.addEventListener('quote-sent', playQuoteEmitAnimation)
    window.addEventListener('scroll-to-line', (e) => { scrollToLine(e.detail.line) })
    document.addEventListener('click', handleOverflowOutsideClick)
    window.addEventListener('clawbench-theme-change', (e) => {
        const resolved = e.detail
        theme.value = resolved
        initMermaid()
        reRenderMermaid()
    })
    window.addEventListener('clawbench-showhidden-change', (e) => {
        showHidden.value = e.detail
    })
    window.addEventListener('clawbench-sort-change', (e) => {
        if (e.detail.field !== undefined) sortField.value = e.detail.field
        if (e.detail.dir !== undefined) sortDir.value = e.detail.dir
    })
    loadTasks()
    loadConfig()

    // Load project first so backend sets clawbench_project cookie
    try { await store.loadProject() } catch (_) { /* best effort */ }

    // Load agents and session data BEFORE switching to main UI
    // to prevent error flashes (e.g., "no agent configured")
    try {
        await sessionIdentity.initSessionFromAPI()
    } catch { /* best effort */ }
    loadSessionsOnce().catch(() => {})
    store.loadGitBranch().catch(() => {})

    // Now switch to main UI — agents and session are loaded
    needsSetup.value = false
    initMermaid()

    // Continue with remaining init that was deferred
    if (isAppMode.value) syncToNative().catch(() => {})
    loadSSHInfo().catch(() => {})
    loadTerminalStatus().catch(() => {})
    try { await store.loadFiles('') } catch (_) {}
}

const projectDialogOpen = ref(false)

function handleOpenProjectDialog() {
    projectDialogOpen.value = true
}

const theme = ref(localConfig.theme === 'auto'
    ? (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')
    : (localConfig.theme || (window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light')))

const dirEntries = computed(() => store.state.dirEntries)
const currentDir = computed(() => store.state.currentDir)
const currentFile = computed(() => store.state.currentFile)
const currentFileIsMarkdown = computed(() => {
    const f = currentFile.value
    if (!f) return false
    const ft = getFileType(f.name)
    return ft?.isMarkdown || ft?.isHtml || false
})
const projectRoot = computed(() => store.state.projectRoot)
const homeDir = computed(() => store.state.homeDir)

const tocFile = computed(() => {
    const f = currentFile.value
    if (!f || f.isImage || f.isAudio) return null
    // PDF: pass file even without content (outline comes from pdfOutline prop)
    if (f.isPdf) return f
    if (!f.content) return null
    const ft = getFileType(f.name)
    if (ft.isImage || ft.isAudio) return null
    return f
})

// PDF TOC integration
const fileOverlayRef = ref(null)
const fileManagerRef = ref(null)
const pdfOutline = computed(() => fileOverlayRef.value?.pdfOutline || [])
function handleJumpPdfPage(pageNum) {
    fileOverlayRef.value?.pdfScrollToPage(pageNum)
}

watch(() => currentFile.value, (f) => {
    tocOpen.value = false
    detailsOpen.value = false
    markdownViewMode.value = 'rendered'
})

function toggleHidden() {
    showHidden.value = !showHidden.value
    setSetting('showHidden', showHidden.value)
    store.loadFiles(store.state.currentDir)
}

function handleToggleSort(field) {
    if (sortField.value === field) {
        if (sortDir.value === 'asc') {
            sortDir.value = 'desc'
        } else {
            sortField.value = null
            sortDir.value = 'asc'
        }
    } else {
        sortField.value = field
        sortDir.value = 'asc'
    }
    setSetting('sortField', sortField.value)
    setSetting('sortDir', sortDir.value)
}

async function handleNavigateDir(path) {
    if (store.state.dirLoading) return
    await store.navigateToDir(path)
}

async function handleSelectFile(path) {
    const ok = await store.selectFile(path)
    if (ok) {
        activeTab.value = 'browse'
        fileNav.openFile(path)
    }
}

async function handleBrowseSelectFile(path) {
    if (fileManagerRef.value?.multiSelectState?.active) return
    const ok = await store.selectFile(path)
    if (ok) {
        fileNav.openFile(path)
    }
}

async function handleTaskOpenFile(filePath, lineStart) {
    const ok = await store.selectFile(filePath)
    if (ok) {
        activeTab.value = 'browse'
        fileNav.openFile(filePath)
        if (lineStart) scrollToLine(lineStart)
    }
}

function handleOverlayClose() {
    fileNav.closeOverlay()
    tocOpen.value = false
    detailsOpen.value = false
    searchOpen.value = false
    fileHistoryOpen.value = false
}

async function handleOverlayGoBack() {
    const prevPath = fileNav.goBack()
    if (prevPath) {
        await store.selectFile(prevPath)
    }
}

async function handleOverlayOpenFile(path) {
    const ok = await store.selectFile(path)
    if (ok) {
        fileNav.openFile(path)
    }
}

function handleOpenFileOverlay(e) {
    const { path, lineStart } = e.detail || {}
    if (!path) return
    activeTab.value = 'browse'
    fileNav.openFile(path)
    if (lineStart) scrollToLine(lineStart)
}

function onTaskCardClick(taskId) {
    navigateToTaskSettings(taskId)
    switchTab('tasks')
}

async function handleRename({ path, name }) {
    await store.renameFile(path, name)
}

async function handleDelete(path) {
    await store.deleteFile(path)
}

async function handleBatchDelete(paths) {
    await store.deleteFiles(paths)
}

async function handleRefresh() {
    await refreshCurrentFile({ loadDir: true, clearOnError: true })
}

function handleDockTerminal() {
    terminalRequestedCwd.value = null
    switchTab('terminal')
}

// Overflow menu state
const overflowMenuOpen = ref(false)
let overflowBtnRef = null // 模板中通过 :ref 函数直接赋值
const overflowTabs = computed(() => {
  const tabs = ['tasks']
  if (!isSSHDisabled.value) tabs.push('proxy')
  if (!isTerminalDisabled.value) tabs.push('terminal')
  tabs.push('settings')
  return tabs
})
const overflowTabMeta = {
  tasks:   { icon: CalendarClock, titleKey: 'nav.tasks' },
  proxy:   { icon: EthernetPort, titleKey: 'nav.portForward' },
  terminal:{ icon: TerminalIcon, titleKey: 'terminal.title' },
  settings:{ icon: Settings, titleKey: 'nav.settings' },
}

// Dock slot 4: dynamic slot showing user's selected overflow item
const STORAGE_KEY_DOCK_SLOT4 = 'clawbench_dock_slot4'
const dockSlot4Tab = ref(localStorage.getItem(STORAGE_KEY_DOCK_SLOT4) || 'tasks')
const dockSlot4Icon = computed(() => overflowTabMeta[dockSlot4Tab.value]?.icon ?? CalendarClock)
const dockSlot4Title = computed(() => overflowTabMeta[dockSlot4Tab.value] ? t(overflowTabMeta[dockSlot4Tab.value].titleKey) : t('nav.tasks'))

function setDockSlot4(tab) {
  dockSlot4Tab.value = tab
  localStorage.setItem(STORAGE_KEY_DOCK_SLOT4, tab)
}

function handleDockSlot4Click() {
  const tab = dockSlot4Tab.value
  if (tab === 'terminal') {
    handleDockTerminal()
  } else {
    switchTab(tab)
  }
}

const isOverflowTabActive = computed(() => overflowTabs.value.includes(activeTab.value) && activeTab.value !== dockSlot4Tab.value)

// If the saved dock-slot4 tab becomes unavailable (e.g. terminal disabled), fall back to tasks
watch(overflowTabs, (tabs) => {
  if (!tabs.includes(dockSlot4Tab.value)) {
    setDockSlot4('tasks')
  }
})

const overflowPopupStyle = computed(() => {
  const btn = overflowBtnRef
  if (!btn) return {}
  const rect = btn.getBoundingClientRect()
  return {
    position: 'fixed',
    bottom: `${window.innerHeight - rect.top + 8}px`,
    right: `${window.innerWidth - rect.right}px`,
  }
})

// Dock badge change animations
const historyBadgeAnim = ref(false)
const taskBadgeAnim = ref(false)
const terminalBadgeAnim = ref(false)
const proxyBadgeAnim = ref(false)
const overflowBadgeAnim = ref(false)

function triggerBadgeAnim(animRef) {
  animRef.value = false
  nextTick(() => { animRef.value = true })
}

watch(() => store.state.gitWorkingTreeChangeCount, (n, o) => { if (o !== undefined && n !== o) triggerBadgeAnim(historyBadgeAnim) })
watch(() => store.state.taskUnreadCount, (n, o) => {
  if (o !== undefined && n !== o) {
    triggerBadgeAnim(taskBadgeAnim)
    triggerBadgeAnim(overflowBadgeAnim)
  }
})
watch(() => store.state.terminalSessionCount, (n, o) => {
  if (o !== undefined && n !== o) {
    triggerBadgeAnim(terminalBadgeAnim)
    triggerBadgeAnim(overflowBadgeAnim)
  }
})
watch(() => store.state.portForwardActiveCount, (n, o) => {
  if (o !== undefined && n !== o) {
    triggerBadgeAnim(proxyBadgeAnim)
    triggerBadgeAnim(overflowBadgeAnim)
  }
})

const overflowButtonIcon = computed(() => {
  // Show the active overflow tab's icon, unless it's the dock-slot4 tab (which has its own button)
  if (activeTab.value === dockSlot4Tab.value) return MoreHorizontal
  return overflowTabMeta[activeTab.value]?.icon ?? MoreHorizontal
})

const overflowButtonTitle = computed(() => {
  if (activeTab.value === dockSlot4Tab.value) return t('nav.more')
  return overflowTabMeta[activeTab.value] ? t(overflowTabMeta[activeTab.value].titleKey) : t('nav.more')
})

function toggleOverflowMenu() {
  if (isOverflowTabActive.value && !overflowMenuOpen.value) {
    // If already on an overflow tab, first click opens menu to allow switching
    overflowMenuOpen.value = true
  } else if (overflowMenuOpen.value) {
    overflowMenuOpen.value = false
  } else {
    overflowMenuOpen.value = true
  }
}

function handleOverflowSelect(tab) {
  if (activeTab.value === tab) {
    // Already on this tab, just close the menu
    overflowMenuOpen.value = false
    return
  }
  overflowMenuOpen.value = false

  // 计算切换方向（从"更多"菜单切换时，默认向左滑入）
  tabDirection.value = 'left'

  // 同步批量更新 dockSlot4Tab + activeTab，避免 Vue 分两次 DOM 更新导致指示器错位
  if (tab !== 'settings') {
    dockSlot4Tab.value = tab
    localStorage.setItem(STORAGE_KEY_DOCK_SLOT4, tab)
  }
  activeTab.value = tab
  // terminal 需要额外处理（清空 cwd）
  if (tab === 'terminal') {
    terminalRequestedCwd.value = null
  }
}

function handleOverflowSettings() {
  overflowMenuOpen.value = false
  switchTab('settings')
}

// Close overflow menu on outside click
function handleOverflowOutsideClick(e) {
  if (overflowMenuOpen.value && !e.target.closest('.dock-overflow-popup') && !e.target.closest('.dock-overflow-btn')) {
    overflowMenuOpen.value = false
  }
}

function handleOpenTerminal(cwd) {
    terminalRequestedCwd.value = cwd || null
    switchTab('terminal')
}

function scrollToLine(line) {
    nextTick(() => {
        const el = document.querySelector(`.code-line[data-line="${line}"]`)
        if (!el) return
        el.scrollIntoView({ behavior: 'smooth', block: 'center' })
        el.classList.add('line-flash')
        el.addEventListener('animationend', () => el.classList.remove('line-flash'), { once: true })
    })
}

function toggleTheme() {
    theme.value = theme.value === 'dark' ? 'light' : 'dark'
    applyTheme(theme.value)
}

function applyTheme(t) {
    document.documentElement.setAttribute('data-theme', t)
    setSetting('theme', t)
    document.documentElement.setAttribute('data-hljs-theme', t)
    initMermaid()
    reRenderMermaid()
    // iOS: 同步更新状态栏样式
    updateStatusBarStyle(t)
}

// iOS: 根据主题更新状态栏样式
async function updateStatusBarStyle(t) {
    try {
        const { StatusBar } = await import('@capacitor/status-bar')
        if (t === 'dark') {
            await StatusBar.setStyle({ style: 'DARK' })
        } else {
            await StatusBar.setStyle({ style: 'LIGHT' })
        }
    } catch {
        // 非 iOS 环境或插件不可用时忽略
    }
}

provide('theme', theme)
provide('applyTheme', applyTheme)
provide('activeTab', activeTab)
provide('switchTab', switchTab)
provide('hotSwitchProject', hotSwitchProject)

function handleOpenFileManager() {
    activeTab.value = 'browse'
}

function handleNavigateToCommit(e) {
    const sha = e?.detail?.sha
    if (sha) {
        setPendingCommitNavigation(sha)
    }
    activeTab.value = 'history'
}

function playQuoteEmitAnimation(e) {
  const { from, to } = e?.detail ?? {}
  if (!from || !to) return
  const x0 = from.x, y0 = from.y, x1 = to.x, y1 = to.y
  const mx = (x0 + x1) / 2
  const my = Math.min(y0, y1) - 30
  const dot = document.createElement('div')
  dot.className = 'quote-emit-dot'
  dot.style.cssText = `
    position: fixed; width: 8px; height: 8px; border-radius: 50%;
    background: var(--accent-color, #0066cc);
    box-shadow: 0 0 10px 3px color-mix(in srgb, var(--accent-color, #0066cc) 50%, transparent);
    z-index: 9999; pointer-events: none; left: 0; top: 0; will-change: transform, opacity;
  `
  document.body.appendChild(dot)
  const duration = 420, start = performance.now()
  function animate(now) {
    const t = Math.min((now - start) / duration, 1)
    const ease = 1 - Math.pow(1 - t, 3)
    const x = (1 - ease) ** 2 * x0 + 2 * (1 - ease) * ease * mx + ease ** 2 * x1
    const y = (1 - ease) ** 2 * y0 + 2 * (1 - ease) * ease * my + ease ** 2 * y1
    const scale = t < 0.1 ? t / 0.1 : t > 0.85 ? 1 - (t - 0.85) / 0.15 : 1
    const opacity = t < 0.08 ? t / 0.08 : t > 0.7 ? 1 - (t - 0.7) / 0.3 : 1
    dot.style.transform = `translate(${x - 4}px, ${y - 4}px) scale(${scale})`
    dot.style.opacity = opacity
    if (t < 1) requestAnimationFrame(animate)
    else {
      dot.remove()
      const chatDockBtn = document.querySelector('.dock-center')?.querySelector('.dock-btn')
      if (chatDockBtn) {
        chatDockBtn.classList.add('quote-emit-receive')
        chatDockBtn.addEventListener('animationend', () => chatDockBtn.classList.remove('quote-emit-receive'), { once: true })
      }
    }
  }
  requestAnimationFrame(animate)
}

onMounted(async () => {
    // === 平台标识挂载 ===
    var _ios = isIOSApp()
    if (_ios) {
        document.body.classList.add('ios-app')
    }

    applyTheme(theme.value)
    let resp
    try {
        resp = await fetch('/api/me')
    } catch (_) {
        isAuthenticated.value = false
        if (isAppMode.value) {
            toast.show(t('toast.serverUnreachableApp'), { icon: '⚠️', type: 'error', duration: 5000 })
        } else {
            toast.show(t('toast.serverUnreachableWeb'), { icon: '⚠️', type: 'error', duration: 0, onClick: () => location.reload() })
        }
        return
    }
    if (resp.ok) {
        isAuthenticated.value = true
    } else if (resp.status === 401 || resp.status === 403) {
        isAuthenticated.value = false
        return
    } else {
        isAuthenticated.value = false
        if (isAppMode.value) {
            toast.show(t('toast.serverError'), { icon: '⚠️', type: 'error', duration: 5000 })
        } else {
            toast.show(t('toast.serverError'), { icon: '⚠️', type: 'error', duration: 0, onClick: () => location.reload() })
        }
        return
    }
    // Check if setup wizard is needed BEFORE any main app initialization.
    // If needs_setup, show wizard and skip all main UI loading to prevent
    // error flashes (e.g., "no agent configured", "failed to load chat history").
    try {
      const setupResp = await fetch('/api/setup/status')
      if (setupResp.ok) {
        const setupData = await setupResp.json()
        if (setupData.needs_setup) {
          needsSetup.value = true
          initGlobalEvents() // Needed for WS connection (setup wizard uses API)
          return  // Skip ALL main app initialization — wizard will handle it
        }
      }
    } catch { /* proceed to normal app if check fails */ }

    // ── Main app initialization (only when setup is NOT needed) ──
    initGlobalEvents()
    initMermaid()
    loadTasks()
    loadConfig()
    window.addEventListener('open-file-manager', handleOpenFileManager)
    window.addEventListener('open-file-overlay', handleOpenFileOverlay)
    window.addEventListener('navigate-to-commit', handleNavigateToCommit)
    window.addEventListener('quote-sent', playQuoteEmitAnimation)
    window.addEventListener('scroll-to-line', (e) => { scrollToLine(e.detail.line) })
    document.addEventListener('click', handleOverflowOutsideClick)
    window.addEventListener('clawbench-theme-change', (e) => {
        const resolved = e.detail
        theme.value = resolved
        initMermaid()
        reRenderMermaid()
    })
    window.addEventListener('clawbench-showhidden-change', (e) => {
        showHidden.value = e.detail
    })
    window.addEventListener('clawbench-sort-change', (e) => {
        if (e.detail.field !== undefined) sortField.value = e.detail.field
        if (e.detail.dir !== undefined) sortDir.value = e.detail.dir
    })
    // Load project first so the backend sets the clawbench_project cookie.
    // Without this, subsequent chat/session API calls fail with NoProjectSelected
    // on first login (no cookie yet) and show "加载聊天记录失败".
    try { await store.loadProject() } catch (_) {
        toast.show(t('toast.projectLoadFailed'), { icon: '⚠️', type: 'error', duration: 0, onClick: () => location.reload() }); return
    }
    await sessionIdentity.initSessionFromAPI()
    // Use loadSessionsOnce() which correctly sets chatUnread to true OR false.
    // The old code only set chatUnread=true and never corrected a stale true.
    loadSessionsOnce()
    if (isAppMode.value) syncToNative().catch(() => {})
    // Resume Android log capture if previously enabled
    loadSSHInfo().catch(() => {})
    loadTerminalStatus().catch(() => {})
    try { await store.loadFiles('') } catch (_) {
        toast.show(t('toast.fileListLoadFailed'), { icon: '⚠️', type: 'error', duration: 6000 })
    }
    // Handle pending navigation from push notification deep link
    // (cross-project reload or cold start via AndroidNative bridge)
    const processPendingSessionNav = (navSessionId) => {
      // Wait for sessions to load before switching (max 3 seconds)
      let attempts = 0
      const checkReady = () => {
        if (sessionIdentity.currentSessionId.value) {
          switchTab('chat')
          sessionIdentity.switchSession(navSessionId)
        } else if (attempts < 30) {
          attempts++
          setTimeout(checkReady, 100)
        }
      }
      checkReady()
    }

    const processPendingTaskNav = async (navTaskId, navExecutionId) => {
      // Ensure tasks are loaded before navigating
      try {
        await loadTasks()
      } catch (_) {
        // Proceed anyway — the task list may already be populated
      }
      switchTab('tasks')
      navigateToTaskHistory(Number(navTaskId))
      if (navExecutionId) {
        // openExecDetail without execData will auto-fetch from API via refreshExecDetail
        openExecDetail(navExecutionId)
      }
    }

    // Check localStorage for pending navigation (cross-project reload)
    const pendingNav = localStorage.getItem('clawbenchPendingNav')
    if (pendingNav) {
      localStorage.removeItem('clawbenchPendingNav')
      try {
        const nav = JSON.parse(pendingNav)
        if (nav.taskId) {
          processPendingTaskNav(nav.taskId, nav.executionId)
        } else if (nav.sessionId) {
          processPendingSessionNav(nav.sessionId)
        }
      } catch (_) {}
    }

})

// 悬浮输入框发送消息处理
function handleFloatingSend(text) {
  // 这里需要调用 ChatPanelContent 的发送方法
  // 暂时使用 console.log 调试
  console.log('[FloatingInput] send:', text)
  // TODO: 对接实际发送逻辑
}

// 键盘监听已移至 FloatingInputBar 组件内部处理

onUnmounted(() => {
    removeTaskHandler()
    window.removeEventListener('clawbench-foreground', handleForeground)
    destroyGlobalEvents()
    window.removeEventListener('open-file-manager', handleOpenFileManager)
    window.removeEventListener('open-file-overlay', handleOpenFileOverlay)
    window.removeEventListener('navigate-to-commit', handleNavigateToCommit)
    window.removeEventListener('quote-sent', playQuoteEmitAnimation)
    document.removeEventListener('click', handleOverflowOutsideClick)
    document.removeEventListener('focusin', handleFocus)
    document.removeEventListener('focusout', handleBlur)
    Keyboard.removeAllListeners()
    window.removeEventListener('nativeTabSwitch', listenNativeTabEvent)
})
</script>

<style scoped>
/* SPA hot project switch: fade transition to mask intermediate state */
.app-container {
    transition: opacity 0.15s ease;
}
.app-container.project-switching {
    opacity: 0;
}

.browse-panel {
  position: relative;
  width: 100%;
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

/* When terminal tab is active, remove header padding so content expands to top */
.chrome-hidden {
    padding-top: 0 !important;
}

/* When chat keyboard is open on iOS (no adjustResize), shrink the app container
   from the bottom so content stays above the keyboard. */
.chat-keyboard-open {
    bottom: v-bind(chatKeyboardHeight + 'px') !important;
}

/* ═══════════════════════════════════════════════════════════════════════
 * Apple iOS Liquid Glass Tab Bar
 * Inspired by iOS 26 Liquid Glass + Apple Music + visionOS Material
 * Floating Capsule | Liquid Glass | Refraction | Optical Depth
 * Dynamic theme adaptation for light/dark modes
 * ═══════════════════════════════════════════════════════════════════════ */

/* Header 行 ACP 恢复按钮（与会话名称同一行靠右） */
.header-acp-btn {
    flex-shrink: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 28px;
    height: 28px;
    margin-left: auto;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--accent-color, #0066cc);
    cursor: pointer;
    transition: background 0.15s;
}

.header-acp-btn:hover {
    background: color-mix(in srgb, var(--accent-color, #0066cc) 10%, transparent);
}

.header-acp-btn:active {
    background: color-mix(in srgb, var(--accent-color, #0066cc) 25%, transparent);
    transform: scale(0.92);
}

.bottom-dock-wrapper {
    position: fixed;
    bottom: 0;
    left: 0;
    right: 0;
    z-index: 100;
    padding: 0 16px 27px;
    display: flex;
    justify-content: center;
    width: 100%;
}

/* 原厂导航项的事件挂载与排版壳子 */
.original-tab-proxy-node {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    width: 100%;
    height: 100%;
}

/* 原厂消息徽章小红点定位 */
.dock-badge-dot {
    position: absolute;
    top: -2px;
    right: -2px;
    width: 8px;
    height: 8px;
    background-color: #ff3b30;
    border-radius: 50%;
    box-shadow: 0 0 0 2px rgba(22, 24, 30, 0.8);
}

/* 浅色模式：徽章红点描边改为白色 */
[data-theme="light"] .dock-badge-dot {
    box-shadow: 0 0 0 2px rgba(255, 255, 255, 0.9);
}

/* 会话图标运行状态容器 */
.radar-icon-wrapper {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
}

/* 呼吸脉冲：图标明暗交替 + 微缩放，像心跳 */
.radar-icon-wrapper.has-running {
    animation: dock-breathe 1.5s ease-in-out infinite;
    color: var(--accent-color, #0066cc) !important;
}

@keyframes dock-breathe {
    0%, 100% { opacity: 0.4; transform: scale(0.95); }
    50%      { opacity: 1;   transform: scale(1.05); }
}

.bottom-dock {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 8px 16px;
    border-radius: 28px; /* Floating capsule shape */
    position: relative;
    overflow: hidden;
    touch-action: none; /* 禁止浏览器默认触摸行为 */

    /* ── Liquid Glass Material (Dark Theme - Default) ── */
    background: linear-gradient(
        135deg,
        rgba(255, 255, 255, 0.08) 0%,
        rgba(255, 255, 255, 0.04) 50%,
        rgba(255, 255, 255, 0.06) 100%
    );

    /* Backdrop blur for refraction effect */
    -webkit-backdrop-filter: blur(16px) saturate(180%) brightness(1.05);
    backdrop-filter: blur(16px) saturate(180%) brightness(1.05);

    /* Subtle depth shadow - floating effect */
    box-shadow:
        0 8px 32px rgba(0, 0, 0, 0.12),
        0 2px 8px rgba(0, 0, 0, 0.08),
        inset 0 1px 0 rgba(255, 255, 255, 0.15); /* Top highlight */
}

/* ── Light Theme: Liquid Glass Material ── */
[data-theme="light"] .bottom-dock {
    background: linear-gradient(
        135deg,
        rgba(0, 0, 0, 0.06) 0%,
        rgba(0, 0, 0, 0.03) 50%,
        rgba(0, 0, 0, 0.05) 100%
    );

    box-shadow:
        0 8px 32px rgba(0, 0, 0, 0.08),
        0 2px 8px rgba(0, 0, 0, 0.05),
        inset 0 1px 0 rgba(255, 255, 255, 0.8); /* Top highlight */
}

/* ── Environmental Lighting (Ambient Glow) ── */
.bottom-dock::before {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 28px;
    pointer-events: none;
    z-index: 0;

    /* Left: Cool blue | Right: Warm pink-orange */
    background: linear-gradient(
        90deg,
        rgba(100, 140, 255, 0.03) 0%,    /* Cold blue - 3% */
        transparent 30%,
        transparent 70%,
        rgba(255, 140, 100, 0.03) 100%   /* Warm pink-orange - 3% */
    );
}

/* ── Light Theme: Environmental Lighting ── */
[data-theme="light"] .bottom-dock::before {
    background: linear-gradient(
        90deg,
        rgba(100, 140, 255, 0.02) 0%,    /* Cold blue - 2% */
        transparent 30%,
        transparent 70%,
        rgba(255, 140, 100, 0.02) 100%   /* Warm pink-orange - 2% */
    );
}

/* ── High-Edge Highlight (Glass Reflection) ── */
.bottom-dock::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: 28px;
    pointer-events: none;
    z-index: 1;

    /* Top bright, bottom dark - natural glass reflection */
    background: linear-gradient(
        180deg,
        rgba(255, 255, 255, 0.12) 0%,    /* Top highlight */
        rgba(255, 255, 255, 0.02) 40%,
        rgba(0, 0, 0, 0.02) 60%,
        rgba(0, 0, 0, 0.04) 100%          /* Bottom shadow */
    );

    /* Ultra-thin border effect using box-shadow */
    box-shadow:
        inset 0 0.5px 0 rgba(255, 255, 255, 0.2),   /* Top edge highlight */
        inset 0 -0.5px 0 rgba(0, 0, 0, 0.08);        /* Bottom edge shadow */
}

/* ── Light Theme: High-Edge Highlight ── */
[data-theme="light"] .bottom-dock::after {
    background: linear-gradient(
        180deg,
        rgba(255, 255, 255, 0.9) 0%,     /* Top highlight */
        rgba(255, 255, 255, 0.1) 40%,
        rgba(0, 0, 0, 0.01) 60%,
        rgba(0, 0, 0, 0.03) 100%          /* Bottom shadow */
    );

    box-shadow:
        inset 0 0.5px 0 rgba(255, 255, 255, 0.9),   /* Top edge highlight */
        inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);        /* Bottom edge shadow */
}

.dock-safe-area {
    height: env(safe-area-inset-bottom, 0px);
}

.dock-center {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    position: relative;
    z-index: 2; /* Above glass effects */
}

/* ── Overflow Wrapper & Button ── */
.dock-overflow-wrapper {
    display: flex;
    align-items: center;
    justify-content: center;
}

.dock-overflow-btn {
    width: 34px;
    height: 34px;
}

/* ── Liquid Gooey Wrapper ── */
.liquid-gooey-wrapper {
    position: absolute;
    top: 50%;
    transform: translateY(-50%);
    left: 0;
    right: 0;
    height: 44px;
    z-index: 1;
    pointer-events: none;
    filter: url(#gooey);
}

/* ── Liquid Indicator (唯一滑块) ── */
.liquid-indicator {
    position: absolute;
    top: 0;
    left: 0;
    width: 0;
    opacity: 0;
    height: 44px;
    border-radius: 14px;
    transform-origin: left center;
    will-change: left, width, opacity;
    pointer-events: none;

    /* Liquid Glass Material */
    background: linear-gradient(
        135deg,
        rgba(255, 255, 255, 0.18) 0%,
        rgba(255, 255, 255, 0.12) 50%,
        rgba(255, 255, 255, 0.15) 100%
    );
    -webkit-backdrop-filter: blur(8px) brightness(1.1);
    backdrop-filter: blur(8px) brightness(1.1);

    /* Inner glow */
    box-shadow:
        0 2px 8px rgba(255, 255, 255, 0.08),
        inset 0 1px 0 rgba(255, 255, 255, 0.2),
        inset 0 -1px 0 rgba(0, 0, 0, 0.05);
}

[data-theme="light"] .liquid-indicator {
    background: linear-gradient(
        135deg,
        rgba(0, 0, 0, 0.12) 0%,
        rgba(0, 0, 0, 0.08) 50%,
        rgba(0, 0, 0, 0.1) 100%
    );
    box-shadow:
        0 2px 8px rgba(0, 0, 0, 0.06),
        inset 0 1px 0 rgba(255, 255, 255, 0.5),
        inset 0 -1px 0 rgba(0, 0, 0, 0.03);
}

/* ── Gooey Active State ── */
.bottom-dock.gooey-active .liquid-gooey-wrapper {
    filter: url(#gooey) blur(2px);
}

/* ── Tab Button (SF Symbols Style - Dark Theme) ── */

/* ── Tab Button (SF Symbols Style - Dark Theme) ── */
.dock-btn {
    position: relative;
    width: 44px;
    height: 44px;
    border: none;
    border-radius: 14px;
    background: transparent;
    color: rgba(255, 255, 255, 0.6); /* Unselected: 60% white */
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
    -webkit-tap-highlight-color: transparent;
}

/* ── Light Theme: Tab Button ── */
[data-theme="light"] .dock-btn {
    color: rgba(0, 0, 0, 0.5); /* Unselected: 50% black */
}

.dock-btn:hover {
    color: rgba(255, 255, 255, 0.8);
    background: rgba(255, 255, 255, 0.06);
}

[data-theme="light"] .dock-btn:hover {
    color: rgba(0, 0, 0, 0.7);
    background: rgba(0, 0, 0, 0.05);
}

.dock-btn:active {
    transform: scale(0.92);
}

/* ── Active Tab: 只改文字颜色，不加背景（背景由 .liquid-indicator 唯一滑块负责） ── */
.dock-btn.active {
    color: rgba(255, 255, 255, 0.95); /* Selected: 95% white */
}

[data-theme="light"] .dock-btn.active {
    color: rgba(0, 0, 0, 0.9); /* Selected: 90% black */
}

/* ── Icon Styling (SF Symbols - Dark Theme) ── */
.dock-btn svg {
    width: 20px;
    height: 20px;
    transition: all 0.3s cubic-bezier(0.4, 0, 0.2, 1);
}

/* Selected icon: subtle inner glow */
.dock-btn.active svg {
    filter: drop-shadow(0 0 2px rgba(255, 255, 255, 0.3));
}

/* Light Theme: Selected icon glow */
[data-theme="light"] .dock-btn.active svg {
    filter: drop-shadow(0 0 2px rgba(0, 0, 0, 0.15));
}

.dock-btn.disabled {
    opacity: 0.3;
    cursor: default;
}

/* ── Badge Wrapper (Liquid Animation Support) ── */
.dock-btn-wrap {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    will-change: transform, opacity;
    transform-origin: center center;
    transition: transform 0.2s cubic-bezier(0.25, 1, 0.5, 1), opacity 0.2s;
    /* 磁吸过渡必须轻快敏感 (0.2s) */
    transition: transform 0.2s cubic-bezier(0.25, 1, 0.5, 1), opacity 0.2s;
}

/* ── Unread Badge (Liquid Glass Style) ── */
.dock-badge {
    position: absolute;
    top: 2px;
    right: 2px;
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: linear-gradient(135deg, #ff6b6b, #ff3b3b);
    z-index: 2;
    pointer-events: none;
    box-shadow: 0 2px 4px rgba(255, 59, 59, 0.3);
}

.dock-badge-count {
    width: auto;
    height: auto;
    min-width: 18px;
    padding: 0 5px;
    border-radius: 10px;
    font-size: 10px;
    font-weight: 700;
    line-height: 18px;
    text-align: center;
    color: #fff;
    top: -4px;
    right: -6px;
    background: linear-gradient(135deg, #ff6b6b, #ff3b3b);
    box-shadow: 0 2px 6px rgba(255, 59, 59, 0.4);
}

/* Dock badge pop animation on count change */
.dock-badge-pop {
    animation: badge-pop 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

@keyframes badge-pop {
    0% {
        transform: scale(1);
    }
    40% {
        transform: scale(1.35);
        box-shadow: 0 0 8px 2px color-mix(in srgb, var(--accent-color) 50%, transparent);
    }
    70% {
        transform: scale(0.9);
    }
    100% {
        transform: scale(1);
        box-shadow: 0 0 0 0 transparent;
    }
}

/* ── Running State (Liquid Glass Pulse) ── */
.dock-btn.has-running {
    position: relative;
    isolation: isolate;
    overflow: hidden;
    border-color: transparent;
    box-shadow: 0 0 8px 2px rgba(100, 180, 255, 0.15);
}

.dock-btn.has-running::before {
    content: '';
    position: absolute;
    inset: -2px;
    border-radius: inherit;
    background: conic-gradient(
        from 0deg,
        transparent 0%,
        rgba(100, 180, 255, 0.15) 8%,
        rgba(100, 180, 255, 0.4) 16%,
        rgba(100, 180, 255, 0.6) 22%,
        rgba(100, 180, 255, 0.4) 28%,
        rgba(100, 180, 255, 0.15) 36%,
        transparent 50%
    );
    animation: dock-spin-light 2s linear infinite;
    z-index: -2;
}

.dock-btn.has-running::after {
    content: '';
    position: absolute;
    inset: 1.5px;
    border-radius: inherit;
    background: rgba(255, 255, 255, 0.08);
    z-index: -1;
}

@keyframes dock-spin-light {
    to { transform: rotate(360deg); }
}

/* ── Completed Animation (Liquid Glass) ── */
.dock-btn.just-completed {
    animation: dock-completed-flash 0.3s ease-out;
}

@keyframes dock-completed-flash {
    0% {
        transform: scale(1);
        box-shadow: 0 0 0 0 rgba(100, 255, 150, 0);
    }
    30% {
        transform: scale(1.15);
        box-shadow: 0 0 16px 6px rgba(100, 255, 150, 0.3);
    }
    60% {
        transform: scale(1.08);
        box-shadow: 0 0 8px 3px rgba(100, 255, 150, 0.15);
    }
    100% {
        transform: scale(1);
        box-shadow: 0 0 0 0 rgba(100, 255, 150, 0);
    }
}

/* ── Quote Receive Animation (Liquid Glass) ── */
.dock-btn.quote-emit-receive {
    animation: quote-emit-pulse 0.4s ease-out;
}

@keyframes quote-emit-pulse {
    0% {
        transform: scale(1);
        box-shadow: 0 0 0 0 rgba(100, 180, 255, 0.4);
    }
    40% {
        transform: scale(1.15);
        box-shadow: 0 0 16px 6px rgba(100, 180, 255, 0.3);
    }
    100% {
        transform: scale(1);
        box-shadow: 0 0 0 0 rgba(100, 180, 255, 0);
    }
}

/* Overflow menu */
/* ── Overflow Popup (Liquid Glass Style - Dark Theme) ── */
.dock-overflow-popup {
    background: rgba(30, 30, 30, 0.85);
    -webkit-backdrop-filter: blur(20px) saturate(180%);
    backdrop-filter: blur(20px) saturate(180%);
    border: 0.5px solid rgba(255, 255, 255, 0.12);
    border-radius: 16px;
    padding: 6px;
    box-shadow:
        0 12px 40px rgba(0, 0, 0, 0.3),
        0 4px 12px rgba(0, 0, 0, 0.2),
        inset 0 0.5px 0 rgba(255, 255, 255, 0.15);
    z-index: 9999;
    min-width: 160px;
}

/* ── Light Theme: Overflow Popup ── */
[data-theme="light"] .dock-overflow-popup {
    background: rgba(255, 255, 255, 0.9);
    border: 0.5px solid rgba(0, 0, 0, 0.1);
    box-shadow:
        0 12px 40px rgba(0, 0, 0, 0.15),
        0 4px 12px rgba(0, 0, 0, 0.1),
        inset 0 0.5px 0 rgba(255, 255, 255, 0.8);
}

.dock-overflow-popup::after {
    content: '';
    position: absolute;
    bottom: -6px;
    right: 14px;
    width: 12px;
    height: 12px;
    background: rgba(30, 30, 30, 0.85);
    -webkit-backdrop-filter: blur(20px);
    backdrop-filter: blur(20px);
    border-right: 0.5px solid rgba(255, 255, 255, 0.12);
    border-bottom: 0.5px solid rgba(255, 255, 255, 0.12);
    transform: rotate(45deg);
}

[data-theme="light"] .dock-overflow-popup::after {
    background: rgba(255, 255, 255, 0.9);
    border-right: 0.5px solid rgba(0, 0, 0, 0.1);
    border-bottom: 0.5px solid rgba(0, 0, 0, 0.1);
}

.dock-overflow-item {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    padding: 10px 14px;
    border: none;
    border-radius: 12px;
    background: transparent;
    color: rgba(255, 255, 255, 0.7);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    transition: all 0.2s cubic-bezier(0.4, 0, 0.2, 1);
    white-space: nowrap;
}

[data-theme="light"] .dock-overflow-item {
    color: rgba(0, 0, 0, 0.7);
}

.dock-overflow-item:hover {
    background: rgba(255, 255, 255, 0.08);
    color: rgba(255, 255, 255, 0.95);
}

[data-theme="light"] .dock-overflow-item:hover {
    background: rgba(0, 0, 0, 0.05);
    color: rgba(0, 0, 0, 0.95);
}

@media (hover: none) {
    .dock-overflow-item:hover {
        background: transparent;
        color: rgba(255, 255, 255, 0.7);
    }

    [data-theme="light"] .dock-overflow-item:hover {
        color: rgba(0, 0, 0, 0.7);
    }
}

.dock-overflow-item.active {
    background: rgba(255, 255, 255, 0.12);
    color: rgba(255, 255, 255, 0.95);
}

[data-theme="light"] .dock-overflow-item.active {
    background: rgba(0, 0, 0, 0.08);
    color: rgba(0, 0, 0, 0.95);
}

.dock-overflow-count {
    margin-left: auto;
    min-width: 20px;
    padding: 0 6px;
    border-radius: 10px;
    background: linear-gradient(135deg, #ff6b6b, #ff3b3b);
    color: #fff;
    font-size: 11px;
    font-weight: 700;
    line-height: 20px;
    text-align: center;
    flex-shrink: 0;
    box-shadow: 0 2px 4px rgba(255, 59, 59, 0.3);
}

.dock-overflow-divider {
    height: 0.5px;
    background: rgba(255, 255, 255, 0.1);
    margin: 4px 8px;
}

[data-theme="light"] .dock-overflow-divider {
    background: rgba(0, 0, 0, 0.1);
}

/* Popup transition */
.dock-popup-enter-active {
    transition: opacity 0.15s ease, transform 0.15s ease;
}
.dock-popup-leave-active {
    transition: opacity 0.1s ease, transform 0.1s ease;
}
.dock-popup-enter-from,
.dock-popup-leave-to {
    opacity: 0;
    transform: translateY(4px) scale(0.95);
}

/* ── 深色模式蓝色适配 ── */
[data-theme="dark"] .header-acp-btn {
    color: #1E3A5F;
}

[data-theme="dark"] .header-acp-btn:hover {
    background: color-mix(in srgb, #1E3A5F 10%, transparent);
}

[data-theme="dark"] .header-acp-btn:active {
    background: color-mix(in srgb, #1E3A5F 25%, transparent);
}

[data-theme="dark"] .radar-icon-wrapper.has-running {
    color: #1E3A5F !important;
}

</style>
