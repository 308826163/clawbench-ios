<template>
  <div
    class="floating-input-float"
    :class="{ collapsed: isCollapsed, dragging: isDragging }"
    :style="floatStyle"
    @touchstart.passive="onTouchStart"
    @touchmove.passive="onTouchMove"
    @touchend="onTouchEnd"
    @mousedown="onMouseDown"
  >
    <!-- 收起状态：小圆形图标 -->
    <div v-if="isCollapsed" class="float-icon" @click="expand">
      <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
        <path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"></path>
      </svg>
    </div>

    <!-- 展开状态：完整输入框 -->
    <div v-else class="float-input-box">
      <!-- 拖动手柄 + 收起按钮 -->
      <div class="float-input-header">
        <div class="drag-handle">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="currentColor">
            <circle cx="8" cy="6" r="2"/>
            <circle cx="16" cy="6" r="2"/>
            <circle cx="8" cy="12" r="2"/>
            <circle cx="16" cy="12" r="2"/>
            <circle cx="8" cy="18" r="2"/>
            <circle cx="16" cy="18" r="2"/>
          </svg>
        </div>
        <button class="collapse-btn" @click.stop="collapse">
          <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M6 9l6 6 6-6"></path>
          </svg>
        </button>
      </div>

      <!-- 输入区域 -->
      <div class="float-input-content">
        <textarea
          ref="textareaRef"
          v-model="inputText"
          :placeholder="placeholder"
          @keydown.enter.exact.prevent="handleSend"
          @focus="onFocus"
          @blur="onBlur"
        ></textarea>
        <button class="float-send-btn" @click="handleSend" :disabled="!inputText.trim()">
          <svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
            <path d="M22 2L11 13"></path>
            <path d="M22 2L15 22L11 13L2 9L22 2Z"></path>
          </svg>
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { Keyboard } from '@capacitor/keyboard'

const props = defineProps({
  placeholder: { type: String, default: '输入消息...' }
})

const emit = defineEmits(['send'])

// 状态
const isCollapsed = ref(true)
const isDragging = ref(false)
const inputText = ref('')
const textareaRef = ref(null)

// 位置状态
const position = ref({ x: 0, y: 0 })
const collapsedPosition = ref({ x: 0, y: 0 })

// 拖动状态
let startX = 0
let startY = 0
let startPosX = 0
let startPosY = 0
let dragStartTime = 0
let hasMoved = false

// 屏幕尺寸
const screenWidth = ref(window.innerWidth)
const screenHeight = ref(window.innerHeight)

// 计算样式
const floatStyle = computed(() => {
  if (isCollapsed.value) {
    return {
      transform: `translate(${collapsedPosition.value.x}px, ${collapsedPosition.value.y}px)`
    }
  } else {
    return {
      transform: `translate(0px, ${position.value.y}px)`
    }
  }
})

// 初始化位置
onMounted(() => {
  loadPosition()
  window.addEventListener('resize', onResize)
  Keyboard.addListener('keyboardWillShow', onKeyboardShow)
  Keyboard.addListener('keyboardDidHide', onKeyboardHide)
})

onUnmounted(() => {
  window.removeEventListener('resize', onResize)
  Keyboard.removeAllListeners()
})

// 从 localStorage 加载位置
function loadPosition() {
  try {
    const saved = localStorage.getItem('floatingInputPosition')
    if (saved) {
      const data = JSON.parse(saved)
      if (data.collapsed) {
        collapsedPosition.value = data.collapsed
      }
      if (data.expanded) {
        position.value = data.expanded
      }
    } else {
      // 默认位置：右下角
      collapsedPosition.value = {
        x: screenWidth.value - 72,
        y: screenHeight.value - 150
      }
      position.value = { x: 0, y: 0 }
    }
  } catch {
    collapsedPosition.value = {
      x: screenWidth.value - 72,
      y: screenHeight.value - 150
    }
    position.value = { x: 0, y: 0 }
  }
}

// 保存位置到 localStorage
function savePosition() {
  try {
    localStorage.setItem('floatingInputPosition', JSON.stringify({
      collapsed: collapsedPosition.value,
      expanded: position.value
    }))
  } catch {}
}

// 展开
function expand() {
  if (hasMoved) return
  isCollapsed.value = false
  nextTick(() => {
    textareaRef.value?.focus()
  })
}

