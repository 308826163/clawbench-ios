<template>
  <div class="settings-restart-overlay" @click.self="$emit('later')">
    <div class="settings-restart-dialog">
      <div class="settings-restart-dialog__header">{{ t('settings.restartConfirmTitle') }}</div>
      <p class="settings-restart-dialog__message">{{ t('settings.restartConfirmMessage') }}</p>
      <ul v-if="changedFields.length > 0" class="settings-restart-dialog__list">
        <li v-for="field in displayFields" :key="field">{{ field }}</li>
      </ul>
      <div class="settings-restart-dialog__actions">
        <button class="settings-restart-dialog__btn settings-restart-dialog__btn--later" @click="$emit('later')">
          {{ t('settings.restartLater') }}
        </button>
        <button class="settings-restart-dialog__btn settings-restart-dialog__btn--restart" @click="$emit('restart')">
          {{ t('settings.restartNow') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { serverFieldToLabelKey } from './settingsFieldMap'

const props = defineProps<{
  changedFields: string[]
}>()

defineEmits<{
  restart: []
  later: []
}>()

const { t } = useI18n()

const displayFields = computed(() =>
  props.changedFields.map(key => {
    const labelKey = serverFieldToLabelKey[key]
    return labelKey ? t(labelKey) : key
  })
)
</script>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════
 * SettingsRestartDialog - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.settings-restart-overlay {
  position: absolute;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 10;
}

/* Light Theme Overlay */
[data-theme="light"] .settings-restart-overlay {
  background: rgba(0, 0, 0, 0.3);
}

/* ── Dialog (Liquid Glass - Dark Theme) ── */
.settings-restart-dialog {
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
  padding: 24px;
  margin: 24px;
  max-width: 320px;
  width: 100%;

  /* Floating elevation */
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.3),
    0 8px 24px rgba(0, 0, 0, 0.2),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.15),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);
}

/* Light Theme Dialog */
[data-theme="light"] .settings-restart-dialog {
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

.settings-restart-dialog__header {
  font-size: 17px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 8px;
  text-align: center;
}

[data-theme="light"] .settings-restart-dialog__header {
  color: rgba(0, 0, 0, 0.85);
}

.settings-restart-dialog__message {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  margin: 0 0 12px;
  text-align: center;
}

[data-theme="light"] .settings-restart-dialog__message {
  color: rgba(0, 0, 0, 0.55);
}

.settings-restart-dialog__list {
  margin: 0 0 20px;
  padding-left: 20px;
  font-size: 14px;
  color: rgba(255, 255, 255, 0.6);
  line-height: 1.6;
}

[data-theme="light"] .settings-restart-dialog__list {
  color: rgba(0, 0, 0, 0.55);
}

.settings-restart-dialog__list li {
  margin-bottom: 2px;
}

.settings-restart-dialog__actions {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.settings-restart-dialog__btn {
  width: 100%;
  padding: 12px 16px;
  border: none;
  border-radius: 12px;
  font-size: 16px;
  font-weight: 500;
  cursor: pointer;
  text-align: center;
  transition: all 0.2s ease;
}

.settings-restart-dialog__btn--later {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}

[data-theme="light"] .settings-restart-dialog__btn--later {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.7);
}

@media (hover: hover) {
  .settings-restart-dialog__btn--later:hover {
    background: rgba(255, 255, 255, 0.15);
  }
}

.settings-restart-dialog__btn--later:active {
  background: rgba(255, 255, 255, 0.12);
}

.settings-restart-dialog__btn--restart {
  background: rgba(59, 130, 246, 0.9);
  color: #fff;
}

@media (hover: hover) {
  .settings-restart-dialog__btn--restart:hover {
    background: rgba(59, 130, 246, 1);
  }
}

.settings-restart-dialog__btn--restart:active {
  background: rgba(37, 99, 235, 1);
}
</style>
