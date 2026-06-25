<!--
  文件内容预览弹窗组件
  功能：显示文件内容，支持关闭、跳转到文件浏览器
  位置：聊天页面，点击附件标签时弹出
-->

<template>
  <Teleport to="body">
    <Transition name="modal">
      <div v-if="visible" class="file-preview-overlay" @click.self="onClose">
        <div class="file-preview-modal">
          <!-- 头部：文件名 + 路径 + 关闭按钮 -->
          <div class="file-preview-header">
            <div class="file-preview-title">
              <span class="file-name">{{ fileName }}</span>
              <span class="file-path" @click="onPathClick" :title="filePath">
                📁 {{ filePath }}
              </span>
            </div>
            <button class="file-preview-close" @click="onClose" title="关闭">
              <X :size="16" />
            </button>
          </div>

          <!-- 调试信息 -->
          <div v-if="false" style="display:none">
            fileData: {{ fileData ? { name: fileData.name, hasContent: !!fileData.content, path: fileData.path } : 'null' }}
          </div>

          <!-- 操作按钮栏（复用原生 FileHeader） -->
          <FileHeader
            v-if="fileData && !loading && !error"
            :file="fileData"
            :view-mode="markdownViewMode"
            :toc-open="tocOpen"
            :search-open="searchOpen"
            :word-wrap="wordWrap"
            :show-line-numbers="showLineNumbers"
            :sticky-scroll="false"
            :overlay-open="false"
            :overlay-can-go-back="false"
            :show-attach-menu="false"
            @toggle-view="toggleView"
            @show-details="onClose"
            @toggle-toc="toggleToc"
            @toggle-search="toggleSearch"
            @open-as-text="() => {}"
            @toggle-word-wrap="onToggleWordWrap"
            @toggle-line-numbers="onToggleLineNumbers"
            @toggle-sticky-scroll="() => {}"
            @refresh="onRefresh"
            @overlay-close="onClose"
            @overlay-go-back="() => {}"
          />

          <!-- 内容区域 -->
          <div class="file-preview-content" ref="contentRef">
            <!-- 加载中 -->
            <div v-if="loading" class="file-preview-loading">
              <div class="spinner"></div>
            </div>

            <!-- 错误 -->
            <div v-else-if="error" class="file-preview-error">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
                <circle cx="12" cy="12" r="10"/>
                <line x1="12" y1="8" x2="12" y2="12"/>
                <line x1="12" y1="16" x2="12.01" y2="16"/>
              </svg>
              <span>{{ error }}</span>
            </div>

            <!-- PDF -->
            <PdfPreview
              v-else-if="fileType === 'pdf'"
              ref="pdfPreviewRef"
              :file="fileData"
            />

            <!-- 图片 -->
            <ImagePreview
              v-else-if="fileType === 'image'"
              :file="fileData"
            />

            <!-- 音频 -->
            <AudioPreview
              v-else-if="fileType === 'audio'"
              :file="fileData"
            />

            <!-- 视频 -->
            <VideoPreview
              v-else-if="fileType === 'video'"
              :file="fileData"
            />

            <!-- 代码/文本（使用原生 CodePreview 组件） -->
            <CodePreview
              v-else
              :content="fileContent"
              :language="fileLanguage"
              :file-path="props.filePath"
              :word-wrap="wordWrap"
              :show-line-numbers="showLineNumbers"
              :sticky-scroll="false"
            />
          </div>

          <!-- 内嵌目录面板（从屏幕下方弹出） -->
          <Transition name="toc-slide">
            <div v-if="tocOpen" class="file-preview-toc">
              <div class="toc-header">
                <List :size="16" class="toc-header-icon" />
                <span class="toc-header-title">{{ t('toc.title') }}</span>
                <button class="toc-close" @click="tocOpen = false">
                  <X :size="14" />
                </button>
              </div>
              <div class="toc-body">
                <div class="toc-search">
                  <Search :size="14" class="toc-search-icon" />
                  <input
                    v-model="tocSearchQuery"
                    class="toc-search-input"
                    :placeholder="t('toc.searchPlaceholder')"
                    @dblclick="tocSearchQuery = ''"
                  />
                </div>
                <div class="toc-list">
                  <div v-if="filteredToc.length === 0" class="toc-empty">
                    {{ tocSearchQuery ? t('toc.noMatch') : t('toc.noHeadings') }}
                  </div>
                  <a
                    v-for="item in filteredToc"
                    :key="item.id"
                    class="toc-item"
                    :class="{ active: tocActiveId === item.id }"
                    :data-level="item.level"
                    @click.prevent="scrollToTocItem(item)"
                  >
                    <span v-if="item.kind" class="toc-kind">{{ item.kind }}</span>
                    {{ item.text }}
                  </a>
                </div>
              </div>
            </div>
          </Transition>

          <!-- 内嵌搜索面板（从屏幕下方弹出） -->
          <Transition name="search-slide">
            <div v-if="searchOpen" class="file-preview-search">
              <div class="search-header">
                <Search :size="16" class="search-header-icon" />
                <span class="search-header-title">{{ t('search.title') }}</span>
                <button class="search-close" @click="searchOpen = false">
                  <X :size="14" />
                </button>
              </div>
              <div class="search-body">
                <div class="search-input-row">
                  <SearchInput
                    ref="searchInputRef"
                    v-model="searchQuery"
                    :placeholder="t('search.placeholder')"
                    @enter="jumpToFirstSearchResult"
                    @dblclick="searchQuery = ''"
                  />
                </div>
                <div class="search-content">
                  <div v-if="!fileContent" class="search-empty">{{ t('search.noContent') }}</div>
                  <div v-else-if="!searchQuery.trim()" class="search-empty">{{ t('search.enterKeyword') }}</div>
                  <div v-else-if="searchResults.length === 0" class="search-empty">{{ t('search.notFound', { query: searchQuery }) }}</div>
                  <div v-else class="search-results">
                    <div class="search-results-count">{{ t('search.matchCount', { count: searchResults.length }) }}</div>
                    <div
                      v-for="(r, idx) in searchResults"
                      :key="r.line"
                      class="search-result-item"
                      @click="jumpToSearchResult(r)"
                    >
                      <span class="search-result-lnum">{{ r.line }}</span>
                      <span class="search-result-text" v-html="r.highlighted" />
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </Transition>

          <!-- 文件信息栏（对称居中布局） -->
          <div class="file-preview-info">
            <span class="info-item info-left">
              <FileText :size="12" />
              {{ fileTypeLabel }}
            </span>
            <span class="info-item info-center">
              <Clock :size="12" />
              {{ fileTimeLabel }}
            </span>
            <span class="info-item info-right">
              <HardDrive :size="12" />
              {{ fileSizeLabel }}
            </span>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { X, RotateCw, TextWrap, Hash, Download, FileText, HardDrive, Clock, List, Search } from 'lucide-vue-next'
