<!--
  附件菜单组件
  功能：显示附件选项菜单，支持附加当前文件、当前目录、上传文件
  位置：文件浏览器头部（FileHeader）的附件按钮弹出菜单
-->

<template>
  <div class="attach-menu-wrapper" ref="wrapperRef">
    <!-- 附件按钮 -->
    <button
      class="file-header-btn attach-btn"
      :class="{ 'has-attachment': hasAttachments }"
      @click.stop="toggleMenu"
      :title="t('file.attach.title')"
    >
      <Paperclip :size="14" />
      <span v-if="attachmentCount > 0" class="attach-badge">{{ attachmentCount }}</span>
    </button>

    <!-- 附件菜单 -->
    <Teleport to="body">
      <div v-if="menuOpen" ref="menuRef" class="attach-menu" :style="menuStyle">
        <!-- 当前文件 -->
        <button
          v-if="currentFile"
          class="attach-menu-item"
          :class="{ 'is-attached': hasAttachedFile(currentFile.path) }"
          @click="handleAttachFile"
        >
          <FileText :size="14" />
          <div class="attach-menu-item-content">
            <span class="attach-menu-item-name">{{ t('file.attach.currentFile') }}</span>
            <span class="attach-menu-item-path">{{ currentFile.name }}</span>
          </div>
          <Check v-if="hasAttachedFile(currentFile.path)" :size="14" class="attach-check" />
        </button>

        <!-- 当前目录 -->
        <button
          v-if="currentDir"
          class="attach-menu-item"
          :class="{ 'is-attached': hasAttachedFile(currentDir) }"
          @click="handleAttachDir"
        >
          <Folder :size="14" />
          <div class="attach-menu-item-content">
            <span class="attach-menu-item-name">{{ t('file.attach.currentDir') }}</span>
            <span class="attach-menu-item-path">{{ currentDir }}</span>
          </div>
          <Check v-if="hasAttachedFile(currentDir)" :size="14" class="attach-check" />
        </button>

        <!-- 分隔线 -->
        <div v-if="currentFile || currentDir" class="attach-menu-separator"></div>

        <!-- 上传文件 -->
        <button class="attach-menu-item" @click="handleUploadFile">
          <Upload :size="14" />
          <div class="attach-menu-item-content">
            <span class="attach-menu-item-name">{{ t('file.attach.uploadFile') }}</span>
          </div>
        </button>

        <!-- 已附加列表 -->
        <template v-if="attachedFiles.length > 0">
          <div class="attach-menu-separator"></div>
          <div class="attach-menu-group-title">{{ t('file.attach.attachedList') }}</div>
          <div class="attach-menu-attached-list">
            <div
              v-for="(file, index) in attachedFiles"
              :key="file"
              class="attach-menu-attached-item"
            >
              <FileText :size="12" />
              <span class="attach-menu-attached-name" :title="file">{{ getFileName(file) }}</span>
              <button class="attach-menu-attached-remove" @click.stop="handleRemoveAttached(index)">
                <X :size="12" />
              </button>
            </div>
          </div>
          <button class="attach-menu-item clear-all" @click="handleClearAll">
            <Trash2 :size="14" />
            <span>{{ t('file.attach.clearAll') }}</span>
          </button>
        </template>
      </div>
    </Teleport>

    <!-- 隐藏的文件上传输入框 -->
    <input
      ref="fileInputRef"
      type="file"
      multiple
      style="display: none"
      @change="handleFileInputChange"
    />
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onBeforeUnmount, nextTick } from 'vue'
import { useI18n } from 'vue-i18n'
import { Paperclip, FileText, Folder, Upload, Check, X, Trash2 } from 'lucide-vue-next'
import { useChatContext } from '@/composables/useChatContext.ts'
import { store } from '@/stores/app.ts'

const { t } = useI18n()
const {
    attachedFiles,
    addAttachedFile,
    removeAttachedFile,
    clearAll,
} = useChatContext()

// 计算属性
const attachmentCount = computed(() => attachedFiles.value.length)
const hasAttachments = computed(() => attachedFiles.value.length > 0)

/**
 * 检查文件是否已附加
 * @param path 文件或目录的路径
 * @returns 是否已附加
 */
function hasAttachedFile(path) {
    return attachedFiles.value.includes(path)
}

// 当前文件和目录
const currentFile = computed(() => store.state.currentFile)
const currentDir = computed(() => store.state.currentDir)

// 菜单状态
const menuOpen = ref(false)
const wrapperRef = ref(null)
const menuRef = ref(null)
const menuStyle = ref({})
const fileInputRef = ref(null)

/**
 * 切换菜单显示/隐藏
 */
function toggleMenu() {
    menuOpen.value = !menuOpen.value
    if (menuOpen.value) {
        nextTick(() => updateMenuPosition())
    }
}

/**
 * 更新菜单位置（固定定位，跟随附件按钮）
 */
