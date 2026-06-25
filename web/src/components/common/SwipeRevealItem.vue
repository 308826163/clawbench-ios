<template>
  <div
    class="swipe-reveal-item"
    @touchstart="onTouchStart"
    @touchmove="onTouchMove"
    @touchend="onTouchEnd"
  >
    <div
      class="swipe-reveal-item__content"
      :style="{ transform: `translateX(${offset}px)` }"
      :class="{ 'swipe-reveal-item__content--swiping': swiping }"
    >
      <slot />
    </div>
    <div
      class="swipe-reveal-item__actions"
      :style="{ opacity: Math.abs(offset) / threshold }"
    >
      <slot name="actions" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'

const props = defineProps<{
  threshold?: number
}>()

const emit = defineEmits<{
  reveal: []
  hide: []
}>()

const threshold = props.threshold || 80
const offset = ref(0)
const swiping = ref(false)
const isRevealed = ref(false)
let startX = 0
let startY = 0
let swipingHorizontally = false
let baseOffset = 0

// 全局事件总线：关闭其他 SwipeRevealItem
const REVEAL_EVENT = 'swipe-reveal-item-reveal'
let revealTime = 0

function closeOthers() {
  revealTime = Date.now()
  window.dispatchEvent(new CustomEvent(REVEAL_EVENT))
}

function onCloseOthers() {
  // 忽略自己触发的事件（100ms 内）
  if (isRevealed.value && Date.now() - revealTime > 100) {
    close()
  }
}

onMounted(() => {
  window.addEventListener(REVEAL_EVENT, onCloseOthers)
})

onUnmounted(() => {
  window.removeEventListener(REVEAL_EVENT, onCloseOthers)
})

function onTouchStart(e: TouchEvent) {
  const touch = e.touches[0]
  startX = touch.clientX
  startY = touch.clientY
  baseOffset = offset.value
  swiping.value = true
  swipingHorizontally = false
}

function onTouchMove(e: TouchEvent) {
  if (!swiping.value) return

  const touch = e.touches[0]
  const deltaX = touch.clientX - startX
  const deltaY = touch.clientY - startY

  // Determine swipe direction on first significant move
  if (!swipingHorizontally && Math.abs(deltaX) < 10 && Math.abs(deltaY) < 10) {
    return
  }

  if (!swipingHorizontally) {
    if (Math.abs(deltaY) > Math.abs(deltaX)) {
      // Vertical swipe — don't interfere
      swiping.value = false
      return
    }
    swipingHorizontally = true
  }

  // Calculate new offset based on base offset
  let newOffset = baseOffset + deltaX

  // Clamp: can't go right of 0, can't go left past -threshold * 1.5
  if (newOffset > 0) newOffset = 0
  if (newOffset < -threshold * 1.5) newOffset = -threshold * 1.5

  offset.value = newOffset
  e.preventDefault()
}

function onTouchEnd() {
  swiping.value = false

  if (Math.abs(offset.value) > threshold * 0.6) {
    // Snap to revealed position
    offset.value = -threshold
    isRevealed.value = true
    // 关闭其他 SwipeRevealItem
    closeOthers()
    emit('reveal')
  } else {
    // Snap back to closed
    offset.value = 0
    isRevealed.value = false
    emit('hide')
  }
}

/** Reset to closed position (call after delete) */
function close() {
  offset.value = 0
  isRevealed.value = false
  emit('hide')
}

defineExpose({ close })
</script>

<style scoped>
.swipe-reveal-item {
  position: relative;
  overflow: hidden;
}

.swipe-reveal-item__content {
  position: relative;
  z-index: 1;
  background: var(--bg-primary);
  transition: transform 0.2s ease-out;
}

.swipe-reveal-item__content--swiping {
  transition: none;
}

.swipe-reveal-item__actions {
  position: absolute;
  top: 0;
  right: 0;
  bottom: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  width: 80px;
  background: #ff4444;
  color: #fff;
  font-size: 14px;
  font-weight: 500;
  transition: opacity 0.2s ease-out;
}
</style>
