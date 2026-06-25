<template>
  <div class="setup-step setup-welcome">
    <div class="welcome-icon">
      <span class="welcome-emoji">🥧</span>
    </div>
    <h2 class="welcome-title">{{ t('setup.welcomeTitle') }}</h2>
    <p class="welcome-desc">{{ t('setup.welcomeDesc') }}</p>
    <p v-if="agentVersion" class="welcome-version">Pi v{{ agentVersion }}</p>
    <button class="setup-btn-primary" @click="$emit('next')" :disabled="!embeddedAgent">
      <span>{{ t('setup.configureAgent') }}</span>
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="16" height="16">
        <path d="M5 12h14M12 5l7 7-7 7"/>
      </svg>
    </button>
    <button class="setup-btn-secondary view-backends-btn" @click="showBackends = true">
      <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
        <circle cx="12" cy="12" r="10"/>
        <path d="M12 16v-4M12 8h.01"/>
      </svg>
      <span>{{ t('setup.viewSupportedBackends') }}</span>
    </button>
    <p v-if="embeddedAgent === false" class="welcome-hint">{{ t('setup.noEmbeddedAgent') }}</p>

    <!-- Backends info overlay -->
    <Transition name="backends-fade">
      <div v-if="showBackends" class="backends-overlay" @click.self="showBackends = false">
        <div class="backends-panel">
          <div class="backends-header">
            <h3>{{ t('setup.supportedBackends') }}</h3>
            <button class="backends-close" @click="showBackends = false">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="18" height="18">
                <path d="M18 6L6 18M6 6l12 12"/>
              </svg>
            </button>
          </div>
          <p class="backends-desc">{{ t('setup.supportedBackendsDesc') }}</p>
          <div class="backends-list">
            <div v-for="b in backends" :key="b.id" class="backend-item">
              <div class="backend-icon">{{ b.icon }}</div>
              <div class="backend-info">
                <div class="backend-name">{{ b.name }}</div>
                <div class="backend-detail">
                  <span class="backend-tag">{{ t('setup.backendCmd') }}: <code>{{ b.default_cmd }}</code></span>
                  <span class="backend-tag">{{ b.specialty }}</span>
                </div>
                <div v-if="b.thinking_effort_levels?.length" class="backend-detail">
                  <span class="backend-tag">{{ t('setup.backendThinking') }}: {{ b.thinking_effort_levels.join(', ') }}</span>
                </div>
              </div>
            </div>
          </div>
          <div class="backends-footer">
            <p class="backends-restart-hint">
              <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" width="14" height="14">
                <path d="M23 4v6h-6M1 20v-6h6"/>
                <path d="M3.51 9a9 9 0 0114.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0020.49 15"/>
              </svg>
              {{ t('setup.restartToDetect') }}
            </p>
            <button class="setup-btn-primary backends-close-btn" @click="showBackends = false">
              {{ t('setup.close') }}
            </button>
          </div>
        </div>
      </div>
    </Transition>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useSetup, type BackendInfo } from '@/composables/useSetup'

defineProps<{
  embeddedAgent: boolean | null
  agentVersion: string
}>()

defineEmits<{
  next: []
}>()

const { t } = useI18n()
const { getBackends } = useSetup()

const showBackends = ref(false)
const backends = ref<BackendInfo[]>([])

onMounted(async () => {
  try {
    backends.value = await getBackends()
  } catch { /* will show empty list */ }
})
</script>

<style scoped>
.setup-welcome {
  display: flex;
  flex-direction: column;
  align-items: center;
  text-align: center;
  gap: 10px;
  padding: 12px 0;
}

.welcome-icon {
  position: relative;
  width: 64px;
  height: 64px;
}

.welcome-emoji {
  font-size: 48px;
  line-height: 64px;
  display: block;
}

.welcome-title {
  font-size: 18px;
  font-weight: 700;
  color: var(--text-primary);
  margin: 0;
}

.welcome-desc {
  font-size: 13px;
  color: var(--text-secondary);
  margin: 0;
  max-width: 300px;
  line-height: 1.5;
}

.welcome-version {
  font-size: 11px;
  color: var(--text-muted);
  margin: 0;
  padding: 2px 10px;
  background: var(--bg-tertiary);
  border-radius: 12px;
}