import FileHeader from '@/components/file/FileHeader.vue'
import CodePreview from '@/components/file/CodePreview.vue'
import SearchInput from '@/components/common/SearchInput.vue'
import ImagePreview from '@/components/media/ImagePreview.vue'
import PdfPreview from '@/components/media/PdfPreview.vue'
import AudioPreview from '@/components/media/AudioPreview.vue'
import VideoPreview from '@/components/media/VideoPreview.vue'
import { extractToc, slugify } from '@/utils/toc.ts'
import { getFileType } from '@/utils/fileType.ts'
import { fetchCodeSymbols } from '@/composables/useCodeSymbols'
import { escapeHtml } from '@/utils/html.ts'
import { searchRawContent, highlightText, BLOCK_TAGS } from '@/utils/searchUtils.ts'

const { t } = useI18n()

const props = defineProps({
  visible: Boolean,
  filePath: String,
})

const emit = defineEmits(['close', 'path-click'])

// 状态
const loading = ref(false)
const error = ref('')
const fileContent = ref('')
const fileData = ref(null)
const wordWrap = ref(false)
const showLineNumbers = ref(true)
const contentRef = ref(null)
const markdownViewMode = ref('rendered')
const pdfPreviewRef = ref(null)

// 目录状态
const tocOpen = ref(false)
const tocItems = ref([])
const tocActiveId = ref('')
const tocSearchQuery = ref('')
const filteredToc = ref([])