function updateMenuPosition() {
    if (!wrapperRef.value) return
    const rect = wrapperRef.value.getBoundingClientRect()
    menuStyle.value = {
        position: 'fixed',
        top: `${rect.bottom + 4}px`,
        right: `${window.innerWidth - rect.right}px`,
        left: 'auto',
    }
}

/**
 * 获取文件名（从完整路径中提取）
 * @param path 完整文件路径
 * @returns 文件名
 */
function getFileName(path) {
    if (!path) return ''
    const parts = path.split(/[/\\]/)
    return parts[parts.length - 1] || path
}

/**
 * 将文件路径转换为相对于项目根目录的完整路径
 * @param path 文件或目录的路径（可能是相对于currentDir的路径）
 * @returns 相对于项目根目录的完整路径
 */
function resolveFullPath(path) {
    if (!path) return path

    // 已经是绝对路径（包含盘符或以/开头），直接返回
    if (/^[A-Za-z]:[\\\/]/.test(path) || path.startsWith('/')) {
        return path
    }

    // 相对路径：与currentDir组合
    const dir = currentDir.value || ''
    if (dir) {
        // 如果dir是绝对路径，说明path是相对于dir的
        // 直接返回path（batch-exists API会用项目路径解析）
        if (/^[A-Za-z]:[\\\/]/.test(dir) || dir.startsWith('/')) {
            return path
        }
        // 如果dir是相对路径，将dir + "/" + path组合
        // 使其成为相对于项目根目录的完整路径
        return dir + '/' + path
    }

    // currentDir为空，path是相对于项目根目录的
    return path
}

/**
 * 附加当前文件到聊天
 */
function handleAttachFile() {
    if (currentFile.value?.path) {
        const fullPath = resolveFullPath(currentFile.value.path)
        addAttachedFile(fullPath)
    }
    menuOpen.value = false
}

/**
 * 附加当前目录到聊天
 */
function handleAttachDir() {
    if (currentDir.value) {
        addAttachedFile(currentDir.value)
    }
    menuOpen.value = false
}

/**
 * 打开文件选择器上传文件
 */
function handleUploadFile() {
    menuOpen.value = false
    fileInputRef.value?.click()
}

/**
 * 处理文件选择变化
 * @param event 文件输入框变化事件
 */
function handleFileInputChange(event) {
    const files = event.target.files
    if (!files || files.length === 0) return

    // 上传文件到当前目录
    // 注意：这里需要实现实际的上传逻辑，目前只是添加到附件列表
    // 实际上传需要调用后端 API: POST /api/upload/file
    console.log('Selected files:', files)

    // 清空输入框，允许重复选择同一文件
    event.target.value = ''
}

/**
 * 移除已附加的文件
 * @param index 附件在列表中的索引
 */
function handleRemoveAttached(index) {
    removeAttachedFile(index)
}

/**
 * 清空所有附件
 */
function handleClearAll() {
    clearAll()
    menuOpen.value = false
}

/**
 * 点击外部关闭菜单
 */
function handleClickOutside(e) {
    if (menuOpen.value &&
        wrapperRef.value && !wrapperRef.value.contains(e.target) &&
        (!menuRef.value || !menuRef.value.contains(e.target))) {
        menuOpen.value = false
    }
}

onMounted(() => {
    document.addEventListener('click', handleClickOutside)
})

