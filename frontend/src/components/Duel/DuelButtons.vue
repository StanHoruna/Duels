<template>
  <div v-if="buttonsVariant === 'result-cancel'" class="duel__buttons">
    <div class="duel__buttons_text">{{ duel?.cancellation_reason }}</div>
  </div>

  <div v-if="buttonsVariant === 'waiting'" class="duel__buttons">
    <div class="duel__buttons_text">Waiting for results</div>
  </div>

  <div v-if="buttonsVariant === 'resolve'" class="duel__buttons">
    <Button :disabled="isResolveLoading" @click="resolveHandler(0)" name="No wins" variant="red"/>
    <Button :disabled="isResolveLoading" @click="resolveHandler(1)" name="Yes wins" variant="green"/>
  </div>

  <div v-if="buttonsVariant === 'voted'" class="duel__buttons">
    <Button
        :disabled="true"
        :name="`Voted ${duel?.your_answer === 0 ? 'No' : 'Yes'}`"
        :variant="duel?.your_answer === 0 ? 'red' : 'green'"
    />
  </div>

  <div v-if="buttonsVariant === 'vote'" class="duel__buttons">
    <Button :disabled="isVoteLoading" @click="voteHandler(0)" name="No" variant="red"/>
    <Button :disabled="isVoteLoading" @click="voteHandler(1)" name="Yes" variant="green"/>
  </div>
</template>

<script setup>
import Button from "../UI/Button.vue";
import {JoinDuel, ResolveDuel, SignJoinDuel} from "../../api/index.js";
import {computed, defineProps, ref} from "vue";
import {useWalletStore} from "../../store/walletStore.js";
import {useUserStore} from "../../store/userStore.js";
import {useRoute} from "vue-router";
import {useNotificationStore} from "../../store/notificationStore.js";

const notificationStore = useNotificationStore();
const userStore = useUserStore();
const walletStore = useWalletStore();

const route = useRoute();

const props = defineProps({
  duel: { type: Object, required: true },
  tab: { type: String, default: null },
})

const emits = defineEmits(['getDuel']);

const isVoteLoading = ref(false);
const isResolveLoading = ref(false);

const buttonsVariant = computed(() => {
  const targetTime = new Date(props.duel?.event_date).getTime();
  const now = Date.now();
  const diff = targetTime - now;

  // status
  if (props.duel?.status === 5 || props.duel?.status === 6) {
    if (props.duel?.status === 5) {
      return 'result';
    }
    if (props.duel?.status === 6) {
      return 'result-cancel';
    }
  }
  // resolve
  if ((props.tab === 'Resolve' || route.name === 'duel') && userStore.userData?.id === props.duel?.owner_id) {
    return 'resolve';
  }
  // waiting for results
  if (diff <= 0) {
    return 'waiting';
  }
  // vote
  if (props.duel?.your_answer === null) {
    return 'vote';
  }
  // voted
  if (props.duel?.your_answer !== null) {
    return 'voted';
  }
})

const voteHandler = async (answer) => {
  isVoteLoading.value = true;

  const id = String(Date.now() * Math.random());

  try {
    if (!userStore.userData) {
      await walletStore.connect();
      await voteHandler(answer);
    } else {
      notificationStore.addNotification({
        type: 'loading',
        text: 'Please confirm transaction'
      }, 0, id);

      const duel_id = props.duel?.id;

      const resp = await SignJoinDuel(duel_id, answer);

      const tx_hash = await walletStore.sendTx(resp.data?.tx);

      if (tx_hash) {
        setTimeout(async () => {
          await JoinDuel(duel_id, answer, tx_hash);

          await walletStore.getBalance();

          notificationStore.addNotification({
            type: 'success',
            text: 'You’ve successfully joined the duel! <br> Now it’s time to wait for the results — good luck!'
          });

          props.duel.your_answer = answer;

          emits('getDuel');
        }, 1000)
      }
    }
  } catch (e) {
    notificationStore.addNotification({type: 'error', text: 'Somthing went wrong'});
  } finally {
    isVoteLoading.value = false;
    notificationStore.removeNotification(id);
  }
}

const resolveHandler = async (answer) => {
  try {
    isResolveLoading.value = true;

    const duel_id = props.duel?.id;

    await ResolveDuel(duel_id, answer);

    await userStore.getResolveCount();

    emits('getDuel');
  } catch (e) {
    notificationStore.addNotification({type: 'error', text: 'Somthing went wrong'});
  } finally {
    isResolveLoading.value = false;
  }
}
</script>

<style scoped lang="scss">
.duel {
  &__buttons {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    &_text {
      border-radius: 12px;
      background: #1C1C1C;
      width: 100%;
      color: #FFF;
      font-size: 12px;
      font-weight: 400;
      line-height: 16px;
      height: 48px;
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
}
</style>
