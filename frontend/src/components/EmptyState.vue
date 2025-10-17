<template>
  <div class="empty">
    <EmptySVG />
    <div class="empty__title">{{ title }}</div>
    <div class="empty__text">{{ text }}</div>
    <Button @click="redirectHandler" variant="green" :name="button" />
  </div>
</template>

<script setup>
import Button from "./UI/Button.vue";
import EmptySVG from "./SVG/EmptySVG.vue";
import {useRouter} from "vue-router";
import {useWalletStore} from "../store/walletStore.js";

const router = useRouter();

const walletStore = useWalletStore();

const props = defineProps({
  title: { type: String, default: '' },
  text: { type: String, default: '' },
  button: { type: String, default: '' },
  redirect: { type: String, default: '' },
})

const redirectHandler = async () => {
  if (props.redirect?.length) {
    if (props.redirect === 'connect') {
      await walletStore.connect();
    } else {
      await router.push(props.redirect);
    }
  }
}
</script>

<style scoped lang="scss">
.empty {
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
  padding: 50px 0;

  &__title {
    color: #F9F8F8;
    text-align: center;
    font-size: 24px;
    font-weight: 400;
    line-height: 32px;
    margin-top: 16px;
    margin-bottom: 20px;
  }
  &__text {
    margin-bottom: 16px;
    color: #A5AABE;
    text-align: center;
    font-size: 12px;
    font-weight: 400;
    line-height: 18px;
    letter-spacing: 0.164px;
  }
}
</style>
