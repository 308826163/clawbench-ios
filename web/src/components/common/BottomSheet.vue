<template>
  <Teleport to="body">
    <div
      v-if="everOpened"
      v-show="open || leaving"
      class="bs-overlay"
      :class="{ 'bs-leaving': leaving, 'bs-instant': instant, 'bs-fullscreen': fullScreen }"
      @click.self="handleClose"
    >
      <div ref="panelRef" class="bs-panel" :class="[panelClass, { 'bs-leaving': leaving, 'bs-instant': instant, 'bs-compact': compact, 'bs-auto': auto, 'bs-handle-only': handleOnly, 'bs-not-ready': !ready }]" :style="dragOffset > 0 ? { transform: `translateY(${dragOffset}px)`, transition: isDragging ? 'none' : undefined } : {}">
        <!-- Header：拖拽手柄区域支持下滑收起 -->
        <div v-if="!noHeader" class="bs-header" :class="{ 'bs-header-handle-only': handleOnly }" @click="handleClose" @touchstart.passive="onDragStart" @touchmove.passive="onDragMove" @touchend="onDragEnd">
          <div class="bs-handle" />
          <slot v-if="!handleOnly" name="header">
            <span class="bs-title">{{ title }}</span>
          </slot>
        </div>
        <!-- Body -->
        <div class="bs-body">
          <slot />
        </div>
        <!-- Footer slot -->
        <footer v-if="$slots.footer" class="bs-footer">
          <slot name="footer" />
        </footer>
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch } from 'vue'

const props = defineProps({
  open: Boolean,
  title: {
    type: String,
    default: '',
  },
  instant: Boolean,  // 立即关闭，无动画
  compact: Boolean,  // 紧凑模式，高度自适应内容，最大50%，无圆角
  auto: Boolean,     // 自适应模式，高度按内容需要，最大全屏
  noHeader: Boolean, // 隐藏Header（含手柄）
  handleOnly: Boolean, // 仅显示拖拽手柄，无标题栏
  fullScreen: Boolean, // overlay 延伸到屏幕底部（覆盖导航栏）
  panelClass: String, // 自定义面板类名
  ready: {           // 面板是否就绪（false 时显示遮罩但不滑入面板）
    type: Boolean,
    default: true,
  },
})

const emit = defineEmits(['close'])

const leaving = ref(false)
const everOpened = ref(false)
let leaveTimer = null

// ── 下滑收起手势 ──
const panelRef = ref(null)
const dragOffset = ref(0)
const isDragging = ref(false)
let startY = 0
let startOffset = 0

function onDragStart(e) {
  if (leaving.value) return
  const touch = e.touches?.[0] || e
  startY = touch.clientY
  startOffset = dragOffset.value
  isDragging.value = true
}

function onDragMove(e) {
  if (!isDragging.value) return
  const touch = e.touches?.[0] || e
  const delta = touch.clientY - startY
  // 只允许下滑（正方向），不允许上滑
  dragOffset.value = Math.max(startOffset + delta, 0)
}

function onDragEnd() {
  if (!isDragging.value) return
  isDragging.value = false

  // 超过阈值则关闭，否则弹回
  const THRESHOLD = 100
  if (dragOffset.value > THRESHOLD) {
    handleClose()
  } else {
    dragOffset.value = 0
  }
}

watch(() => props.open, (val) => {
  clearTimeout(leaveTimer)
  dragOffset.value = 0
  if (val) {
    everOpened.value = true
    leaving.value = false
  } else if (leaving.value) {
    // Close triggered externally while animating — cancel animation, hide now
    leaving.value = false
  }
}, { immediate: true })

function handleClose() {
  if (leaving.value) return
  if (props.instant) {
    emit('close')
    return
  }
  leaving.value = true
  // 从当前位置直接滑出屏幕（不用 CSS 动画，用 transform + transition）
  const screenHeight = window.innerHeight
  dragOffset.value = screenHeight
  leaveTimer = setTimeout(() => {
    leaving.value = false
    dragOffset.value = 0
    leaveTimer = null
    emit('close')
  }, 300)
}

defineExpose({
  close: handleClose,
})
</script>

<style>
/* ═══════════════════════════════════════════════════════════════════════
 * BottomSheet - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.bs-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(4px);
  backdrop-filter: blur(4px);
  z-index: 1000;
  display: flex;
  align-items: flex-end;
  overflow: hidden;
  animation: bs-fadeIn 0.2s ease;
}

/* Light Theme Overlay */
[data-theme="light"] .bs-overlay {
  background: rgba(0, 0, 0, 0.3);
}

.bs-overlay.bs-leaving {
  animation: bs-fadeOut 0.25s ease forwards;
}

/* 全屏模式：overlay 延伸到屏幕底部（覆盖导航栏） */
.bs-overlay.bs-fullscreen {
  bottom: 0;
}

/* 全屏模式：panel 顶部避开状态栏 */
.bs-overlay.bs-fullscreen .bs-panel {
  top: 98px;
}

