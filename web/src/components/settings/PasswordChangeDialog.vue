<template>
  <div class="password-dialog-overlay" @click.self="handleClose">
    <div class="password-dialog">
      <div class="password-dialog__header">{{ t('settings.changePasswordTitle') }}</div>

      <div class="password-dialog__field">
        <label class="password-dialog__label">{{ t('settings.currentPassword') }}</label>
        <div class="password-dialog__input-row">
          <input
            type="password"
            class="password-dialog__input"
            v-model="currentPassword"
            :placeholder="t('settings.currentPasswordPlaceholder')"
            @keydown.enter="focusNew"
            autocomplete="current-password"
          />
        </div>
      </div>

      <div class="password-dialog__field">
        <label class="password-dialog__label">{{ t('settings.newPassword') }}</label>
        <div class="password-dialog__input-row">
          <input
            ref="newPasswordRef"
            type="password"
            class="password-dialog__input"
            v-model="newPassword"
            :placeholder="t('settings.newPasswordPlaceholder')"
            @keydown.enter="focusConfirm"
            autocomplete="new-password"
          />
        </div>
      </div>

      <div class="password-dialog__field">
        <label class="password-dialog__label">{{ t('settings.confirmPassword') }}</label>
        <div class="password-dialog__input-row">
          <input
            ref="confirmPasswordRef"
            type="password"
            class="password-dialog__input"
            v-model="confirmPassword"
            :placeholder="t('settings.confirmPasswordPlaceholder')"
            @keydown.enter="submit"
            autocomplete="new-password"
          />
        </div>
      </div>

      <div v-if="localError" class="password-dialog__error">{{ localError }}</div>
      <div v-if="serverError" class="password-dialog__error">{{ serverError }}</div>

      <div class="password-dialog__actions">
        <button class="password-dialog__btn password-dialog__btn--cancel" @click="handleClose" :disabled="submitting">
          {{ t('common.cancel') }}
        </button>
        <button
          class="password-dialog__btn password-dialog__btn--submit"
          :disabled="!canSubmit || submitting"
          @click="submit"
        >
          {{ submitting ? t('settings.changingPassword') : t('settings.changePasswordBtn') }}
        </button>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { apiPost } from '@/utils/api'

const emit = defineEmits<{
  close: []
  changed: [needsRestart: boolean]
}>()

const { t } = useI18n()

const currentPassword = ref('')
const newPassword = ref('')
const confirmPassword = ref('')
const submitting = ref(false)
const localError = ref('')
const serverError = ref('')

const newPasswordRef = ref<HTMLInputElement | null>(null)
const confirmPasswordRef = ref<HTMLInputElement | null>(null)

function focusNew() {
  newPasswordRef.value?.focus()
}

function focusConfirm() {
  confirmPasswordRef.value?.focus()
}

const canSubmit = computed(() => {
  return (
    currentPassword.value !== '' &&
    newPassword.value.length >= 6 &&
    confirmPassword.value !== '' &&
    newPassword.value === confirmPassword.value
  )
})

function validate(): string | null {
  if (!currentPassword.value) {
    return t('settings.currentPasswordRequired')
  }
  if (newPassword.value.length < 6) {
    return t('settings.passwordTooShort')
  }
  if (newPassword.value !== confirmPassword.value) {
    return t('settings.passwordMismatch')
  }
  if (newPassword.value === currentPassword.value) {
    return t('settings.passwordSameAsOld')
  }
  return null
}

async function submit() {
  localError.value = ''
  serverError.value = ''

  const validationError = validate()
  if (validationError) {
    localError.value = validationError
    return
  }

  submitting.value = true
  try {
    const result = await apiPost<{ needs_restart?: boolean }>('/api/config/password', {
      current_password: currentPassword.value,
      new_password: newPassword.value,
    })
    emit('changed', result.needs_restart ?? true)
  } catch (err: any) {
    const errorCode = err?.message || ''
    if (errorCode === 'wrong_password') {
      serverError.value = t('settings.wrongCurrentPassword')
    } else if (errorCode === 'password_too_short') {
      serverError.value = t('settings.passwordTooShort')
    } else if (errorCode === 'empty_password') {
      serverError.value = t('settings.currentPasswordRequired')
    } else if (err?.message?.includes('Too Many Requests') || errorCode === 'TooManyLoginAttempts') {
      serverError.value = t('settings.passwordTooManyAttempts')
    } else {
      serverError.value = t('settings.passwordChangeFailed')
    }
  } finally {
    submitting.value = false
  }
}

