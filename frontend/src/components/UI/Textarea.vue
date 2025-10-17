<template>
  <div class="textarea--wrap">
    <textarea
        :placeholder="placeholder"
        v-model="modelValue"
        class="textarea"
        :class="{'error': error}"
        :maxlength="maxlength"
        :disabled="disabled"
        @input="$emit('input')"
    />
  </div>
</template>

<script setup>
const modelValue = defineModel({default: ''});

defineProps({
  placeholder: String,
  maxlength: {type: String, default: 10000},
  isSearch: {type: Boolean, default: false},
  disabled: {type: Boolean, default: false},
  error: {type: String, default: ''},
});
</script>

<style scoped lang="scss">

.textarea {
  resize: none;
  min-height: 100px;
  max-height: 100px;
  height: 100px;
  outline: none;
  width: 100%;
  transition: 0.3s;
  display: flex;
  padding: 12px 16px;
  border-radius: 12px;
  border: 1px solid #292929;
  background: #1C1C1C;
  color: #F9F8F8;
  font-size: 14px;
  font-weight: 400;
  letter-spacing: 0.2px;
  line-height: 24px;
  &--wrap {
    width: 100%;
    position: relative;
  }
  &:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  &::placeholder {
    color: #6B6F89;
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
