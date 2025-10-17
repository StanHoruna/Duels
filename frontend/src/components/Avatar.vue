<template>
  <div
      class="avatar"
      :style="avatar && !isError ? {} : { background: bgGradient }"
  >
    <img
        v-if="avatar && !isError"
        :src="getFile(avatar)"
        @error="handleImageError"
        draggable="false"
        @dragstart.prevent
    />
    <span v-else>
      <LogoSVG/>
    </span>
  </div>
</template>

<script setup>
import {computed, ref} from "vue";
import {getFile} from "../helpers/filters.js";
import LogoSVG from "./SVG/LogoSVG.vue";

const props = defineProps({
  source: {type: String, required: true,},
});

const isError = ref(false);

const avatar = computed(() => props.source);

const colors = [
  ["#FF845E", "#D45246"], // Red
  ["#FEBB5B", "#F68136"], // Orange
  ["#B694F9", "#6C61DF"], // Violet
  ["#9AD164", "#46BA43"], // Green
  ["#53EDD6", "#28C9B7"], // Cyan
  ["#5CAFFA", "#408ACF"], // Blue
  ["#FF8AAC", "#D95574"], // Pink
];

const bgGradient = computed(() => {
  const randomIndex = Math.floor(Math.random() * colors.length);
  const [topColor, bottomColor] = colors[randomIndex];
  return `linear-gradient(135deg, ${topColor}, ${bottomColor})`;
});

const handleImageError = () => {
  isError.value = true;
}
</script>

<style scoped lang="scss">
.avatar {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
  text-transform: uppercase;
  overflow: hidden;
  border-radius: inherit;

  span {
    display: flex;
    align-items: center;
    justify-content: center;
    padding: 2px;

    ::v-deep(svg) {
      max-width: 100%;
      height: auto;
    }
  }

  img {
    width: 100%;
    height: 100%;
    border-radius: inherit;
    user-drag: none;
    -webkit-user-drag: none;
    pointer-events: none;
  }
}
</style>
