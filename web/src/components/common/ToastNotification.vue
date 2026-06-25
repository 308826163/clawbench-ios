<template>
  <Teleport to="body">
    <Transition name="toast">
      <div v-if="toast.visible.value" :class="['toast', `toast-${toast.type.value}`]" @click="toast.onClick.value ? (toast.onClick.value(), toast.dismiss()) : toast.dismiss()">
        <span v-if="toast.icon.value" class="toast-icon">{{ toast.icon.value }}</span>
        <span class="toast-text">{{ toast.message.value }}</span>
      </div>
    </Transition>
  </Teleport>
</template>

<script setup>
defineProps({
    toast: {
        type: Object,
        required: true,
    },
})
</script>

<style>
/* ═══════════════════════════════════════════════════════════════════════
 * ToastNotification - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Toast (Liquid Glass - Dark Theme) ── */
.toast {
    position: fixed;
    top: calc(8px + var(--header-safe-area-top, 0px));
    left: 0;
    right: 0;
    margin: 0 auto;

    /* Liquid Glass Material */
    background: linear-gradient(
        135deg,
        rgba(30, 30, 30, 0.85) 0%,
        rgba(20, 20, 20, 0.9) 50%,
        rgba(25, 25, 25, 0.88) 100%
    );
    -webkit-backdrop-filter: blur(16px) saturate(180%);
    backdrop-filter: blur(16px) saturate(180%);

    color: rgba(255, 255, 255, 0.9);
    border-radius: 18px;
    padding: 10px 16px;
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 14px;
    font-weight: 500;

    /* Floating elevation */
    box-shadow:
        0 8px 24px rgba(0, 0, 0, 0.2),
        0 2px 8px rgba(0, 0, 0, 0.15),
        inset 0 0.5px 0 rgba(255, 255, 255, 0.12);

    border: 0.5px solid rgba(255, 255, 255, 0.1);
    cursor: pointer;
    z-index: 9999;
    white-space: normal;
    width: fit-content;
    min-width: 80px;
    max-width: 88vw;
    text-align: left;
    line-height: 1.4;
    -webkit-tap-highlight-color: transparent;
    user-select: none;
    transition: all 0.2s ease;
}

/* Light Theme Toast */
[data-theme="light"] .toast {
    background: linear-gradient(
        135deg,
        rgba(255, 255, 255, 0.9) 0%,
        rgba(245, 245, 245, 0.92) 50%,
        rgba(250, 250, 250, 0.91) 100%
    );
    color: rgba(0, 0, 0, 0.85);
    border: 0.5px solid rgba(255, 255, 255, 0.5);
    box-shadow:
        0 8px 24px rgba(0, 0, 0, 0.1),
        0 2px 8px rgba(0, 0, 0, 0.06),
        inset 0 0.5px 0 rgba(255, 255, 255, 0.8);
}

/* Error Toast */
.toast-error {
    background: linear-gradient(
        135deg,
        rgba(239, 68, 68, 0.85) 0%,
        rgba(220, 38, 38, 0.9) 50%,
        rgba(239, 68, 68, 0.88) 100%
    );
}

[data-theme="light"] .toast-error {
    background: linear-gradient(
        135deg,
        rgba(239, 68, 68, 0.9) 0%,
        rgba(220, 38, 38, 0.92) 50%,
        rgba(239, 68, 68, 0.91) 100%
    );
}

/* Success Toast */
.toast-success {
    background: linear-gradient(
        135deg,
        rgba(34, 197, 94, 0.85) 0%,
        rgba(22, 163, 74, 0.9) 50%,
        rgba(34, 197, 94, 0.88) 100%
    );
}

[data-theme="light"] .toast-success {
    background: linear-gradient(
        135deg,
        rgba(34, 197, 94, 0.9) 0%,
        rgba(22, 163, 74, 0.92) 50%,
        rgba(34, 197, 94, 0.91) 100%
    );
}

/* Info Toast */
.toast-info {
    background: linear-gradient(
        135deg,
        rgba(59, 130, 246, 0.85) 0%,
        rgba(37, 99, 235, 0.9) 50%,
        rgba(59, 130, 246, 0.88) 100%
    );
}

[data-theme="light"] .toast-info {
    background: linear-gradient(
        135deg,
        rgba(59, 130, 246, 0.9) 0%,
        rgba(37, 99, 235, 0.92) 50%,
        rgba(59, 130, 246, 0.91) 100%
    );
}

.toast:active {
    opacity: 0.8;
    transform: scale(0.97);
}

.toast-icon {
    font-size: 16px;
}

.toast-text {
    flex: 1;
    min-width: 0;
    overflow-wrap: break-word;
}

.toast-enter-active,
.toast-leave-active {
    transition: opacity 0.25s ease, transform 0.25s ease;
}

.toast-enter-from,
.toast-leave-to {
    opacity: 0;
    transform: translateY(-8px) scale(0.96);
}
</style>
