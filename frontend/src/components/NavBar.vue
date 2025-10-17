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
      <HistorySVG />
      <span>History</span>
    </router-link>
  </div>
</template>

<script setup>
import LogoColorSVG from "./SVG/LogoColorSVG.vue";
import PlusSVG from "./SVG/PlusSVG.vue";
import HistorySVG from "./SVG/HistorySVG.vue";
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
