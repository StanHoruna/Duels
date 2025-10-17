<template>
  <div class="duel__title">
    <div class="duel__question">{{ duel?.question }}</div>
    <div class="duel__wrapper">
      <div class="duel__owner">
        <div class="duel__owner_img">
          <Avatar :source="duel?.owner_image_url" />
        </div>
        <span>{{ duel?.username }}</span>
      </div>

      <div class="duel__bank">
        <span>bank:</span>
        <UserSVG/>
        <span>{{ duel?.players_count }}</span>
        <div class="duel__bank_line"></div>
        <USDCSVG/>
        <span>{{ toFix(duel?.players_count * duel?.duel_price) }}</span>
      </div>
    </div>
  </div>
</template>

<script setup>
import {toFix} from "../../helpers/filters.js";
import USDCSVG from "../SVG/USDCSVG.vue";
import {defineProps} from "vue";
import UserSVG from "../SVG/UserSVG.vue";
import Avatar from "../Avatar.vue";

const props = defineProps({
  duel: { type: Object, required: true },
})

</script>

<style scoped lang="scss">
.duel {
  &__question {
    color: #F9F8F8;
    font-size: 14px;
    font-weight: 400;
    line-height: 18px;
    //-webkit-line-clamp: 1;
    //line-clamp: 1;
    //-webkit-box-orient: vertical;
    //display: -webkit-box;
    //overflow: hidden;
    //text-overflow: ellipsis;
  }
  &__owner {
    display: flex;
    gap: 4px;
    align-items: center;

    &_img {
      width: 20px;
      height: 20px;
      border-radius: 50%;
      border: 0.25px solid #292929;
    }

    span {
      color: #A5AABE;
      font-size: 12px;
      font-weight: 400;
      line-height: 14px;
      letter-spacing: 0.164px;
    }
  }
  &__wrapper {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-bottom: 12px;
    margin-top: 6px;
  }
  &__bank {
    display: flex;
    align-items: center;
    gap: 4px;
    height: 12px;

    &_line {
      background: rgba(255, 255, 255, 0.10);
      width: 1px;
      height: 100%;
    }

    span {
      color: #519BFC;
      font-size: 12px;
      font-weight: 400;
      line-height: 100%;
      letter-spacing: 0.164px;
    }

    ::v-deep(svg) {
      min-width: 14px;
      width: 14px;
      height: auto;
    }
  }
}
</style>