// 搜索状态
const searchOpen = ref(false)
const searchQuery = ref('')
const searchInputRef = ref(null)
const searchResults = ref([])

// 计算属性
const fileLanguage = computed(() => {
  if (!props.filePath) return 'plaintext'
  return getFileType(props.filePath)?.lang || 'plaintext'
})

// 计算属性
const fileName = computed(() => {
  if (!props.filePath) return ''
  const parts = props.filePath.split(/[/\\]/)
  return parts[parts.length - 1] || props.filePath
})

const fileType = computed(() => {
  if (!props.filePath) return 'text'
  const ext = props.filePath.split('.').pop()?.toLowerCase() || ''
  const imageExts = ['png', 'jpg', 'jpeg', 'gif', 'webp', 'svg', 'bmp', 'ico', 'tiff', 'tif', 'avif']
  const audioExts = ['mp3', 'wav', 'ogg', 'm4a', 'aac', 'flac', 'wma', 'opus']
  const videoExts = ['mp4', 'mkv', 'avi', 'mov', 'webm', 'flv', 'wmv', 'm4v', '3gp', 'm3u8']
  if (ext === 'pdf') return 'pdf'
  if (imageExts.includes(ext)) return 'image'
  if (audioExts.includes(ext)) return 'audio'
  if (videoExts.includes(ext)) return 'video'
  return 'text'
})

const fileTypeLabel = computed(() => {
  if (!props.filePath) return ''
  const ext = props.filePath.split('.').pop()?.toUpperCase() || 'TEXT'
  return ext
})

const fileSizeLabel = computed(() => {
  if (!fileData.value?.size) return '-'
  const size = fileData.value.size
  if (size < 1024) return size + ' B'
  if (size < 1024 * 1024) return (size / 1024).toFixed(1) + ' KB'
  return (size / (1024 * 1024)).toFixed(1) + ' MB'
})

const fileTimeLabel = computed(() => {
  // 暂时显示当前时间，后续可以从API获取
  const now = new Date()
  return now.toLocaleString('zh-CN', { year: 'numeric', month: '2-digit', day: '2-digit', hour: '2-digit', minute: '2-digit' })
})

// 监听visible变化，加载文件内容
watch(() => props.visible, async (val) => {
  if (val && props.filePath) {
    await loadFileContent()
  } else {
    // 关闭时清空内容
    fileContent.value = ''
    fileData.value = null
    error.value = ''
  }
})

// 加载文件内容
async function loadFileContent() {
  if (!props.filePath) return

  loading.value = true
  error.value = ''
  fileContent.value = ''
  fileData.value = null

  try {
    // 根据文件类型处理
    if (fileType.value === 'image' || fileType.value === 'audio' || fileType.value === 'video' || fileType.value === 'pdf') {
      // 媒体文件：直接设置路径
      fileData.value = { path: props.filePath, name: fileName.value }
      loading.value = false
      return
    }

    // 文本文件：加载内容
    const cleanPath = props.filePath.replace(/^\/+/, '')
    const url = `/api/file/${encodeURIComponent(cleanPath)}`
    const resp = await fetch(url)

    if (!resp.ok) {
      const data = await resp.json()
      error.value = data.error || '加载失败'
      loading.value = false
      return
    }

    const data = await resp.json()
    fileContent.value = data.content || ''
    fileData.value = data
    loading.value = false
  } catch (err) {
    error.value = err.message || '加载失败'
    loading.value = false
  }
}