onBeforeUnmount(() => {
    document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
.attach-menu-wrapper {
    position: relative;
}

/* 附件按钮 */
.attach-btn {
    position: relative;
}

.attach-btn.has-attachment {
    color: var(--accent-color);
}

/* 附件数量徽章 */
.attach-badge {
    position: absolute;
    top: -2px;
    right: -2px;
    min-width: 14px;
    height: 14px;
    padding: 0 3px;
    background: var(--accent-color);
    color: #fff;
    font-size: 10px;
    font-weight: 600;
    line-height: 14px;
    text-align: center;
    border-radius: 7px;
}
</style>

<!-- 非 scoped 样式：Teleport 到 body 的菜单 -->
<style>
/* ═══════════════════════════════════════════════════════════════════════
 * AttachMenu - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Menu (Liquid Glass - Dark Theme) ── */
.attach-menu {
    position: fixed;

    /* Liquid Glass Material */
    background: linear-gradient(
        135deg,
        rgba(30, 30, 30, 0.85) 0%,
        rgba(20, 20, 20, 0.9) 50%,
        rgba(25, 25, 25, 0.88) 100%
    );
    -webkit-backdrop-filter: blur(18px) saturate(180%);
    backdrop-filter: blur(18px) saturate(180%);

    /* No visible border - use light reflection */
    border: 0.5px solid rgba(255, 255, 255, 0.1);
    border-radius: 18px;

    /* Floating elevation */
    box-shadow:
        0 12px 40px rgba(0, 0, 0, 0.25),
        0 4px 12px rgba(0, 0, 0, 0.15),
        inset 0 0.5px 0 rgba(255, 255, 255, 0.12),
        inset 0 -0.5px 0 rgba(0, 0, 0, 0.04);

    z-index: 9999;
    min-width: 220px;
    max-width: 300px;
    padding: 6px;
    overflow: hidden;
}

/* Light Theme Menu */
[data-theme="light"] .attach-menu {
    background: linear-gradient(
        135deg,
        rgba(255, 255, 255, 0.9) 0%,
        rgba(245, 245, 245, 0.92) 50%,
        rgba(250, 250, 250, 0.91) 100%
    );
    border: 0.5px solid rgba(255, 255, 255, 0.5);
    box-shadow:
        0 12px 40px rgba(0, 0, 0, 0.1),
        0 4px 12px rgba(0, 0, 0, 0.06),
        inset 0 0.5px 0 rgba(255, 255, 255, 0.8),
        inset 0 -0.5px 0 rgba(0, 0, 0, 0.02);
}

/* ── Menu Item (Dark Theme) ── */
.attach-menu-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 10px 14px;
    width: 100%;
    border: none;
    border-radius: 12px;
    background: transparent;
    color: rgba(255, 255, 255, 0.8);
    font-size: 14px;
    font-weight: 500;
    cursor: pointer;
    text-decoration: none;
    white-space: nowrap;
    transition: all 0.2s ease;
}

[data-theme="light"] .attach-menu-item {
    color: rgba(0, 0, 0, 0.75);
}

.attach-menu-item:hover {
    background: rgba(255, 255, 255, 0.1);
    color: rgba(255, 255, 255, 0.95);
}

[data-theme="light"] .attach-menu-item:hover {
    background: rgba(0, 0, 0, 0.06);
    color: rgba(0, 0, 0, 0.95);
}

.attach-menu-item svg {
    flex-shrink: 0;
}

/* Attached state */
.attach-menu-item.is-attached {
    color: rgba(100, 180, 255, 0.9);
}

.attach-menu-item.is-attached:hover {
    color: rgba(255, 255, 255, 0.95);
}

[data-theme="light"] .attach-menu-item.is-attached {
    color: rgba(37, 99, 235, 0.9);
}

.attach-check {
    margin-left: auto;
    color: rgba(100, 180, 255, 0.9);
}

.attach-menu-item.is-attached:hover .attach-check {
    color: rgba(255, 255, 255, 0.95);
}

/* Menu item content */
.attach-menu-item-content {
    display: flex;
    flex-direction: column;
    min-width: 0;
}

.attach-menu-item-name {
    font-weight: 500;
}

.attach-menu-item-path {
    font-size: 12px;
    color: rgba(255, 255, 255, 0.5);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

[data-theme="light"] .attach-menu-item-path {
    color: rgba(0, 0, 0, 0.45);
}

.attach-menu-item:hover .attach-menu-item-path {
    color: rgba(255, 255, 255, 0.6);
}

[data-theme="light"] .attach-menu-item:hover .attach-menu-item-path {
    color: rgba(0, 0, 0, 0.6);
}

/* Separator */
.attach-menu-separator {
    height: 0.5px;
    background: rgba(255, 255, 255, 0.08);
    margin: 4px 8px;
}

[data-theme="light"] .attach-menu-separator {
    background: rgba(0, 0, 0, 0.06);
}

/* Group title */
.attach-menu-group-title {
    padding: 6px 14px;
    font-size: 11px;
    font-weight: 600;
    color: rgba(255, 255, 255, 0.4);
    text-transform: uppercase;
    letter-spacing: 0.5px;
}

[data-theme="light"] .attach-menu-group-title {
    color: rgba(0, 0, 0, 0.4);
}

/* Attached list */
.attach-menu-attached-list {
    max-height: 200px;
    overflow-y: auto;
    padding: 0 4px;
}

.attach-menu-attached-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 6px 10px;
    border-radius: 8px;
    font-size: 13px;
    color: rgba(255, 255, 255, 0.7);
}

[data-theme="light"] .attach-menu-attached-item {
    color: rgba(0, 0, 0, 0.65);
}

.attach-menu-attached-item:hover {
    background: rgba(255, 255, 255, 0.06);
}

[data-theme="light"] .attach-menu-attached-item:hover {
    background: rgba(0, 0, 0, 0.04);
}

.attach-menu-attached-name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.attach-menu-attached-remove {
    flex-shrink: 0;
    padding: 2px;
    border: none;
    background: none;
    color: rgba(255, 255, 255, 0.4);
    cursor: pointer;
    border-radius: 4px;
    opacity: 0;
    transition: all 0.15s ease;
}

[data-theme="light"] .attach-menu-attached-remove {
    color: rgba(0, 0, 0, 0.35);
}

.attach-menu-attached-item:hover .attach-menu-attached-remove {
    opacity: 1;
}

.attach-menu-attached-remove:hover {
    color: rgba(239, 68, 68, 0.9);
}

/* Clear all button */
.attach-menu-item.clear-all {
    color: rgba(239, 68, 68, 0.8);
}

.attach-menu-item.clear-all:hover {
    background: rgba(239, 68, 68, 0.15);
    color: rgba(239, 68, 68, 0.95);
}
</style>
