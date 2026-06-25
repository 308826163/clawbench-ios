<template>
  <Teleport to="body">
    <Transition name="dialog">
      <div v-if="visible" class="delete-dialog-overlay" @click.self="onCancel">
        <div class="delete-dialog">
          <div class="delete-dialog__title">删除此项目？</div>
          <div class="delete-dialog__path">
            📁 {{ path }}
          </div>
          <div class="delete-dialog__warning">
            所有聊天记录将被永久删除
          </div>
          <div class="delete-dialog__actions">
            <button class="delete-dialog__btn delete-dialog__btn--cancel" @click="onCancel">
              取消
            </button>
            <button
              class="delete-dialog__btn delete-dialog__btn--confirm"
              :class="{
                'delete-dialog__btn--confirming': confirming,
                'delete-dialog__btn--animating': animating
              }"
              @click="onConfirm"
            >
              <span class="delete-dialog__btn-text">{{ confirming ? '确认' : '删除' }}</span>
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'

const props = defineProps<{
  visible: boolean
  path: string
}>()

const emit = defineEmits<{
  confirm: []
  cancel: []
}>()

const confirming = ref(false)
let confirmTimer: ReturnType<typeof setTimeout> | null = null

// Reset state when dialog opens
watch(() => props.visible, (val) => {
  if (val) {
    confirming.value = false
    if (confirmTimer) {
      clearTimeout(confirmTimer)
      confirmTimer = null
    }
  }
})

function onConfirm() {
  if (!confirming.value) {
    // First click: enter confirming state
    confirming.value = true

    // Set timeout to reset if not confirmed
    confirmTimer = setTimeout(() => {
      confirming.value = false
      confirmTimer = null
    }, 1000)
  } else {
    // Second click: execute delete
    if (confirmTimer) {
      clearTimeout(confirmTimer)
      confirmTimer = null
    }
    emit('confirm')
  }
}

function onCancel() {
  if (confirmTimer) {
    clearTimeout(confirmTimer)
    confirmTimer = null
  }
  confirming.value = false
  emit('cancel')
}
</script>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════
 * DeleteConfirmDialog - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.delete-dialog-overlay {
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
[data-theme="light"] .delete-dialog-overlay {
  background: rgba(0, 0, 0, 0.3);
}

/* ── Dialog (Liquid Glass - Dark Theme) ── */
.delete-dialog {
  width: 80%;
  max-width: 260px;

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
  padding: 20px;
  text-align: center;

  /* Floating elevation */
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.3),
    0 8px 24px rgba(0, 0, 0, 0.2),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.15),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);
}

/* Light Theme Dialog */
[data-theme="light"] .delete-dialog {
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

.delete-dialog__icon {
  font-size: 32px;
  margin-bottom: 8px;
}

.delete-dialog__title {
  font-size: 15px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 10px;
}

[data-theme="light"] .delete-dialog__title {
  color: rgba(0, 0, 0, 0.85);
}

.delete-dialog__path {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.6);
  background: rgba(255, 255, 255, 0.06);
  padding: 8px 10px;
  border-radius: 10px;
  margin-bottom: 10px;
  word-break: break-all;
  text-align: left;
}

[data-theme="light"] .delete-dialog__path {
  color: rgba(0, 0, 0, 0.55);
  background: rgba(0, 0, 0, 0.04);
}

.delete-dialog__warning {
  font-size: 12px;
  color: rgba(239, 68, 68, 0.9);
  margin-bottom: 16px;
}

.delete-dialog__actions {
  display: flex;
  gap: 10px;
}

.delete-dialog__btn {
  flex: 1;
  padding: 10px 16px;
  border: none;
  border-radius: 12px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  position: relative;
  overflow: visible;
}

.delete-dialog__btn:active {
  transform: scale(0.95);
}

.delete-dialog__btn--cancel {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}

[data-theme="light"] .delete-dialog__btn--cancel {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.7);
}

.delete-dialog__btn--cancel:hover {
  background: rgba(255, 255, 255, 0.15);
}

[data-theme="light"] .delete-dialog__btn--cancel:hover {
  background: rgba(0, 0, 0, 0.1);
}

.delete-dialog__btn--confirm {
  background: rgba(239, 68, 68, 0.9);
  color: #fff;
  overflow: visible;
}

.delete-dialog__btn--confirm:hover {
  background: rgba(239, 68, 68, 1);
}

.delete-dialog__btn--confirming {
  background: rgba(220, 38, 38, 1);
}

.delete-dialog__btn-text {
  position: relative;
  z-index: 1;
}

/* Transition */
.dialog-enter-active {
  transition: opacity 0.2s ease;
}

.dialog-leave-active {
  transition: opacity 0.15s ease;
}

.dialog-enter-from,
.dialog-leave-to {
  opacity: 0;
}

.dialog-enter-active .delete-dialog {
  animation: dialog-in 0.25s ease-out;
}

.dialog-leave-active .delete-dialog {
  animation: dialog-out 0.15s ease-in;
}

@keyframes dialog-in {
  from {
    transform: scale(0.9);
    opacity: 0;
  }
  to {
    transform: scale(1);
    opacity: 1;
  }
}

@keyframes dialog-out {
  from {
    transform: scale(1);
    opacity: 1;
  }
  to {
    transform: scale(0.9);
    opacity: 0;
  }
}
</style>
