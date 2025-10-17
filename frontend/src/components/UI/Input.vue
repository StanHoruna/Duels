<template>
  <div class="input--wrap">
    <input
        ref="inputRef"
        :type="type"
        :inputmode="inputmode"
        :placeholder="placeholder"
        v-model="modelValue"
        class="input"
        :class="{'error': error}"
        :maxlength="maxlength"
        :disabled="disabled"
        @input="$emit('input')"
        @keydown.enter="$emit('onEnter')"
    />
    <slot name="icon"></slot>
  </div>
</template>

<script setup>
import {ref} from "vue";

const modelValue = defineModel({default: ''});
const inputRef = ref(null);

defineProps({
  placeholder: String,
  type: {type: String, default: "text"},
  inputmode: {type: String, default: "text"},
  maxlength: {type: String, default: 10000},
  disabled: {type: Boolean, default: false},
  error: {type: String, default: ''},
});

defineExpose({
  focus: () => inputRef.value?.focus()
});
</script>

<style scoped lang="scss">
.input {
  outline: none;
  width: 100%;
  height: 48px;
  transition: 0.3s;
  line-height: 100%;
  display: flex;
  padding: 0 16px;
  border-radius: 12px;
  border: 1px solid #292929;
  background: #1C1C1C;
  color: #F9F8F8;
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0.2px;
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  &::placeholder {
    color: #6B6F89;
  }
  &--wrap {
    width: 100%;
    position: relative;
    ::v-deep(svg) {
      width: 22px;
      height: auto;
      position: absolute;
      right: 16px;
      top: 50%;
      transform: translate(0%, -50%);
    }
  }
  &:focus {
    border: 1px solid #D0F267;
  }
  &.error {
    border-color: #E44E2D;
    color: #7D2A16;
  }
}
</style>