// 收起
function collapse() {
  isCollapsed.value = true
  savePosition()
}

// 发送消息
function handleSend() {
  if (inputText.value.trim()) {
    emit('send', inputText.value.trim())
    inputText.value = ''
  }
}

// ── 触摸拖动 ──

function onTouchStart(e) {
  if (e.touches.length !== 1) return
  const touch = e.touches[0]
  startDrag(touch.clientX, touch.clientY)
}

function onTouchMove(e) {
  if (!isDragging.value) return
  const touch = e.touches[0]
  moveDrag(touch.clientX, touch.clientY)
}

function onTouchEnd() {
  endDrag()
}

// ── 鼠标拖动（桌面端） ──

function onMouseDown(e) {
  startDrag(e.clientX, e.clientY)
  document.addEventListener('mousemove', onMouseMove)
  document.addEventListener('mouseup', onMouseUp)
}

function onMouseMove(e) {
  if (!isDragging.value) return
  moveDrag(e.clientX, e.clientY)
}

function onMouseUp() {
  endDrag()
  document.removeEventListener('mousemove', onMouseMove)
  document.removeEventListener('mouseup', onMouseUp)
}

// ── 拖动逻辑 ──

function startDrag(clientX, clientY) {
  isDragging.value = true
  hasMoved = false
  dragStartTime = Date.now()
  startX = clientX
  startY = clientY
  startPosX = isCollapsed.value ? collapsedPosition.value.x : 0
  startPosY = isCollapsed.value ? collapsedPosition.value.y : position.value.y
}

function moveDrag(clientX, clientY) {
  const dx = clientX - startX
  const dy = clientY - startY

  // 判断是否移动超过阈值
  if (Math.abs(dx) > 5 || Math.abs(dy) > 5) {
    hasMoved = true
  }

  if (isCollapsed.value) {
    // 收起状态：随意移动
    let newX = startPosX + dx
    let newY = startPosY + dy

    // 限制在屏幕范围内
    newX = Math.max(0, Math.min(newX, screenWidth.value - 56))
    newY = Math.max(0, Math.min(newY, screenHeight.value - 56))

    collapsedPosition.value = { x: newX, y: newY }
  } else {
    // 展开状态：只允许纵向移动
    let newY = startPosY + dy

    // 限制在屏幕范围内
    newY = Math.max(-screenHeight.value + 200, Math.min(newY, screenHeight.value - 200))

    position.value = { ...position.value, y: newY }
  }
}

function endDrag() {
  isDragging.value = false

  if (isCollapsed.value && hasMoved) {
    // 收起状态：吸附到最近的边缘
    snapToEdge()
  }

  savePosition()
}

// ── 吸附边缘 ──

function snapToEdge() {
  const { x, y } = collapsedPosition.value
  const iconSize = 56
  const margin = 16

  // 计算距离四个边缘的距离
  const distLeft = x
  const distRight = screenWidth.value - x - iconSize
  const distTop = y
  const distBottom = screenHeight.value - y - iconSize

  // 找到最近的边缘
  const minDist = Math.min(distLeft, distRight, distTop, distBottom)

  let targetX = x
  let targetY = y

  if (minDist === distLeft) {
    targetX = margin
  } else if (minDist === distRight) {
    targetX = screenWidth.value - iconSize - margin
  } else if (minDist === distTop) {
    targetY = margin
  } else {
    targetY = screenHeight.value - iconSize - margin
  }

  // 动画过渡到目标位置
  collapsedPosition.value = { x: targetX, y: targetY }
}

// ── 键盘处理 ──

function onKeyboardShow(res) {
  if (isCollapsed.value) return
  const kbHeight = res.keyboardHeight
  if (!kbHeight) return

  // 展开状态：向上移动输入框
  position.value = { ...position.value, y: -kbHeight }
}

function onKeyboardHide() {
  if (isCollapsed.value) return
  // 恢复原位
  position.value = { ...position.value, y: 0 }
}

// ── 窗口大小变化 ──

function onResize() {
  screenWidth.value = window.innerWidth
  screenHeight.value = window.innerHeight

  // 确保图标在屏幕范围内
  if (isCollapsed.value) {
    collapsedPosition.value = {
      x: Math.min(collapsedPosition.value.x, screenWidth.value - 56),
      y: Math.min(collapsedPosition.value.y, screenHeight.value - 56)
    }
  }
}

