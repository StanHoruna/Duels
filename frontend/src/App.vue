<template>
  <div class="page">
    <div class="page__container">
      <Header v-if="route.name !== 'duel'"/>
      <div
          class="page__content"
          :class="{ 'empty': route.name === 'duel' }"
      >
        <RouterView/>
      </div>
      <NavBar v-if="route.name !== 'duel'"/>
      <Notifications/>
    </div>
  </div>
</template>

<script setup>
import Header from "./components/Header.vue";
import NavBar from "./components/NavBar.vue";
import {useUserStore} from "./store/userStore.js";
import {onBeforeMount, onMounted} from "vue";
import {useRoute} from "vue-router";
import Notifications from "./components/Notifications.vue";

const route = useRoute();

const userStore = useUserStore();

onBeforeMount(async () => {
  await userStore.getUserData();
});
</script>

<style lang="scss">
@import "./assets/scss/main";

.page {
  background-image: url("assets/images/bcg.svg");
  background-size: 50vw;
  background-repeat: repeat;

  animation: animatedBackground 30s linear infinite;

  min-height: 100dvh;
  overflow: hidden;
  padding: 24px;
  display: flex;
  justify-content: center;
  align-items: center;

  &__container {
    width: 375px;
    height: 780px;
    border-radius: 24px;
    background: #141414;
    overflow: hidden;

    display: flex;
    flex-direction: column;
    justify-content: space-between;

    position: relative;
  }

  &__content {
    padding: 16px;
    height: 100%;
    overflow-y: auto;
    width: 100%;

    &::-webkit-scrollbar {
      width: 0;
      height: 0;
      display: none;
    }

    scrollbar-width: none;
    -ms-overflow-style: none;

    &.empty {
      padding: 0;
    }
  }
}

@keyframes animatedBackground {
  from { background-position: 0 0; }
  to { background-position: -100% -100%; }
}

@media (max-width: 500px) {
  .page {
    background: #141414;
    padding: 0;
    &__container {
      width: 100vw;
      height: 100dvh;
      border-radius: 0;
    }
  }
}
</style>
