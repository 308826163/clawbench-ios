<template>
  <Teleport to="body">
    <div
      v-if="everOpened"
      v-show="open || leaving"
      ref="overlayRef"
      class="modal-overlay"
      :class="{ 'modal-leaving': leaving }"
      :style="{ zIndex }"
      tabindex="-1"
      @click.self="handleClose"
      @keydown.escape="handleClose"
    >
      <div class="modal-dialog" :class="{ 'modal-leaving': leaving, 'modal-full-height': fullHeight }" @click.stop>
        <div class="modal-header">
          <slot name="header">
            <span class="modal-title">{{ title }}</span>
          </slot>
          <button class="modal-close-btn" @click="handleClose" :title="'Close'">
            <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="3" y1="3" x2="11" y2="11"/><line x1="11" y1="3" x2="3" y2="11"/></svg>
          </button>
        </div>
        <div class="modal-body">
          <slot />
        </div>
        <div class="modal-footer" :class="{ 'modal-footer-default': !$slots.footer }">
          <slot name="footer" />
        </div>
        <slot name="after" />
      </div>
    </div>
  </Teleport>
</template>

<script setup>
import { ref, watch, nextTick } from 'vue'

const props = defineProps({
  open: Boolean,
  title: { type: String, default: '' },
  zIndex: { type: Number, default: 2100 },
  fullHeight: { type: Boolean, default: false },
})

const emit = defineEmits(['close'])

const leaving = ref(false)
const everOpened = ref(false)
const overlayRef = ref(null)
let leaveTimer = null

watch(() => props.open, (val) => {
  clearTimeout(leaveTimer)
  if (val) {
    everOpened.value = true
    leaving.value = false
    // Auto-focus overlay so Escape key works immediately
    nextTick(() => {
      overlayRef.value?.focus()
    })
  } else if (leaving.value) {
    // Close triggered externally while animating — cancel animation, hide now
    leaving.value = false
  }
}, { immediate: true })

function handleClose() {
  if (leaving.value) return
  leaving.value = true
  leaveTimer = setTimeout(() => {
    leaving.value = false
    leaveTimer = null
    emit('close')
  }, 250)
}

defineExpose({
  close: handleClose,
})
</script>

<style>
/* ═══════════════════════════════════════════════════════════════════════
 * ModalDialog - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.modal-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  z-index: 2100;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 44px 16px 48px;
  animation: modal-fadeIn 0.2s ease;
}

/* Light Theme Overlay */
[data-theme="light"] .modal-overlay {
  background: rgba(0, 0, 0, 0.3);
}

.modal-overlay.modal-leaving {
  animation: modal-fadeOut 0.25s ease forwards;
}

/* ── Dialog (Liquid Glass - Dark Theme) ── */
.modal-dialog {
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
  border-radius: 28px;
  width: 100%;
  max-height: 100%;
  height: auto;
  display: flex;
  flex-direction: column;
  overflow: hidden;

  /* Floating elevation */
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.3),
    0 8px 24px rgba(0, 0, 0, 0.2),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.15),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);

  animation: modal-scaleIn 0.25s cubic-bezier(0.34, 1.56, 0.64, 1);
}

/* Light Theme Dialog */
[data-theme="light"] .modal-dialog {
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

.modal-dialog.modal-full-height {
  height: 100%;
}

.modal-dialog.modal-leaving {
  animation: modal-scaleOut 0.25s ease forwards;
}

@keyframes modal-fadeIn {
  from { opacity: 0; }
  to   { opacity: 1; }
}

@keyframes modal-fadeOut {
  from { opacity: 1; }
  to   { opacity: 0; }
}

@keyframes modal-scaleIn {
  from {
    opacity: 0;
    transform: translateY(24px) scale(0.94);
  }
  to {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
}

@keyframes modal-scaleOut {
  from {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  to {
    opacity: 0;
    transform: translateY(24px) scale(0.94);
  }
}

/* ── Header (Liquid Glass - Dark Theme) ── */
.modal-header {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-bottom: 0.5px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.03);
  flex-shrink: 0;
}

/* Light Theme Header */
[data-theme="light"] .modal-header {
  border-bottom: 0.5px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.02);
}

.modal-close-btn {
  margin-left: auto;
  width: 28px;
  height: 28px;
  border-radius: 50%;
  border: none;
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.6);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  flex-shrink: 0;
  padding: 0;
  transition: all 0.2s ease;
}

[data-theme="light"] .modal-close-btn {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.5);
}

.modal-close-btn:hover {
  background: rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .modal-close-btn:hover {
  background: rgba(0, 0, 0, 0.1);
  color: rgba(0, 0, 0, 0.8);
}

.modal-close-btn:active {
  transform: scale(0.92);
}

.modal-header-icon {
  flex-shrink: 0;
  color: rgba(255, 255, 255, 0.8);
  display: flex;
  align-items: center;
}

[data-theme="light"] .modal-header-icon {
  color: rgba(0, 0, 0, 0.7);
}

.modal-title {
  font-weight: 600;
  font-size: 15px;
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .modal-title {
  color: rgba(0, 0, 0, 0.85);
}

/* ── Body ── */
.modal-body {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  overflow-x: hidden;
  display: flex;
  flex-direction: column;
  color: rgba(255, 255, 255, 0.8);
}

[data-theme="light"] .modal-body {
  color: rgba(0, 0, 0, 0.75);
}

/* ── Footer (Liquid Glass) ── */
.modal-footer {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 12px 16px;
  border-top: 0.5px solid rgba(255, 255, 255, 0.08);
  background: rgba(255, 255, 255, 0.02);
  flex-shrink: 0;
  justify-content: flex-end;
  border-radius: 0 0 28px 28px;
}

[data-theme="light"] .modal-footer {
  border-top: 0.5px solid rgba(0, 0, 0, 0.06);
  background: rgba(0, 0, 0, 0.01);
}

.modal-footer-default {
  min-height: 8px;
  padding: 12px 16px;
  justify-content: center;
}
</style>