/* ── Panel (Liquid Glass - Dark Theme) ── */
.bs-panel {
  position: absolute;
  bottom: env(safe-area-inset-bottom, 0px);
  left: 0;
  right: 0;
  top: 0;

  /* Liquid Glass Material */
  background: linear-gradient(
    135deg,
    rgba(30, 30, 30, 0.85) 0%,
    rgba(20, 20, 20, 0.9) 50%,
    rgba(25, 25, 25, 0.88) 100%
  );
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  backdrop-filter: blur(20px) saturate(180%);

  /* Top highlight, bottom shadow */
  box-shadow:
    inset 0 0.5px 0 rgba(255, 255, 255, 0.12),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.06),
    0 -8px 32px rgba(0, 0, 0, 0.2);

  display: flex;
  flex-direction: column;
  overflow: hidden;
  transform: translateY(100%);
  transition: transform 0.25s ease;
}

/* 数据就绪后，滑入屏幕 */
.bs-panel:not(.bs-not-ready) {
  transform: translateY(0);
}

/* 数据未就绪时，面板隐藏在屏幕外（已是默认状态，无需额外样式） */

/* Light Theme Panel */
[data-theme="light"] .bs-panel {
  background: linear-gradient(
    135deg,
    rgba(255, 255, 255, 0.9) 0%,
    rgba(245, 245, 245, 0.92) 50%,
    rgba(250, 250, 250, 0.91) 100%
  );
  box-shadow:
    inset 0 0.5px 0 rgba(255, 255, 255, 0.8),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.04),
    0 -8px 32px rgba(0, 0, 0, 0.1);
}

/* Compact mode - auto height, rounded top corners */
.bs-panel.bs-compact {
  top: auto;
  height: auto;
  max-height: 50%;
  border-radius: 20px 20px 0 0;
}

.bs-panel.bs-compact .bs-header {
  border-radius: 20px 20px 0 0;
}

.bs-panel.bs-leaving {
  animation: none;
}

@keyframes bs-slideUp {
  from { transform: translateY(100%); }
  to   { transform: translateY(0); }
}

@keyframes bs-slideDown {
  from { transform: translateY(0); }
  to   { transform: translateY(100%); }
}

@keyframes bs-fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

@keyframes bs-fadeOut {
  from { opacity: 1; }
  to   { opacity: 0; }
}

/* Instant close (no animation) */
.bs-overlay.bs-instant {
  animation: none;
}

.bs-panel.bs-instant {
  animation: none;
}

.bs-overlay.bs-instant.bs-leaving {
  display: none;
}

.bs-panel.bs-instant.bs-leaving {
  display: none;
}

/* ── Header (Liquid Glass - Dark Theme) ── */
.bs-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 0 16px;
  height: 44px;
  border-bottom: none;
  box-shadow: 0 0.5px 0 rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.04);
  flex-shrink: 0;
  cursor: pointer;
  position: relative;
}

/* Light Theme Header */
[data-theme="light"] .bs-header {
  box-shadow: 0 0.5px 0 rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.02);
}

/* Drag handle (Liquid Glass style) */
.bs-handle {
  position: absolute;
  top: 8px;
  left: 50%;
  transform: translateX(-50%);
  width: 36px;
  height: 4px;
  border-radius: 2px;
  background: rgba(255, 255, 255, 0.3);
}

[data-theme="light"] .bs-handle {
  background: rgba(0, 0, 0, 0.2);
}

/* Handle-only header — compact, centered handle */
.bs-header-handle-only {
  justify-content: center;
  height: 20px;
  padding: 0;
  box-shadow: none;
  background: transparent;
}

.bs-header-handle-only .bs-handle {
  top: 8px;
}

.bs-header-icon {
  flex-shrink: 0;
  color: rgba(255, 255, 255, 0.9);
  display: flex;
  align-items: center;
}

[data-theme="light"] .bs-header-icon {
  color: rgba(0, 0, 0, 0.8);
}

.bs-header-title {
  font-weight: 600;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.9);
  flex-shrink: 0;
  white-space: nowrap;
}

[data-theme="light"] .bs-header-title {
  color: rgba(0, 0, 0, 0.85);
}

.bs-header-description {
  flex: 1;
  min-width: 0;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.5);
  white-space: nowrap;
  overflow: hidden;
  display: flex;
  align-items: center;
}

[data-theme="light"] .bs-header-description {
  color: rgba(0, 0, 0, 0.5);
}

/* ── Body ── */
.bs-body {
  flex: 1;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

/* Compact mode body - flex container for sticky tab bar */
.bs-panel.bs-compact .bs-body {
  overflow-y: hidden;
}

/* Auto mode - auto height based on content, max full screen */
.bs-panel.bs-auto {
  top: auto;
  height: auto;
  max-height: 100%;
  border-radius: 20px 20px 0 0;
}

.bs-panel.bs-auto .bs-header {
  border-radius: 20px 20px 0 0;
}

.bs-panel.bs-auto .bs-body {
  overflow-y: auto;
}

/* ── Footer (Liquid Glass) ── */
.bs-panel > .bs-footer {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  padding: 12px 16px;
  border-top: 0.5px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.02);
  flex-shrink: 0;
  gap: 8px;
}

[data-theme="light"] .bs-panel > .bs-footer {
  border-top: 0.5px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.01);
}

/* Compact mode footer — add bottom padding for dock bar clearance */
.bs-panel.bs-compact > .bs-footer {
    padding-bottom: calc(12px + env(safe-area-inset-bottom, 0px));
}
</style>
