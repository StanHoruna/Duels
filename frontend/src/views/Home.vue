<template>
  <div class="home">
    <DuelCard @getDuel="getDuels" v-for="duel in duelsData" :duel="duel" />
  </div>
</template>

<script setup>
import {ref, watch} from "vue";
import {GetDuels, GetDuelsPublic} from "../api/index.js";
import DuelCard from "../components/Duel/DuelCard.vue";
import {useToken} from "../composables/useToken.js";
import {useUserStore} from "../store/userStore.js";

const { getToken } = useToken();

const userStore = useUserStore();

const duelsData = ref([]);

const getDuels = async () => {
  const token = await getToken();

  let resp;

  if (token) {
    resp = await GetDuels();
  } else {
    resp = await GetDuelsPublic();
  }

  duelsData.value = resp.data;
}

watch(() => userStore.userData, async () => {
  await getDuels();
}, { immediate: true });
</script>

<style lang="scss" scoped>
.home {
  display: flex;
  flex-direction: column;
  gap: 20px;
}
</style>
