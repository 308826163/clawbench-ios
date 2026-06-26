<template>
  <div class="skeleton-list">
    <div v-for="i in count" :key="i" class="skeleton-item">
      <div class="skeleton-icon" :style="iconStyle"></div>
      <div class="skeleton-text">
        <div class="skeleton-line skeleton-line-name" :style="nameStyle"></div>
        <div class="skeleton-line skeleton-line-meta" :style="metaStyle"></div>
      </div>
    </div>
  </div>
</template>

<script setup>
const props = defineProps({
  count: { type: Number, default: 8 },
  iconSize: { type: Number, default: 32 },
  nameWidth: { type: String, default: '60%' },
  metaWidth: { type: String, default: '40%' },
})

const iconStyle = { width: `${props.iconSize}px`, height: `${props.iconSize}px` }
const nameStyle = { width: props.nameWidth }
const metaStyle = { width: props.metaWidth }
</script>

<style scoped>
.skeleton-list {
  display: flex;
  flex-direction: column;
  gap: 0;
}

.skeleton-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 8px;
}

.skeleton-icon {
  border-radius: 8px;
  background: var(--bg-tertiary, #e5e5e5);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

.skeleton-text {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.skeleton-line {
  height: 12px;
  border-radius: 6px;
  background: var(--bg-tertiary, #e5e5e5);
  animation: skeleton-pulse 1.5s ease-in-out infinite;
}

@keyframes skeleton-pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
</style>
