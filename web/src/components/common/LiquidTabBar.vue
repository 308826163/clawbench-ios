<template>
  <div
    class="dock-wrapper-main"
    ref="containerEl"
    @pointerdown.stop="onPointerDown"
    @click.stop.prevent
    @touchstart.stop
    @touchend.stop
    :style="{
      '--active-index': activeIndex,
      '--pointer-offset': `${pointerOffset}px`
    }"
  >
    <!-- 底座层 -->
    <div class="static-interactive-layer base-icons-layer">
      <div v-for="(tab, index) in tabs" :key="'base-' + index" class="custom-tab-item" :data-index="index">
        <div class="custom-tab-content">
          <slot name="icon" :tab="tab" :index="index" layer="base"></slot>
        </div>
      </div>
    </div>

    <!-- 水滴层 -->
    <div class="water-lens-stage">
      <div ref="indicatorEl" class="tab-active-indicator">
        <div class="lens-view-viewport">
          <div class="lens-mirror-container">
            <div v-for="(tab, index) in tabs" :key="'lens-' + index" class="custom-tab-item">
              <div class="custom-tab-content magnified">
                <slot name="icon" :tab="tab" :index="index" layer="lens"></slot>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, watch, onMounted, onUnmounted } from 'vue';

interface TabItem {
  name: string;
  icon?: any;
  [key: string]: any;
}

const props = withDefaults(defineProps<{
  modelValue: number;
  tabs: TabItem[];
  activeTabId?: string;
}>(), {
  modelValue: 0,
  tabs: () => [],
  activeTabId: ''
});

const emit = defineEmits(['update:modelValue', 'change']);

const activeIndex = ref(props.modelValue);
const containerEl = ref<HTMLElement | null>(null);
const indicatorEl = ref<HTMLElement | null>(null);

const isDragging = ref(false);
const pointerOffset = ref(0);

const triggerHaptic = (type = 'light') => {
  if (!navigator.vibrate) return;
  // light: 5ms 轻微反馈，medium: 10ms 吸附反馈
  navigator.vibrate(type === 'light' ? 5 : 10);
};

// --- 缓出 + 微弹参数 ---
const DECAY = 0.18;       // 缓出衰减系数（越大越快）
const BOUNCE = 5;         // 微弹像素
const SNAP_THRESHOLD = 0.6; // 吸附阈值

// --- 运行状态 ---
const currentOffset = ref(0);  // 实际渲染的偏移量
const targetOffset = ref(0);   // 目标应该在的偏移量
const phase = ref('idle');     // idle | approach | bounce | tabSwitch
let physicsRafId = null;       // 动画循环 ID

const stopPhysicsEngine = () => {
  if (physicsRafId) {
    cancelAnimationFrame(physicsRafId);
    physicsRafId = null; // 核心修复：彻底清空状态！
  }
};

const startPhysicsEngine = () => {
  if (!physicsRafId) {
    runPhysicsLoop();
  }
};

let pointerHandledClick = false;
let startX = 0;
let clickedTabIndex = -1; // 记录点击的 tab 索引
const anchorIndex = ref(props.modelValue);

const animateToTab = (targetIndex, indexDiff) => {
  const container = containerEl.value;
  if (!container) return;

  const containerWidth = container.offsetWidth - 16;
  const tabWidth = containerWidth / 5;

  // 停止物理引擎
  stopPhysicsEngine();

  // 用 pointerOffset 补偿 activeIndex 的跳变，视觉位置不变
  // activeIndex 从旧→新，pointerOffset 设为反向偏移量
  const compensation = -indexDiff * tabWidth;
  pointerOffset.value = compensation;
  currentOffset.value = compensation;
  targetOffset.value = 0;

  // 更新状态（CSS 基准变了，但 pointerOffset 补偿了）
  activeIndex.value = targetIndex;
  anchorIndex.value = targetIndex;

  // 启动 tabSwitch 动画：缓出 + 微弹归零
  phase.value = 'tabSwitch';
  startPhysicsEngine();
};

