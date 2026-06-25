<template>
  <div
    v-if="everOpened"
    class="tab-panel"
    :class="{
      'tab-panel-active': isActive,
      'tab-panel-slide-left': !isActive && direction === 'left',
      'tab-panel-slide-right': !isActive && direction === 'right'
    }"
  >
    <!-- Header -->
    <div v-if="!noHeader" class="bs-header" @click="handleHeaderClick">
      <slot name="header">
        <span class="bs-title">{{ title }}</span>
      </slot>
    </div>
    <!-- Body -->
    <div class="bs-body">
      <slot />
    </div>
    <!-- Footer slot -->
    <footer v-if="$slots.footer" class="bs-footer">
      <slot name="footer" />
    </footer>
  </div>
</template>

<script setup>
import { ref, watch, computed } from 'vue'

const props = defineProps({
  tabId: {
    type: String,
    required: true,
  },
  activeTab: {
    type: String,
    required: true,
  },
  title: {
    type: String,
    default: '',
  },
  noHeader: Boolean,
  direction: {
    type: String,
    default: 'right', // 'left' 或 'right'
  },
})

const emit = defineEmits(['header-click'])

const isActive = computed(() => props.activeTab === props.tabId)
const everOpened = ref(false)

watch(isActive, (val) => {
  if (val) {
    everOpened.value = true
  }
}, { immediate: true })

function handleHeaderClick() {
  emit('header-click')
}
</script>

<style>
.tab-panel {
  position: absolute;
  inset: 0;
  background: var(--bg-secondary, #fff);
  display: flex;
  flex-direction: column;
  opacity: 0;
  transition: opacity 300ms ease, transform 300ms ease;
  pointer-events: none;
}

.tab-panel-active {
  opacity: 1;
  transform: translateX(0);
  pointer-events: auto;
}

/* 从右侧滑入（向左切换） */
.tab-panel-slide-left {
  transform: translateX(20px);
}

/* 从左侧滑入（向右切换） */
.tab-panel-slide-right {
  transform: translateX(-20px);
}

/* Tab panels sit directly below AppHeader which already has a border-bottom;
   no need for another border on the tab header — child components like
   .dir-nav provide their own dividers where needed.
   Also hide the drag handle — tab headers are not bottom sheets. */
.tab-panel > .bs-header {
  border-bottom: none;
  box-shadow: none;
}

.tab-panel > .bs-header > .bs-handle {
  display: none;
}
</style>
