<template>
  <div class="duel__time_inner">
    <div class="duel__time">
      <ButtonTag
          :name="formattedTime"
          variant="time"
      />
      <ButtonTag
          :name="toFix(duel?.duel_price)"
          variant="time"
      >
        <template #icon>
          <USDCSVG/>
        </template>
      </ButtonTag>
    </div>
  </div>
</template>

<script setup>
import {toFix} from "../../helpers/filters.js";
import ButtonTag from "../UI/ButtonTag.vue";
import USDCSVG from "../SVG/USDCSVG.vue";
import {defineProps, onMounted, onUnmounted, ref} from "vue";

const props = defineProps({
  duel: { type: Object, required: true },
})

let timerInterval = null;
const formattedTime = ref("00h : 00m : 00s");

const updateTimer = () => {
  if (!props.duel?.event_date) return;

  const targetTime = new Date(props.duel.event_date).getTime();
  const now = Date.now();
  const diff = targetTime - now;

  if (diff <= 0) {
    formattedTime.value = "Waiting for results";
    return;
  }

  const hours = Math.floor(diff / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  const seconds = Math.floor((diff % (1000 * 60)) / 1000);

  formattedTime.value = `${hours.toString().padStart(2, "0")}h : ${minutes.toString().padStart(2, "0")}m : ${seconds.toString().padStart(2, "0")}s`;
};

onMounted(() => {
  updateTimer();
  timerInterval = setInterval(updateTimer, 1000);
});

onUnmounted(() => {
  clearInterval(timerInterval);
});
</script>

<style scoped lang="scss">
.duel {
  &__time {
    display: flex;
    align-items: center;
    gap: 8px;
    position: relative;
    z-index: 20;
    &_inner {
      display: flex;
      align-items: flex-end;
      width: 100%;
      height: 85px;
      bottom: 0;
      left: 0;
      position: absolute;
      &::after {
        content: '';
        position: absolute;
        bottom: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: linear-gradient(180deg, rgba(0, 0, 0, 0.00) 14.86%, #0A1214 100%);
        z-index: 10;
      }
    }
  }
}
</style>
