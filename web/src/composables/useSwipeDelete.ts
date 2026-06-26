import { ref, onMounted, onUnmounted } from 'vue'

export const DELETE_WIDTH = 70 // px – width of the revealed delete button area

// 全局事件总线：关闭其他 useSwipeDelete 实例
const REVEAL_EVENT = 'swipe-delete-reveal'

export function useSwipeDelete() {
  const offset = ref(0) // current translateX offset (0 = closed, -DELETE_WIDTH = fully revealed)
  const isRevealed = ref(false)

  let startX = 0
  let startY = 0
  let startTime = 0
  let dragging = false
  let directionLocked: 'h' | 'v' | null = null
  let revealTime = 0

  const SWIPE_THRESHOLD = 20 // px to lock direction
  const MAX_DURATION = 400 // ms

  // 关闭其他已展开的实例
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
    startTime = Date.now()
    dragging = true
    directionLocked = null
  }

  function onTouchMove(e: TouchEvent) {
    if (!dragging) return
    const touch = e.touches[0]
    const dx = touch.clientX - startX
    const dy = touch.clientY - startY

    // Lock direction on first meaningful movement
    if (!directionLocked) {
      if (Math.abs(dx) < SWIPE_THRESHOLD && Math.abs(dy) < SWIPE_THRESHOLD) return
      directionLocked = Math.abs(dx) >= Math.abs(dy) ? 'h' : 'v'
    }

    if (directionLocked === 'v') {
      dragging = false
      return
    }

    // Prevent vertical scroll while swiping horizontally
    e.preventDefault()

    // Calculate offset: left swipe = negative offset, right swipe = reduce offset
    const baseOffset = isRevealed.value ? -DELETE_WIDTH : 0
    let newOffset = baseOffset + dx

    // Clamp: can't go right of 0, can't go left past -DELETE_WIDTH
    if (newOffset > 0) newOffset = 0
    if (newOffset < -DELETE_WIDTH) newOffset = -DELETE_WIDTH

    offset.value = newOffset
  }

  function onTouchEnd() {
    if (!dragging) return
    dragging = false

    const duration = Date.now() - startTime

    // Determine target: snap open or closed
    const currentOff = offset.value

    if (isRevealed.value) {
      // Currently revealed: close if swiped right past midpoint, or fast swipe right
      if (currentOff > -DELETE_WIDTH / 2 || (duration < MAX_DURATION && currentOff > -DELETE_WIDTH * 0.3)) {
        close()
      } else {
        snapOpen()
      }
    } else {
      // Currently closed: open if swiped left past midpoint, or fast swipe left
      if (currentOff < -DELETE_WIDTH / 2 || (duration < MAX_DURATION && currentOff < -DELETE_WIDTH * 0.3)) {
        snapOpen()
      } else {
        close()
      }
    }
  }

  function snapOpen() {
    offset.value = -DELETE_WIDTH
    isRevealed.value = true
    // 关闭其他已展开的实例
    closeOthers()
  }

  function close() {
    offset.value = 0
    isRevealed.value = false
  }

  /** Call when content area is clicked while revealed — close the swipe */
  function onContentClick() {
    if (isRevealed.value) {
      close()
      return true // signal that click was consumed
    }
    return false
  }

  return {
    offset,
    isRevealed,
    onTouchStart,
    onTouchMove,
    onTouchEnd,
    onContentClick,
    close,
  }
}
