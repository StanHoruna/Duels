<template>
  <div class="nav">
    <router-link to="/create" class="nav__item" active-class="active">
      <PlusSVG />
      <span>Create</span>
    </router-link>
    <router-link to="/" class="nav__item nav__logo" active-class="active">
      <LogoColorSVG />
      <span>Duels</span>
    </router-link>
    <router-link to="/history" class="nav__item" active-class="active">
      <div class="nav__item_wrap">
        <HistorySVG />
        <div v-if="userStore.resolveCount" class="nav__item_count">
          <p>{{ userStore.resolveCount }}</p>
        </div>
      </div>
      <span>History</span>
    </router-link>
  </div>
</template>

<script setup>
import LogoColorSVG from "./SVG/LogoColorSVG.vue";
import PlusSVG from "./SVG/PlusSVG.vue";
import HistorySVG from "./SVG/HistorySVG.vue";
import {watch} from "vue";
import {useUserStore} from "../store/userStore.js";

const userStore = useUserStore();

watch(() => userStore.userData, async (value) => {
  if (value) {
    await userStore.getResolveCount();
  }
}, { immediate: true });
</script>

<style scoped lang="scss">
.nav {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  height: 64px;
  padding: 16px;
  border-top: 1px solid #292929;
  background: #111;
  &__item {
    width: calc(100% / 3);
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    span {
      color: #6B6F89;
      font-size: 14px;
      font-weight: 400;
      line-height: 100%;
      letter-spacing: 0.167px;
      transition: 0.3s;
    }
    &:first-child {
      justify-content: flex-start;
    }
    &:last-child {
      justify-content: flex-end;
    }
    ::v-deep(svg) {
      path {
        transition: 0.3s;
      }
    }
    &.active, &:hover {
      span {
        color: #D0F267;
      }
      ::v-deep(svg) {
        path {
          stroke: #D0F267;
        }
      }
    }
    &_wrap {
      position: relative;
      display: flex;
      align-items: center;
    }
    &_count {
      position: absolute;
      border-radius: 50%;
      width: 16px;
      height: 16px;
      background: #1C1C1C;
      border: 1px solid #D0F267;
      display: flex;
      align-items: center;
      justify-content: center;
      bottom: -4px;
      right: -5px;
      overflow: hidden;
      &:after {
        content: '';
        background: linear-gradient(90deg, rgba(208, 242, 103, 0.40) 0%, rgba(28, 28, 28, 0.13) 87.98%);
        width: 100%;
        height: 100%;
        top: 0;
        left: 0;
        position: absolute;
      }
      p {
        color: #D0F267;
        font-size: 8px;
        font-weight: 400;
        line-height: 100%;
        margin-top: 1px;
        margin-right: 0.7px;
      }
    }
  }
  &__logo {
    &.active, &:hover {
      ::v-deep(svg) {
        path {
          stroke: none !important;
          fill: #D0F267;
        }
      }
    }
  }
}
</style>