// 关闭弹窗
function onClose() {
  // 重置目录和搜索面板状态
  tocOpen.value = false
  searchOpen.value = false
  emit('close')
}

// 点击路径
function onPathClick() {
  emit('path-click', props.filePath)
}

// 刷新
function onRefresh() {
  loadFileContent()
}

// 切换自动换行
function onToggleWordWrap() {
  wordWrap.value = !wordWrap.value
}

// 切换行号
function onToggleLineNumbers() {
  showLineNumbers.value = !showLineNumbers.value
}

// 切换视图模式（源码/渲染）
function toggleView() {
  markdownViewMode.value = markdownViewMode.value === 'rendered' ? 'raw' : 'rendered'
}

// 切换目录
function toggleToc() {
  tocOpen.value = !tocOpen.value
  if (tocOpen.value) {
    // 打开目录时关闭搜索
    searchOpen.value = false
    if (tocItems.value.length === 0) {
      loadToc()
    }
  }
}

// 切换搜索
function toggleSearch() {
  searchOpen.value = !searchOpen.value
  if (searchOpen.value) {
    // 打开搜索时关闭目录
    tocOpen.value = false
  }
}

// 加载目录
async function loadToc() {
  if (!props.filePath || !fileContent.value) {
    tocItems.value = []
    filteredToc.value = []
    return
  }

  const lang = fileLanguage.value
  const isCode = lang !== 'markdown'

  // 尝试从后端获取代码符号
  try {
    const result = await fetchCodeSymbols(props.filePath)
    if (result && result.symbols.length > 0) {
      tocItems.value = result.symbols.map(s => ({
        level: s.level,
        text: s.name,
        kind: s.kind,
        id: s.kind === 'heading' ? slugify(s.name) : 'toc-l' + s.line,
        line: s.line,
      }))
    } else {
      // 回退到正则提取
      tocItems.value = extractToc(fileContent.value, lang)
    }
  } catch {
    // 回退到正则提取
    tocItems.value = extractToc(fileContent.value, lang)
  }

  tocActiveId.value = tocItems.value[0]?.id || ''
  tocSearchQuery.value = ''
  filteredToc.value = tocItems.value
}

// 目录搜索
watch(tocSearchQuery, () => {
  const query = tocSearchQuery.value.toLowerCase().trim()
  if (!query) {
    filteredToc.value = tocItems.value
    return
  }
  filteredToc.value = tocItems.value.filter(item =>
    item.text.toLowerCase().includes(query)
  )
})

// 目录跳转
function scrollToTocItem(item) {
  // PDF: 跳转到页码
  if (fileType.value === 'pdf' && item.line > 0) {
    if (pdfPreviewRef.value) {
      pdfPreviewRef.value.scrollToPage(item.line)
    }
    tocActiveId.value = item.id
    tocOpen.value = false
    return
  }

  // 代码/文本: 使用 id 查找元素
  const elById = document.getElementById(item.id)
  if (elById) {
    elById.scrollIntoView({ behavior: 'smooth', block: 'start' })
    elById.classList.add('line-flash')
    elById.addEventListener('animationend', () => elById.classList.remove('line-flash'), { once: true })
    tocActiveId.value = item.id
    tocOpen.value = false
    return
  }

  // 回退: 使用行号计算滚动位置
  if (item.line && contentRef.value) {
    const lineHeight = 20 // 假设每行高度为20px
    contentRef.value.scrollTop = (item.line - 1) * lineHeight
    tocActiveId.value = item.id
    tocOpen.value = false
  }
}

// 搜索跳转
function onSearchJump(line) {
  if (line && contentRef.value) {
    const lineHeight = 20 // 假设每行高度为20px
    contentRef.value.scrollTop = (line - 1) * lineHeight
  }
  searchOpen.value = false
}

