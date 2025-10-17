<template>
  <button
      :class="['buttonTag', variant]"
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
    default: '', // time, upload
  },
  name: {
    type: String,
    default: '',
  }
});

defineEmits(['click']);
</script>

<style scoped lang="scss">
.buttonTag {
  cursor: pointer;
  border: none;
  outline: none;
  display: flex;
  height: 28px;
  padding: 0 7px;
  justify-content: center;
  align-items: center;
  gap: 7px;
  border-radius: 10px;
  background: linear-gradient(90deg, rgba(64, 69, 94, 0.80) 1.07%, rgba(37, 40, 58, 0.80) 99.9%);
  span {
    color: #F9F8F8;
    font-size: 12px;
    font-weight: 400;
    line-height: 13px;
  }
  ::v-deep(svg) {
    width: 16px;
    height: 16px;
  }
  &.upload {
    ::v-deep(svg) {
      transform: rotate(-90deg);
      path {
        stroke: #F9F8F8;
      }
    }
  }
}
</style>