// ── 输入框焦点 ──

function onFocus() {
  // 可选：展开时自动滚动到可见区域
}

function onBlur() {
  // 可选：失去焦点时自动收起
}
</script>

<style scoped>
/* ── 悬浮容器 ── */
.floating-input-float {
  position: fixed;
  z-index: 1000;
  touch-action: none;
  user-select: none;
  -webkit-user-select: none;
}

.floating-input-float.dragging {
  opacity: 0.85;
  transform: scale(1.02);
}

/* ── 收起状态：小圆形图标 ── */
.float-icon {
  width: 56px;
  height: 56px;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.15);
  box-shadow:
    0 4px 16px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  display: flex;
  align-items: center;
  justify-content: center;
  color: var(--text-primary, #fff);
  cursor: pointer;
  animation: breathe 2s ease-in-out infinite;
  transition: transform 0.3s cubic-bezier(0.34, 1.56, 0.64, 1);
}

.float-icon:active {
  transform: scale(0.9);
}

/* 呼吸动画 */
@keyframes breathe {
  0%, 100% { transform: scale(1); opacity: 0.9; }
  50% { transform: scale(1.05); opacity: 1; }
}

/* ── 展开状态：完整输入框 ── */
.float-input-box {
  width: calc(100vw - 32px);
  margin: 0 16px;
  border-radius: 20px;
  background: rgba(255, 255, 255, 0.08);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.12);
  box-shadow:
    0 8px 32px rgba(0, 0, 0, 0.3),
    inset 0 1px 0 rgba(255, 255, 255, 0.1);
  overflow: hidden;
  animation: expand-spring 0.4s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* 展开动画 */
@keyframes expand-spring {
  0% {
    transform: scale(0.3);
    opacity: 0;
    border-radius: 50%;
  }
  50% {
    transform: scale(1.05);
    opacity: 1;
  }
  70% {
    transform: scale(0.98);
  }
  100% {
    transform: scale(1);
    border-radius: 20px;
  }
}

/* ── 浅色主题适配 ── */
[data-theme="light"] .float-icon {
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0, 0, 0, 0.1);
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
}

[data-theme="light"] .float-input-box {
  background: rgba(255, 255, 255, 0.9);
  border: 1px solid rgba(0, 0, 0, 0.1);
  box-shadow: 0 8px 32px rgba(0, 0, 0, 0.15);
}

/* ── 拖动手柄 + 收起按钮 ── */
.float-input-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 12px 4px;
}

.drag-handle {
  color: var(--text-muted, rgba(255,255,255,0.4));
  cursor: grab;
}

.drag-handle:active {
  cursor: grabbing;
}

.collapse-btn {
  width: 28px;
  height: 28px;
  border: none;
  border-radius: 50%;
  background: rgba(255, 255, 255, 0.1);
  color: var(--text-muted, rgba(255,255,255,0.6));
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: background 0.2s;
}

.collapse-btn:hover {
  background: rgba(255, 255, 255, 0.2);
}

/* ── 输入区域 ── */
.float-input-content {
  display: flex;
  align-items: flex-end;
  gap: 8px;
  padding: 4px 12px 12px;
}

.float-input-content textarea {
  flex: 1;
  min-height: 36px;
  max-height: 120px;
  padding: 8px 12px;
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.05);
  color: var(--text-primary, #fff);
  font-size: 15px;
  line-height: 1.4;
  resize: none;
  outline: none;
  font-family: inherit;
}

[data-theme="light"] .float-input-content textarea {
  border: 1px solid rgba(0, 0, 0, 0.1);
  background: rgba(0, 0, 0, 0.03);
}

.float-input-content textarea::placeholder {
  color: var(--text-muted, rgba(255,255,255,0.4));
}

.float-send-btn {
  width: 36px;
  height: 36px;
  border: none;
  border-radius: 50%;
  background: var(--accent-color, #0A84FF);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: opacity 0.2s, transform 0.2s;
  flex-shrink: 0;
}

.float-send-btn:disabled {
  opacity: 0.4;
  cursor: not-allowed;
}

.float-send-btn:not(:disabled):active {
  transform: scale(0.9);
}
</style>