watch(() => props.modelValue, (newVal) => {
  if (newVal === undefined || newVal < 0 || newVal >= 5) return;

  const indexDiff = newVal - activeIndex.value;
  if (indexDiff !== 0) {
    // 外部触发的 tab 切换，触发动画
    animateToTab(newVal, indexDiff);
  }
}, { immediate: true });

watch(() => props.activeTabId, () => {
  const idx = props.modelValue;
  if (idx === undefined || idx < 0 || idx >= 5) return;

  const indexDiff = idx - activeIndex.value;
  if (indexDiff !== 0) {
    // 外部触发的 tab 切换，触发动画
    animateToTab(idx, indexDiff);
  }
});

const runPhysicsLoop = () => {
  const delta = targetOffset.value - currentOffset.value;

  if (phase.value === 'approach') {
    // 阶段一：缓出衰减，向目标移动
    const step = delta * DECAY;

    if (Math.abs(delta) < SNAP_THRESHOLD) {
      // 足够近，进入微弹阶段
      currentOffset.value = targetOffset.value;
      pointerOffset.value = targetOffset.value;
      phase.value = 'bounce';
      triggerHaptic('medium');
    } else {
      currentOffset.value += step;
      pointerOffset.value = currentOffset.value;
    }
  } else if (phase.value === 'tabSwitch') {
    // tab 切换动画：从补偿值缓出归零
    const step = delta * DECAY;

    if (Math.abs(delta) < SNAP_THRESHOLD) {
      // 足够近，进入微弹阶段
      currentOffset.value = targetOffset.value;
      pointerOffset.value = targetOffset.value;
      phase.value = 'bounce';
      triggerHaptic('medium');
    } else {
      currentOffset.value += step;
      pointerOffset.value = currentOffset.value;
    }
  } else if (phase.value === 'bounce') {
    // 阶段二：微弹回弹（从 +BOUNCE 衰减到 0）
    const bounceDelta = 0 - currentOffset.value;
    const bounceStep = bounceDelta * 0.2;

    if (Math.abs(bounceDelta) < 0.1) {
      // 微弹结束
      currentOffset.value = 0;
      pointerOffset.value = 0;
      phase.value = 'idle';
      stopPhysicsEngine();
      return;
    }

    currentOffset.value += bounceStep;
    pointerOffset.value = currentOffset.value;
  }

  physicsRafId = requestAnimationFrame(runPhysicsLoop);
};

const onPointerDown = (e: PointerEvent) => {
  stopPhysicsEngine(); // 瞬间冻结动画
  isDragging.value = true;
  startX = e.clientX;
  anchorIndex.value = activeIndex.value;

  // 检查是否点击了底座层的 tab 元素
  const clickedElement = (e.target as HTMLElement).closest('.custom-tab-item');
  if (clickedElement && clickedElement.dataset.index !== undefined) {
    clickedTabIndex = parseInt(clickedElement.dataset.index, 10);
  } else {
    clickedTabIndex = -1;
  }

  // 继承当前偏移位置
  currentOffset.value = pointerOffset.value;
  targetOffset.value = pointerOffset.value;
};

const onPointerMove = (e: PointerEvent) => {
  if (!isDragging.value) return;

  // 如果移动距离超过阈值，说明是滑动，重置点击记录
  if (Math.abs(e.clientX - startX) > 8) {
    clickedTabIndex = -1;
  }

  const container = containerEl.value;
  if (!container) return;

  let deltaX = e.clientX - startX;

  // 边界值计算
  const actualContainerWidth = container.offsetWidth - 16;
  const tabWidth = actualContainerWidth / 5;
  const maxLeft = 0 - (anchorIndex.value * tabWidth);
  const maxRight = actualContainerWidth - tabWidth - (anchorIndex.value * tabWidth);

  // 加入边缘橡皮筋阻力 (Rubber-banding)
  if (deltaX < maxLeft) {
    deltaX = maxLeft - Math.sqrt(maxLeft - deltaX) * 2;
  } else if (deltaX > maxRight) {
    deltaX = maxRight + Math.sqrt(deltaX - maxRight) * 2;
  }

  // 拖拽时直接跟手，不经过物理引擎
  currentOffset.value = deltaX;
  pointerOffset.value = deltaX;
};

