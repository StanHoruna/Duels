<template>
  <div class="duel">
    <div class="duel__image" @click="openDetails">
      <img :src="getFile(duel?.bg_url)" alt="Duel image">

      <DuelTime :duel="duel" />

      <button class="duel__share">
        <ShareSVG/>
      </button>
    </div>

    <DuelTitle :duel="duel" />

    <DuelButtons :tab="tab" :duel="duel" />
  </div>
</template>

<script setup>
import {getFile} from "../../helpers/filters.js";
import ShareSVG from "../SVG/ShareSVG.vue";
import {defineProps} from "vue";
import {useRouter} from "vue-router";
import DuelButtons from "./DuelButtons.vue";
import DuelTime from "./DuelTime.vue";
import DuelTitle from "./DuelTitle.vue";

const router = useRouter();

const props = defineProps({
  duel: { type: Object, required: true },
  tab: { type: String, default: null },
})

const openDetails = () => {
  router.push({ name: 'duel', params: { id: props.duel?.id } });
}
</script>

<style scoped lang="scss">
.duel {
  &__share {
    border-radius: 100px;
    border: 1px solid #292929;
    background: #1C1C1C;
    display: flex;
    width: 32px;
    height: 32px;
    justify-content: center;
    align-items: center;
    outline: none;
    position: absolute;
    top: 8px;
    right: 8px;
    cursor: pointer;
    z-index: 100;
  }

  &__image {
    cursor: pointer;
    width: 100%;
    height: 130px;
    overflow: hidden;
    border-radius: 12px;
    position: relative;
    margin-bottom: 7px;

    ::v-deep(.duel__time) {
      padding: 8px;
    }

    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
      border-radius: 12px;
    }
  }
}
</style>
