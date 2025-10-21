<template>
  <header class="header">
    <div class="header__left">
      <div
          class="header__left_item"
          :class="{'active': true}"
      >
        <USDCSVG/>
        <span>{{ toFix(walletStore.balanceData.ucdc) }}</span>
      </div>
    </div>
    <router-link to="/" class="header__center">
      <LogoSVG :class="{ spin: isSpinning }"/>
    </router-link>
    <div class="header__right">
      <button @click="walletHandler">
        <span v-if="userStore.userData" class="header__right_item">
          {{ userStore.userData?.public_address.slice(0, 5) }}...
        </span>
        <span v-else class="header__right_item">Sing in</span>

        <LogOutSVG v-if="userStore.userData"/>
        <SingInSVG v-else/>
      </button>
    </div>
  </header>
</template>

<script setup>
import {useWalletStore} from "../store/walletStore.js";
import {useUserStore} from "../store/userStore.js";
import {useLogOutStore} from "../store/logOutStore.js";
import LogoSVG from "./SVG/LogoSVG.vue";
import USDCSVG from "./SVG/USDCSVG.vue";
import SingInSVG from "./SVG/SingInSVG.vue";
import LogOutSVG from "./SVG/LogOutSVG.vue";
import {toFix} from "../helpers/filters.js";
import {ref, watch} from "vue";
import {useRoute} from "vue-router";

const route = useRoute();

const walletStore = useWalletStore();
const userStore = useUserStore();
const logOutStore = useLogOutStore();

const isSpinning = ref(false);

const walletHandler = async () => {
  if (userStore.userData) {
    await logOutStore.handleLogOut();
  } else {
    await walletStore.connect();
  }
};

watch(() => route.fullPath, () => {
  if (isSpinning.value) return;

  isSpinning.value = true
  setTimeout(() => {
    isSpinning.value = false
  }, 1000)
}, { immediate: false });
</script>

<style scoped lang="scss">
@import "../assets/scss/main.scss";

.header {
  display: flex;
  width: 100%;
  height: 64px;
  padding: 16px;
  align-items: center;
  justify-content: space-between;
  gap: 16px;

  &__left, &__center, &__right {
    width: calc(100% / 3);
    display: flex;
    align-items: center;
  }

  &__center {
    justify-content: center;
  }

  &__left {
    &_item {
      border-radius: 100px;
      border: 1px solid #292929;
      background: #1B1B1B;
      display: flex;
      height: 32px;
      padding: 0 12px;
      align-items: center;
      gap: 8px;
      width: max-content;

      span {
        color: #6B6F89;
        font-size: 12px;
        font-weight: 400;
        line-height: 100%;
      }

      &.active {
        span {
          color: #F6FCE1;
        }
      }
    }
  }

  &__right {
    justify-content: flex-end;

    &_item {
      color: #F6FCE1;
      font-size: 12px;
      font-weight: 400;
      letter-spacing: -0.233px;
      line-height: 100%;
    }

    button {
      border-radius: 100px;
      border: 1px solid #292929;
      background: #1B1B1B;
      display: flex;
      height: 32px;
      padding: 0 12px;
      align-items: center;
      gap: 8px;
      cursor: pointer;
    }
  }
}

.spin {
  animation: logo-spin 1s ease-out;
}

@keyframes logo-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(720deg);
  }
}
</style>