// 搜索结果计算
const searchResultsComputed = computed(() => {
  if (!fileContent.value || !searchQuery.value.trim()) return []
  const q = searchQuery.value.trim()
  return searchRawContent(q, fileContent.value, fileData.value?.name || '')
})

// 监听搜索结果变化
watch(searchResultsComputed, (val) => {
  searchResults.value = val
})

// 搜索跳转到结果
function jumpToSearchResult(result) {
  if (result.line && contentRef.value) {
    const lineHeight = 20 // 假设每行高度为20px
    contentRef.value.scrollTop = (result.line - 1) * lineHeight
  }
  searchOpen.value = false
}

// 跳转到第一个搜索结果
function jumpToFirstSearchResult() {
  if (searchResults.value.length > 0) {
    jumpToSearchResult(searchResults.value[0])
  }
}

// 监听搜索面板打开状态
watch(searchOpen, async (val) => {
  if (val) {
    await nextTick()
    searchInputRef.value?.focus()
  }
})

// 下载
function onDownload() {
  if (!props.filePath) return
  const cleanPath = props.filePath.replace(/^\/+/, '')
  const url = `/api/local-file/${encodeURIComponent(cleanPath)}?download=1`
  window.open(url, '_blank')
}

// 键盘事件：ESC关闭
function onKeyDown(e) {
  if (e.key === 'Escape' && props.visible) {
    onClose()
  }
}

onMounted(() => {
  document.addEventListener('keydown', onKeyDown)
})

onUnmounted(() => {
  document.removeEventListener('keydown', onKeyDown)
})
</script>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════
 * FilePreviewModal - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.file-preview-overlay {
  position: fixed;
  inset: 0;
  z-index: 9999;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
}

/* Light Theme Overlay */
[data-theme="light"] .file-preview-overlay {
  background: rgba(0, 0, 0, 0.3);
}

/* ── Modal (Liquid Glass - Dark Theme) ── */
.file-preview-modal {
  width: 90%;
  max-width: 600px;
  max-height: 80vh;

  /* Liquid Glass Material */
  background: linear-gradient(
    135deg,
    rgba(30, 30, 30, 0.88) 0%,
    rgba(20, 20, 20, 0.92) 50%,
    rgba(25, 25, 25, 0.9) 100%
  );
  -webkit-backdrop-filter: blur(24px) saturate(180%);
  backdrop-filter: blur(24px) saturate(180%);

  /* No visible border - use light reflection */
  border: 0.5px solid rgba(255, 255, 255, 0.1);
  border-radius: 24px;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  /* Floating elevation */
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.3),
    0 8px 24px rgba(0, 0, 0, 0.2),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.15),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);
}

/* Light Theme Modal */
[data-theme="light"] .file-preview-modal {
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.92) 0%,
    rgba(245, 245, 245, 0.95) 50%,
    rgba(250, 250, 250, 0.93) 100%
  );
  border: 0.5px solid rgba(255, 255, 255, 0.6);
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.12),
    0 8px 24px rgba(0, 0, 0, 0.08),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.9),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.03);
}

/* ── Header (Liquid Glass) ── */
.file-preview-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 14px 20px;
  border-bottom: 0.5px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.03);
}

[data-theme="light"] .file-preview-header {
  border-bottom: 0.5px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.02);
}

.file-preview-title {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1;
}

.file-name {
  font-size: 15px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  white-space: nowrap;
}

[data-theme="light"] .file-name {
  color: rgba(0, 0, 0, 0.85);
}

.file-path {
  font-size: 11px;
  color: #666;
  cursor: pointer;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 300px;
}

.file-path:hover {
  color: var(--accent-color, #4a90d9);
  text-decoration: underline;
}

[data-theme="dark"] .file-path {
  color: #999;
}

.file-preview-close {
  flex-shrink: 0;
  padding: 4px;
  border: none;
  background: transparent;
  color: #666;
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.15s, color 0.15s;
}

.file-preview-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #333;
}

[data-theme="dark"] .file-preview-close {
  color: #999;
}

