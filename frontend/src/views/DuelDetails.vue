<template>
  <div class="duelDetails">
    <div class="duelDetails__image">
      <img :src="getFile(duelData?.bg_url)" alt="Duel image">
      <DuelTime :duel="duelData" />

      <div class="duelDetails__back_inner">
        <div class="duelDetails__back">
          <ArrowBackSVG @click="backHandler" />
          <span>Duel #{{ duelData?.room_number }}</span>
        </div>
      </div>
    </div>

    <div class="duelDetails__content">
      <div class="duelDetails__container">
        <DuelTitle :duel="duelData" />

        <Button name="Share this Duel" variant="share">
          <template #icon>
            <ShareSVG />
          </template>
        </Button>

        <div class="duelDetails__players">
          <div class="duelDetails__players_title">Participants</div>

          <div class="duelDetails__player_wrap">
            <div
                class="duelDetails__player"
                v-for="(player, index) in duelPlayers"
            >
              <div class="duelDetails__player_index">{{ index + 1 }}</div>
              <div class="duelDetails__player_name">
                <div class="duelDetails__player_avatar">
                  <Avatar :source="player.image_url" />
                </div>
                <span>{{ player.username }}</span>
              </div>
              <div
                  class="duelDetails__player_voted"
                  :class="`${player?.answer === 0 ? 'red' : 'green'}`"
              >
                Voted {{ player?.answer === 0 ? 'No' : 'Yes' }}
              </div>
            </div>
          </div>
        </div>
      </div>

      <DuelButtons @getDuel="getDuel" :duel="duelData" />
    </div>
  </div>
</template>

<script setup>
import {useRoute, useRouter} from "vue-router";
import {ref, watch} from "vue";
import {GetDuelByID, GetDuelByIDPublic} from "../api/index.js";
import {getFile} from "../helpers/filters.js";
import DuelTime from "../components/Duel/DuelTime.vue";
import DuelTitle from "../components/Duel/DuelTitle.vue";
import DuelButtons from "../components/Duel/DuelButtons.vue";
import ArrowBackSVG from "../components/SVG/ArrowBackSVG.vue";
import Button from "../components/UI/Button.vue";
import ShareSVG from "../components/SVG/ShareSVG.vue";
import Avatar from "../components/Avatar.vue";
import {useToken} from "../composables/useToken.js";
import {useUserStore} from "../store/userStore.js";

const route = useRoute();
const router = useRouter();

const { getToken } = useToken();

const userStore = useUserStore();

const duelData = ref(null);
const duelPlayers = ref([]);

const backHandler = () => {
  router.push({ name: 'home' });
}

const getDuel = async () => {
  const token = await getToken();

  let resp;

  if (token) {
    resp = await GetDuelByID(route.params.id);
    duelData.value = resp.data.duel;
    duelPlayers.value = resp.data.players;
  } else {
    resp = await GetDuelByIDPublic(route.params.id);
    duelData.value = resp.data;
  }
}

watch(() => userStore.userData, async () => {
  await getDuel();
}, { immediate: true });
</script>

<style scoped lang="scss">
.duelDetails {
  height: 100%;
  &__content {
    padding: 16px;
    height: calc(100% - 240px);
    display: flex;
    flex-direction: column;
    justify-content: space-between;
    gap: 24px;
  }
  &__container {
    height: calc(100% - 48px - 24px);
    display: flex;
    flex-direction: column;
  }
  &__image {
    width: 100%;
    height: 240px;
    overflow: hidden;
    position: relative;
    img {
      width: 100%;
      height: 100%;
      object-fit: cover;
    }
    ::v-deep(.duel__time) {
      padding: 16px;
    }
  }
  &__back {
    &_inner {
      &::after {
        content: '';
        position: absolute;
        top: 0;
        left: 0;
        width: 100%;
        height: 100%;
        background: linear-gradient(0deg, rgba(0, 0, 0, 0.00) 14.86%, #0A1214 100%);
        z-index: 10;
      }
    }
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 64px;
    padding: 16px;
    gap: 16px;
    display: flex;
    align-items: center;
    z-index: 20;
    span {
      color: #F9F8F8;
      font-size: 16px;
      font-weight: 500;
      line-height: 100%;
    }
    ::v-deep(svg) {
      cursor: pointer;
      path {
        transition: 0.3s;
      }
      &:hover {
        path {
          stroke: #D0F267;
        }
      }
    }
  }
  &__players {
    margin-top: 16px;
    height: 100%;
    overflow: hidden;
    &_title {
      color: #F9F8F8;
      font-size: 16px;
      font-weight: 500;
      line-height: 24px;
      margin-bottom: 4px;
    }
  }
  &__player {
    display: grid;
    grid-template-columns: 16px 1fr auto;
    gap: 12px;
    border-bottom: 1px solid #292929;
    padding: 12px 0;
    &_wrap {
      height: calc(100% - 24px);
      overflow-y: auto;
      &::-webkit-scrollbar {
        width: 0;
        height: 0;
        display: none;
      }
      scrollbar-width: none;
      -ms-overflow-style: none;
    }
    &:last-child {
      border-bottom: none;
      padding-bottom: 0;
    }
    &_index {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 16px;
      color: #A5AABE;
      font-size: 12px;
      font-weight: 400;
      line-height: 20px;
    }
    &_avatar {
      border-radius: 50%;
      width: 20px;
      height: 20px;
    }
    &_name {
      display: flex;
      align-items: center;
      gap: 8px;
      span {
        color: #F9F8F8;
        font-size: 14px;
        font-weight: 400;
        line-height: 20px;
      }
    }
    &_voted {
      font-size: 12px;
      font-weight: 400;
      line-height: 20px;
      letter-spacing: 0.182px;
      &.red {
        color: #E44E2D;
      }
      &.green {
        color: #D0F267;
      }
    }
  }
}
</style>
