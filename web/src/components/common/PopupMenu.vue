<template>
  <Teleport to="body">
    <Transition name="menu-fade">
      <div v-if="show" class="popup-menu" role="menu" :style="menuStyle" @click.stop="emit('update:show', false)" @keydown.escape="emit('update:show', false)">
        <slot />
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
import { ref, watch, onBeforeUnmount } from 'vue'
import { computeMenuStyle } from '@/utils/popupMenuPosition'

const props = defineProps({
  show: Boolean,
  targetElement: { type: Object }, // DOM element reference
  maxWidth: { type: Number, default: 220 },
  maxHeight: { type: Number, default: 320 },
  edgeMargin: { type: Number, default: 6 },
  menuItemsCount: { type: Number, default: 10 }, // for height estimation
  anchor: { type: String, default: 'auto', validator: (v) => ['left', 'right', 'auto'].includes(v) }, // force horizontal alignment
})

const emit = defineEmits(['update:show'])

// Reactive style — updated manually so we can react to DOM geometry changes
// (scroll, resize) that Vue's computed cannot track.
const menuStyle = ref({})

/** Recalculate position from current anchor geometry. */
function updatePosition() {
  if (!props.targetElement) { menuStyle.value = {}; return }
  const rect = props.targetElement.getBoundingClientRect()
  menuStyle.value = computeMenuStyle(rect, {
    maxWidth: props.maxWidth,
    maxHeight: props.maxHeight,
    edgeMargin: props.edgeMargin,
    menuItemsCount: props.menuItemsCount,
    anchor: props.anchor,
  })
}

// Close on outside click
function handleClickOutside(e) {
  if (!props.targetElement) return
  if (props.targetElement.contains(e.target)) return
  if (e.target.closest('.popup-menu')) return
  emit('update:show', false)
}

// Recalculate on scroll/resize while open
function onLayoutChange() {
  if (props.show) updatePosition()
}

watch(() => props.show, (val) => {
  if (val) {
    // Compute position synchronously — the target element already exists in DOM
    // and we need the style before the first paint of the menu.
    updatePosition()
    // Listen for layout changes that could move the anchor
    window.addEventListener('scroll', onLayoutChange, true) // capture to catch all scrolls
    window.addEventListener('resize', onLayoutChange)
    // Use setTimeout to avoid the opening click being treated as outside click
    setTimeout(() => {
      if (props.show) {
        document.addEventListener('click', handleClickOutside)
      }
    }, 0)
  } else {
    window.removeEventListener('scroll', onLayoutChange, true)
    window.removeEventListener('resize', onLayoutChange)
    document.removeEventListener('click', handleClickOutside)
  }
})

// Cleanup on unmount
onBeforeUnmount(() => {
  window.removeEventListener('scroll', onLayoutChange, true)
  window.removeEventListener('resize', onLayoutChange)
  document.removeEventListener('click', handleClickOutside)
})
</script>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════
 * PopupMenu - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Menu (Liquid Glass - Dark Theme) ── */
.popup-menu {
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
  padding: 6px;
}

/* Light Theme Menu */
:host([data-theme="light"]) .popup-menu,
[data-theme="light"] .popup-menu {
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

/* Fade animation for menu appearance */
.menu-fade-enter-active,
.menu-fade-leave-active {
  transition: opacity 0.15s ease, transform 0.15s ease;
}

.menu-fade-enter-from,
.menu-fade-leave-to {
  opacity: 0;
  transform: translateY(-4px) scale(0.96);
}
</style>
