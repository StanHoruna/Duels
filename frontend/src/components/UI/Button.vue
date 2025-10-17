<template>
  <button
      :class="['button', variant]"
      :disabled="disabled"
      @click="$emit('click')"
  >
    <slot name="icon"></slot>
    <span>{{ name }}</span>
  </button>
</template>

<script setup>
import { defineProps, defineEmits } from 'vue';

defineProps({
  variant: {
    type: String,
    default: '', // green, red, share, transparent
  },
  disabled: {
    type: Boolean,
    default: false,
  },
  name: {
    type: String,
    default: '',
  }
});

defineEmits(['click']);
</script>

<style scoped lang="scss">
.button {
  cursor: pointer;
  border: none;
  outline: none;
  display: flex;
  width: 100%;
  min-height: 48px;
  height: 48px;
  padding: 0 16px;
  justify-content: center;
  align-items: center;
  gap: 14px;
  border-radius: 12px;
  color: #191B1F;
  font-size: 16px;
  font-weight: 500;
  line-height: 100%;
  transition: 0.3s;
  &.red {
    background: #E44E2D;
  }
  &.green {
    background: #D0F267;
  }
  &.transparent {
    color: #D0F267;
    background: transparent;
  }
  &.share {
    background: #1C1C1C;
    flex-direction: row-reverse;
    color: #D0F267;
    ::v-deep(svg) {
      width: 20px;
      height: 20px;
      path {
        fill: #D0F267;
        stroke: #D0F267;
      }
    }
  }
  &:disabled {
    cursor: not-allowed;
    opacity: 0.5;
  }
}
</style>
