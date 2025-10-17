<template>
  <div class="notifications-container">
    <div class="notification-stack">
      <div
          v-for="item in notificationStore.notifications"
          :key="item.id"
          :class="['notification', item.type]"
      >
        <div class="circle_svg" v-if="item.type === 'success'">
          <CheckSVG />
        </div>
        <div class="circle_svg" v-if="item.type === 'error'">
          <CrossSVG />
        </div>
        <div class="load_svg" v-if="item.type === 'loading'">
          <LoaderSVG />
        </div>
        <div class="notification__message">
          <div v-if="item.type === 'loading'" class="bodySmall">Please wait.</div>
          <div v-if="item.text?.length" class="bodySmall">{{ item.text }}</div>
        </div>
        <div class="notification__close">
          <XSVG @click="notificationStore.removeNotification(item.id)" />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup>
import {useNotificationStore} from "../store/notificationStore.js";
import LoaderSVG from "./SVG/LoaderSVG.vue";
import CheckSVG from "./SVG/CheckSVG.vue";
import CrossSVG from "./SVG/CrossSVG.vue";
import XSVG from "./SVG/XSVG.vue";

const notificationStore = useNotificationStore();
</script>

<style scoped lang="scss">
.notification-stack {
  position: absolute;
  top: 16px;
  left: 16px;
  display: flex;
  flex-direction: column;
  gap: 12px;
  z-index: 100000;
  width: calc(100% - 32px);
  .notification {
    background: #121212;
    display: flex;
    align-items: center;
    width: calc(100%);
    height: auto;
    padding: 16px 12px;
    gap: 12px;
    color: #FFF;
    font-size: 12px;
    font-weight: 500;
    position: relative;
    border-radius: 12px;
    border: 1px solid #222;
    &__message {
      display: flex;
      flex-direction: column;
    }
    &__close {
      margin-left: auto;
      cursor: pointer;
    }
  }
}

.load_svg {
  animation: spin 4s infinite linear;
  width: 48px;
  height: 48px;
  ::v-deep(svg) {
    width: 48px;
    height: 48px;
  }
}

.circle_svg {
  width: 48px;
  height: 48px;
  display: flex;
  align-items: center;
  justify-content: center;
  ::v-deep(svg) {
    width: 36px;
    height: 36px;
  }
}

@keyframes spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}
</style>