const onPointerUp = (e: PointerEvent) => {
  if (!isDragging.value) return;
  isDragging.value = false;

  const container = containerEl.value;
  const indicator = indicatorEl.value;
  if (!container || !indicator) return;

  let closestIndex = activeIndex.value;
  let isClickSwitch = false;  // 标记是否是点击切换

  // 如果有记录的点击 tab 索引，直接使用
  if (clickedTabIndex >= 0) {
    closestIndex = clickedTabIndex;
    isClickSwitch = true;  // 标记为点击切换
    clickedTabIndex = -1; // 重置
  } else {
    // 计算最近的目标 tab
    const indicatorRect = indicator.getBoundingClientRect();
    const indicatorCenterX = indicatorRect.left + (indicatorRect.width / 2);

    let minDistance = Infinity;
    const items = container.querySelectorAll('.base-icons-layer .custom-tab-item');
    items.forEach((item, index) => {
      const itemRect = item.getBoundingClientRect();
      const itemCenterX = itemRect.left + (itemRect.width / 2);
      const distance = Math.abs(indicatorCenterX - itemCenterX);
      if (distance < minDistance) {
        minDistance = distance;
        closestIndex = index;
      }
    });
  }

  // 获取容器实际宽度计算单格步长
  const containerWidth = containerEl.value ? containerEl.value.offsetWidth - 16 : 0;
  const tabWidth = containerWidth / 5;

  const indexDiff = closestIndex - anchorIndex.value;

  // 修正物理坐标系
  if (isClickSwitch && indexDiff !== 0) {
    // 点击切换：用 pointerOffset 补偿 activeIndex 跳变，缓出归零
    stopPhysicsEngine();
    const compensation = -indexDiff * tabWidth;
    pointerOffset.value = compensation;
    currentOffset.value = compensation;
    targetOffset.value = 0;
    activeIndex.value = closestIndex;
    anchorIndex.value = closestIndex;
    phase.value = 'tabSwitch';
    startPhysicsEngine();
  } else {
    // 滑动吸附：缓出 + 微弹归位
    activeIndex.value = closestIndex;
    anchorIndex.value = closestIndex;
    currentOffset.value = currentOffset.value - (indexDiff * tabWidth);
    targetOffset.value = 0;
    stopPhysicsEngine();
    phase.value = 'approach';
    startPhysicsEngine();
  }

  emit('update:modelValue', closestIndex);
  emit('change', closestIndex);
};

onMounted(() => {
  window.addEventListener('pointermove', onPointerMove, { passive: false });
  window.addEventListener('pointerup', onPointerUp);
  window.addEventListener('pointercancel', onPointerUp);
  containerEl.value?.addEventListener('click', (e: MouseEvent) => {
    if (pointerHandledClick) {
      pointerHandledClick = false;
      e.stopPropagation();
      e.preventDefault();
    }
  }, true);
});

onUnmounted(() => {
  window.removeEventListener('pointermove', onPointerMove);
  window.removeEventListener('pointerup', onPointerUp);
  window.removeEventListener('pointercancel', onPointerUp);
  cancelAnimationFrame(physicsRafId);
});
</script>

<style scoped>
.dock-wrapper-main {
  position: relative;
  display: flex;
  flex-direction: column;
  border-radius: 24px;
  padding: 12px 8px;
  background: rgba(22, 24, 30, 0.5);
  border: 1px solid rgba(255, 255, 255, 0.06);
  backdrop-filter: blur(20px);
  -webkit-backdrop-filter: blur(20px);
  user-select: none;
  touch-action: none;
  width: 100%;
}