function handleClose() {
  if (submitting.value) return
  emit('close')
}
</script>

<style scoped>
/* ═══════════════════════════════════════════════════════════════════════
 * PasswordChangeDialog - Liquid Glass Design System
 * iOS 26 Liquid Glass + visionOS Material
 * ═══════════════════════════════════════════════════════════════════════ */

/* ── Overlay (Dark Theme) ── */
.password-dialog-overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.4);
  -webkit-backdrop-filter: blur(8px);
  backdrop-filter: blur(8px);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
  padding: 16px;
}

/* Light Theme Overlay */
[data-theme="light"] .password-dialog-overlay {
  background: rgba(0, 0, 0, 0.3);
}

/* ── Dialog (Liquid Glass - Dark Theme) ── */
.password-dialog {
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
  width: 100%;
  max-width: 380px;

  /* Floating elevation */
  box-shadow:
    0 24px 80px rgba(0, 0, 0, 0.3),
    0 8px 24px rgba(0, 0, 0, 0.2),
    inset 0 0.5px 0 rgba(255, 255, 255, 0.15),
    inset 0 -0.5px 0 rgba(0, 0, 0, 0.05);
}

/* Light Theme Dialog */
[data-theme="light"] .password-dialog {
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

.password-dialog__header {
  font-size: 18px;
  font-weight: 600;
  color: rgba(255, 255, 255, 0.9);
  margin-bottom: 20px;
  text-align: center;
}

[data-theme="light"] .password-dialog__header {
  color: rgba(0, 0, 0, 0.85);
}

.password-dialog__field {
  margin-bottom: 16px;
}

.password-dialog__label {
  display: block;
  font-size: 13px;
  color: rgba(255, 255, 255, 0.6);
  margin-bottom: 6px;
}

[data-theme="light"] .password-dialog__label {
  color: rgba(0, 0, 0, 0.55);
}

.password-dialog__input {
  width: 100%;
  min-width: 0;
  padding: 12px 14px;
  font-size: 15px;
  border: 0.5px solid rgba(255, 255, 255, 0.1);
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.06);
  color: rgba(255, 255, 255, 0.9);
  outline: none;
  box-sizing: border-box;
  transition: border-color 0.2s ease;
}

[data-theme="light"] .password-dialog__input {
  border: 0.5px solid rgba(0, 0, 0, 0.1);
  background: rgba(0, 0, 0, 0.04);
  color: rgba(0, 0, 0, 0.85);
}

.password-dialog__input:focus {
  border-color: rgba(100, 180, 255, 0.5);
}

[data-theme="light"] .password-dialog__input:focus {
  border-color: rgba(37, 99, 235, 0.5);
}

.password-dialog__error {
  font-size: 13px;
  color: rgba(239, 68, 68, 0.9);
  margin-bottom: 12px;
  padding: 10px 12px;
  background: rgba(239, 68, 68, 0.1);
  border-radius: 10px;
}

.password-dialog__actions {
  display: flex;
  gap: 10px;
  margin-top: 20px;
}

.password-dialog__btn {
  flex: 1;
  padding: 12px;
  border: none;
  border-radius: 12px;
  font-size: 15px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
}

.password-dialog__btn--cancel {
  background: rgba(255, 255, 255, 0.1);
  color: rgba(255, 255, 255, 0.8);
}

[data-theme="light"] .password-dialog__btn--cancel {
  background: rgba(0, 0, 0, 0.06);
  color: rgba(0, 0, 0, 0.7);
}

.password-dialog__btn--submit {
  background: rgba(59, 130, 246, 0.9);
  color: #fff;
}

.password-dialog__btn--submit:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

@media (hover: hover) {
  .password-dialog__btn--cancel:hover {
    background: rgba(255, 255, 255, 0.15);
  }
  .password-dialog__btn--submit:not(:disabled):hover {
    background: rgba(59, 130, 246, 1);
  }
}

.password-dialog__btn--cancel:active {
  background: rgba(255, 255, 255, 0.12);
}

.password-dialog__btn--submit:not(:disabled):active {
  background: rgba(37, 99, 235, 1);
}
</style>