.welcome-hint {
  font-size: 12px;
  color: var(--color-red, #dc2626);
  margin: 4px 0 0;
}

.view-backends-btn {
  font-size: 12px;
  padding: 6px 12px;
  gap: 4px;
}

/* ═══════════════════════════════════════════════════════════════════════
 * SetupWelcome Backends Overlay - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.backends-overlay {
  position: fixed;
  inset: 0;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  padding: 16px;
}

/* Light Theme Overlay */
[data-theme="light"] .backends-overlay {
  background: rgba(0, 0, 0, 0.3);
}

/* ── Panel (Liquid Glass - Dark Theme) ── */
.backends-panel {
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
  width: 100%;
  max-width: 420px;
  max-height: 80vh;
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

/* Light Theme Panel */
[data-theme="light"] .backends-panel {
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

.backends-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 16px 20px 12px;
}

.backends-header h3 {
  margin: 0;
  font-size: 16px;
  font-weight: 700;
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .backends-header h3 {
  color: rgba(0, 0, 0, 0.85);
}

.backends-close {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 32px;
  height: 32px;
  border: none;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.6);
  cursor: pointer;
  transition: all 0.2s ease;
}

[data-theme="light"] .backends-close {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.5);
}

.backends-close:hover {
  background: rgba(255, 255, 255, 0.15);
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .backends-close:hover {
  background: rgba(0, 0, 0, 0.1);
  color: rgba(0, 0, 0, 0.8);
}

.backends-desc {
  margin: 0 20px 12px;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.5;
}

[data-theme="light"] .backends-desc {
  color: rgba(0, 0, 0, 0.55);
}

.backends-list {
  flex: 1;
  overflow-y: auto;
  padding: 0 16px;
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.backend-item {
  display: flex;
  gap: 12px;
  padding: 12px;
  border-radius: 14px;
  background: rgba(255, 255, 255, 0.05);
  border: 0.5px solid rgba(255, 255, 255, 0.06);
  text-align: left;
}

[data-theme="light"] .backend-item {
  background: rgba(0, 0, 0, 0.03);
  border: 0.5px solid rgba(0, 0, 0, 0.05);
}

.backend-icon {
  font-size: 24px;
  line-height: 1;
  flex-shrink: 0;
  width: 32px;
  text-align: center;
  padding-top: 2px;
}

.backend-info {
  flex: 1;
  min-width: 0;
}

.backend-name {
  font-size: 14px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
}

[data-theme="light"] .backend-name {
  color: rgba(0, 0, 0, 0.85);
}

.backend-detail {
  margin-top: 4px;
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.backend-tag {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
}

[data-theme="light"] .backend-tag {
  color: rgba(0, 0, 0, 0.45);
}

.backend-tag code {
  font-family: var(--font-mono, monospace);
  font-size: 12px;
  background: rgba(255, 255, 255, 0.08);
  padding: 2px 6px;
  border-radius: 6px;
  color: rgba(100, 180, 255, 0.9);
}

[data-theme="light"] .backend-tag code {
  background: rgba(0, 0, 0, 0.05);
  color: rgba(37, 99, 235, 0.9);
}

.backends-footer {
  padding: 12px 20px 16px;
  display: flex;
  flex-direction: column;
  gap: 10px;
  align-items: center;
  border-top: 0.5px solid rgba(255, 255, 255, 0.06);
}

[data-theme="light"] .backends-footer {
  border-top: 0.5px solid rgba(0, 0, 0, 0.05);
}

.backends-restart-hint {
  margin: 0;
  font-size: 12px;
  color: rgba(255, 255, 255, 0.5);
  display: flex;
  align-items: center;
  gap: 6px;
}

[data-theme="light"] .backends-restart-hint {
  color: rgba(0, 0, 0, 0.45);
}

.backends-close-btn {
  width: 100%;
}

/* ── Transition ── */
.backends-fade-enter-active {
  transition: opacity 0.2s ease;
}
.backends-fade-leave-active {
  transition: opacity 0.15s ease;
}
.backends-fade-enter-from,
.backends-fade-leave-to {
  opacity: 0;
}
</style>