[data-theme="light"] .dock-wrapper-main {
  background: rgba(0, 0, 0, 0.03);
  border: 1px solid rgba(0, 0, 0, 0.1);
}

.static-interactive-layer {
  display: flex;
  position: relative;
  width: 100%;
  z-index: 5;
}

.custom-tab-item {
  flex: 0 0 20% !important;
  width: 20% !important;
  padding: 10px 0;
  display: flex;
  justify-content: center;
  align-items: center;
  cursor: pointer;
}

.custom-tab-content {
  display: flex;
  flex-direction: column;
  align-items: center;
  opacity: 0.65;
  transform: scale(1);
  transition: opacity 0.3s ease;
  color: rgba(255, 255, 255, 0.8);
}

[data-theme="light"] .custom-tab-content {
  color: rgba(0, 0, 0, 0.7);
}

.custom-tab-item.is-active .custom-tab-content {
  opacity: 1;
}

.water-lens-stage {
  position: absolute;
  inset: 0;
  pointer-events: none;
  z-index: 10;
}

.tab-active-indicator {
  position: absolute;
  top: 50%;
  left: 8px;
  height: 34px;
  width: calc((100% - 16px) / 5);
  border-radius: 999px;

  background: linear-gradient(135deg, rgba(0,0,0,0.15) 0%, rgba(0,0,0,0.05) 100%) !important;
  backdrop-filter: blur(8px) brightness(1.1) contrast(1.1) saturate(130%);
  -webkit-backdrop-filter: blur(8px) brightness(1.1) contrast(1.1) saturate(130%);

  border: 1px solid rgba(255, 255, 255, 0.25);

  box-shadow:
    0 4px 16px rgba(0, 0, 0, 0.2),
    inset 0 1px 2px rgba(255, 255, 255, 0.3) !important;

  will-change: transform;
  transform: translateY(-50%) translateX(calc(var(--active-index) * 100% + var(--pointer-offset)));
}

.lens-view-viewport {
  position: absolute;
  inset: -1px;
  border-radius: 999px;
  overflow: hidden;
}

.lens-mirror-container {
  display: flex;
  position: absolute;
  top: -8px;
  bottom: -8px;
  width: calc(100% / 0.2) !important;
  height: calc(100% + 16px);
  left: -1.2vw;

  color: var(--lens-tint, #0A84FF);

  /* translateX 百分比相对自身宽度(500%)，所以用 20% 代替 100% 来对应父元素宽度 */
  transform: translateX(calc(-1 * (var(--active-index) * 20% + var(--pointer-offset))));
  will-change: transform;
}

.lens-mirror-container .custom-tab-item {
  flex: 0 0 20% !important;
  width: 20% !important;
}

.custom-tab-content.magnified {
  opacity: 1 !important;
  transform: scale(1.15) !important;
  color: #0A84FF;
}

.custom-tab-item,
.custom-tab-item *,
.custom-tab-item::before,
.custom-tab-item::after {
  background: transparent !important;
  background-color: transparent !important;
  background-image: none !important;
  box-shadow: none !important;
}

[data-theme="light"] .tab-active-indicator {
  background: linear-gradient(135deg, rgba(0, 0, 0, 0.04) 0%, rgba(0, 0, 0, 0.01) 100%) !important;
  border: 1px solid rgba(0, 0, 0, 0.12);
  box-shadow:
    0 4px 16px rgba(0, 0, 0, 0.1),
    inset 0 0 0 1.5px rgba(0, 0, 0, 0.08),
    inset 0 1px 4px rgba(255, 255, 255, 0.8),
    inset 0 -2px 6px rgba(0, 0, 0, 0.04) !important;
}

:deep(*) {
  -webkit-tap-highlight-color: transparent !important;
}
</style>
