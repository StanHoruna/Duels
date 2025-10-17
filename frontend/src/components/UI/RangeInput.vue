<template>
  <div class="range-input" ref="rangeContainer">
    <div
        class="range-input__tooltip"
        :style="{ left: tooltipLeft + 'px' }"
    >
      <span>{{ modelValue }}%</span>
      <ArrowSVG/>
    </div>

    <input
        ref="rangeInput"
        type="range"
        v-model="modelValue"
        :min="min"
        :max="max"
        :step="step"
        @input="updateValue"
    />
  </div>
</template>

<script setup>
import {nextTick, onMounted, ref, watch} from "vue";
import ArrowSVG from "../SVG/ArrowSVG.vue";

const props = defineProps({
  min: {type: Number, default: 0},
  max: {type: Number, default: 100},
  step: {type: Number, default: 1},
});

const modelValue = defineModel({default: 0});
const emits = defineEmits(["action"]);

const rangeContainer = ref(null);
const rangeInput = ref(null);
const tooltipLeft = ref(0);

const updateTooltipPosition = () => {
  const input = rangeInput.value;
  if (!input) return;

  const rangeWidth = input.offsetWidth;
  const thumbWidth = 24;
  const percent = (modelValue.value - props.min) / (props.max - props.min);
  let newLeft = percent * (rangeWidth - thumbWidth) + thumbWidth / 2;

  if (modelValue.value === props.min) {
    newLeft += 1;
  } else if (modelValue.value === props.max) {
    newLeft -= 1;
  }

  tooltipLeft.value = newLeft;
};

const updateValue = () => {
  emits("action", modelValue.value);
  updateTooltipPosition();
};

onMounted(async () => {
  await nextTick();
  updateTooltipPosition();
});

watch(modelValue, () => {
  updateTooltipPosition();
});
</script>

<style scoped lang="scss">
.range-input {
  position: relative;
  width: 100%;

  height: 58px;
  display: flex;
  align-items: flex-end;

  input[type="range"] {
    width: 100%;
    -webkit-appearance: none;
    height: 24px;
    background: transparent;
    cursor: pointer;

    &::-webkit-slider-runnable-track {
      height: 6px;
      border-radius: 20px;
      border: 1px solid #292929;
      background: #1c1c1c;
    }

    &::-webkit-slider-thumb {
      -webkit-appearance: none;
      position: relative;
      top: -10px;
      width: 24px;
      height: 24px;
      border-radius: 50%;
      border: 3px solid #141414;
      background: #d0f267;
      cursor: pointer;
    }
  }

  &__tooltip {
    position: absolute;
    top: 0;
    transform: translateX(-50%);
    display: flex;
    flex-direction: column;
    align-items: center;

    span {
      border-radius: 26px;
      background: #d0f267;
      display: flex;
      //width: 78px;
      width: 50px;
      height: 26px;
      padding: 0 8px;
      justify-content: center;
      align-items: center;
      color: #000;
      font-size: 14px;
      font-weight: 400;
      line-height: 130%;
    }
  }
}
</style>