[data-theme="dark"] .file-preview-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #eee;
}

/* 操作按钮栏 */
.file-preview-actions {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 6px 12px;
  border-bottom: 1px solid rgba(0, 0, 0, 0.1);
}

[data-theme="dark"] .file-preview-actions {
  border-bottom-color: rgba(255, 255, 255, 0.1);
}

.file-preview-action-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  border: none;
  background: transparent;
  color: #666;
  cursor: pointer;
  border-radius: 4px;
  font-size: 11px;
  transition: background 0.15s, color 0.15s;
}

.file-preview-action-btn:hover {
  background: rgba(0, 0, 0, 0.1);
  color: #333;
}

[data-theme="dark"] .file-preview-action-btn {
  color: #999;
}

[data-theme="dark"] .file-preview-action-btn:hover {
  background: rgba(255, 255, 255, 0.1);
  color: #eee;
}

.action-check {
  color: var(--accent-color, #4a90d9);
  font-weight: 700;
}

.file-preview-action-spacer {
  flex: 1;
}

/* 内容区域 */
.file-preview-content {
  flex: 1;
  min-height: 200px;
  max-height: 50vh;
  overflow: auto;
  padding: 12px;
}

.file-preview-loading {
  display: flex;
  align-items: center;
  justify-content: center;
  height: 200px;
}

.spinner {
  width: 24px;
  height: 24px;
  border: 2px solid rgba(0, 0, 0, 0.1);
  border-top-color: var(--accent-color, #4a90d9);
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.file-preview-error {
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 8px;
  height: 200px;
  color: #dc3545;
  font-size: 13px;
}

.file-preview-code {
  font-family: 'Menlo', 'Monaco', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.5;
}

.file-preview-code pre {
  margin: 0;
  white-space: pre-wrap;
  word-break: break-all;
}

.file-preview-code code {
  color: #333;
}

[data-theme="dark"] .file-preview-code code {
  color: #eee;
}

.code-with-lines {
  counter-reset: line;
}

.code-with-lines code {
  counter-increment: line;
}

/* 文件信息栏（对称居中布局） */
.file-preview-info {
  display: flex;
  align-items: center;
  padding: 4px 16px;
  border-top: 1px solid rgba(0, 0, 0, 0.1);
  font-size: 10px;
  color: #999;
  line-height: 1.2;
}

[data-theme="dark"] .file-preview-info {
  border-top-color: rgba(255, 255, 255, 0.1);
}

.info-item {
  display: flex;
  align-items: center;
  gap: 4px;
}

.info-left {
  flex: 1;
  justify-content: flex-start;
}

.info-center {
  flex: 1;
  justify-content: center;
}

.info-right {
  flex: 1;
  justify-content: flex-end;
}

/* 目录面板 */
.file-preview-toc {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  max-height: 50%;
  background: var(--bg-secondary, #fff);
  border-top: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
  border-radius: 12px 12px 0 0;
  display: flex;
  flex-direction: column;
  z-index: 10;
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.1);
}

[data-theme="dark"] .file-preview-toc {
  background: var(--bg-secondary, #1e1e1e);
  border-top-color: var(--border-color, rgba(255, 255, 255, 0.1));
}

.toc-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
}

[data-theme="dark"] .toc-header {
  border-bottom-color: var(--border-color, rgba(255, 255, 255, 0.1));
}

.toc-header-icon {
  color: var(--text-secondary, #666);
}

[data-theme="dark"] .toc-header-icon {
  color: var(--text-secondary, #999);
}

.toc-header-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #333);
}

[data-theme="dark"] .toc-header-title {
  color: var(--text-primary, #eee);
}

.toc-close {
  padding: 4px;
  border: none;
  background: transparent;
  color: var(--text-secondary, #666);
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.15s, color 0.15s;
}

.toc-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-primary, #333);
}

[data-theme="dark"] .toc-close {
  color: var(--text-secondary, #999);
}

[data-theme="dark"] .toc-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary, #eee);
}

.toc-body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.toc-search {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
}

[data-theme="dark"] .toc-search {
  border-bottom-color: var(--border-color, rgba(255, 255, 255, 0.1));
}

.toc-search-icon {
  color: var(--text-muted, #999);
  flex-shrink: 0;
}

.toc-search-input {
  flex: 1;
  border: none;
  background: transparent;
  font-size: 13px;
  color: var(--text-primary, #333);
  outline: none;
}

.toc-search-input::placeholder {
  color: var(--text-muted, #999);
}

[data-theme="dark"] .toc-search-input {
  color: var(--text-primary, #eee);
}

.toc-list {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 8px 0;
}

.toc-empty {
  padding: 16px;
  text-align: center;
  font-size: 13px;
  color: var(--text-muted, #999);
}

.toc-item {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 16px;
  font-size: 13px;
  color: var(--text-secondary, #666);
  cursor: pointer;
  transition: background 0.15s, color 0.15s;
  text-decoration: none;
}

.toc-item:hover {
  background: rgba(0, 0, 0, 0.05);
  color: var(--text-primary, #333);
}

.toc-item.active {
  background: rgba(0, 0, 0, 0.1);
  color: var(--accent-color, #4a90d9);
  font-weight: 500;
}

[data-theme="dark"] .toc-item {
  color: var(--text-secondary, #999);
}

[data-theme="dark"] .toc-item:hover {
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary, #eee);
}

[data-theme="dark"] .toc-item.active {
  background: rgba(255, 255, 255, 0.1);
  color: var(--accent-color, #4a90d9);
}

.toc-item[data-level="2"] {
  padding-left: 32px;
}

.toc-item[data-level="3"] {
  padding-left: 48px;
}

.toc-item[data-level="4"] {
  padding-left: 64px;
}

.toc-kind {
  font-size: 10px;
  padding: 1px 4px;
  border-radius: 3px;
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-muted, #999);
}

[data-theme="dark"] .toc-kind {
  background: rgba(255, 255, 255, 0.1);
}

/* 搜索面板 */
.file-preview-search {
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  max-height: 50%;
  background: var(--bg-secondary, #fff);
  border-top: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
  border-radius: 12px 12px 0 0;
  display: flex;
  flex-direction: column;
  z-index: 10;
  box-shadow: 0 -4px 16px rgba(0, 0, 0, 0.1);
}

[data-theme="dark"] .file-preview-search {
  background: var(--bg-secondary, #1e1e1e);
  border-top-color: var(--border-color, rgba(255, 255, 255, 0.1));
}

.search-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 1px solid var(--border-color, rgba(0, 0, 0, 0.1));
}

[data-theme="dark"] .search-header {
  border-bottom-color: var(--border-color, rgba(255, 255, 255, 0.1));
}

.search-header-icon {
  color: var(--text-secondary, #666);
}

[data-theme="dark"] .search-header-icon {
  color: var(--text-secondary, #999);
}

.search-header-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: var(--text-primary, #333);
}

[data-theme="dark"] .search-header-title {
  color: var(--text-primary, #eee);
}

.search-close {
  padding: 4px;
  border: none;
  background: transparent;
  color: var(--text-secondary, #666);
  cursor: pointer;
  border-radius: 4px;
  transition: background 0.15s, color 0.15s;
}

.search-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: var(--text-primary, #333);
}

[data-theme="dark"] .search-close {
  color: var(--text-secondary, #999);
}

[data-theme="dark"] .search-close:hover {
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-primary, #eee);
}

.search-body {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-input-row {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 10px 14px;
  border-bottom: 1px solid var(--border-color, #e5e5e5);
  background: var(--bg-secondary, #f8f9fa);
  flex-shrink: 0;
}

[data-theme="dark"] .search-input-row {
  border-bottom-color: var(--border-color, rgba(255, 255, 255, 0.1));
  background: var(--bg-secondary, #2d2d2d);
}

.search-input-row :deep(.search-pill) {
  flex: 1;
}

.search-content {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.search-empty {
  padding: 24px;
  text-align: center;
  color: var(--text-muted, #999);
  font-size: 13px;
  flex-shrink: 0;
}

.search-results {
  flex: 1;
  overflow-y: auto;
}

.search-results-count {
  padding: 6px 14px;
  font-size: 11px;
  color: var(--text-muted, #999);
  border-bottom: 1px solid var(--border-color, #e5e5e5);
  background: var(--bg-secondary, #f8f9fa);
  flex-shrink: 0;
}

[data-theme="dark"] .search-results-count {
  border-bottom-color: var(--border-color, rgba(255, 255, 255, 0.1));
  background: var(--bg-secondary, #2d2d2d);
}

.search-result-item {
  display: flex;
  align-items: baseline;
  gap: 10px;
  padding: 5px 14px;
  cursor: pointer;
  font-family: 'SF Mono', 'Fira Code', Menlo, Monaco, 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.6;
  border-bottom: 1px solid var(--border-color, #f0f0f0);
  transition: background 0.1s;
}

[data-theme="dark"] .search-result-item {
  border-bottom-color: var(--border-color, rgba(255, 255, 255, 0.05));
}

.search-result-item:hover {
  background: var(--bg-secondary, #f8f9fa);
}

[data-theme="dark"] .search-result-item:hover {
  background: var(--bg-secondary, #2d2d2d);
}

.search-result-lnum {
  color: var(--text-muted, #999);
  min-width: 32px;
  text-align: right;
  flex-shrink: 0;
  user-select: none;
}

.search-result-text {
  white-space: pre-wrap;
  word-break: break-all;
}

.search-result-text :deep(em) {
  font-style: normal;
}

.search-result-text :deep(mark) {
  background: rgba(255, 230, 0, 0.5);
  color: inherit;
  border-radius: 2px;
  padding: 0 1px;
}

[data-theme="dark"] .search-result-text :deep(mark) {
  background: rgba(255, 230, 0, 0.35);
}

/* 目录滑入动效 */
.toc-slide-enter-active {
  animation: toc-slide-in 0.3s ease-out;
}

.toc-slide-leave-active {
  animation: toc-slide-out 0.2s ease-in;
}

@keyframes toc-slide-in {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

@keyframes toc-slide-out {
  from {
    transform: translateY(0);
  }
  to {
    transform: translateY(100%);
  }
}

/* 搜索滑入动效 */
.search-slide-enter-active {
  animation: search-slide-in 0.3s ease-out;
}

.search-slide-leave-active {
  animation: search-slide-out 0.2s ease-in;
}

@keyframes search-slide-in {
  from {
    transform: translateY(100%);
  }
  to {
    transform: translateY(0);
  }
}

@keyframes search-slide-out {
  from {
    transform: translateY(0);
  }
  to {
    transform: translateY(100%);
  }
}

/* 苹果风格动效 */
.modal-enter-active {
  transition: opacity 0.3s ease-out;
}

.modal-enter-active .file-preview-modal {
  animation: modal-in 0.3s ease-out;
}

.modal-leave-active {
  transition: opacity 0.2s ease-in;
}

.modal-leave-active .file-preview-modal {
  animation: modal-out 0.2s ease-in;
}

@keyframes modal-in {
  from {
    opacity: 0;
    transform: scale(0.9) translateY(20px);
  }
  to {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
}

@keyframes modal-out {
  from {
    opacity: 1;
    transform: scale(1) translateY(0);
  }
  to {
    opacity: 0;
    transform: scale(0.9) translateY(20px);
  }
}

/* 行高亮动画 */
:deep(.line-flash) {
  animation: line-flash 0.8s ease-out;
}

@keyframes line-flash {
  0% {
    background-color: rgba(74, 144, 217, 0.3);
  }
  100% {
    background-color: transparent;
  }
}
</style>
