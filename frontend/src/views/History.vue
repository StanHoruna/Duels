<template>
  <div class="history">
    <div class="history__avatar_inner">
      <div class="history__avatar" @click="triggerFileUpload">
        <div class="history__avatar_wrap">
          <Avatar :source="userStore.userData?.image_url" />
        </div>
        <div class="history__avatar_icon">
          <UploadSVG />
          <input
              ref="fileInput"
              type="file"
              accept="image/*"
              @change="handleFileUpload"
              style="display: none"
          />
        </div>
      </div>
    </div>

    <div
        v-if="userStore.userData?.username"
        class="history__name"
        @click="isOpenModal = !isOpenModal"
    >
      <span>{{ userStore.userData.username }}</span>
      <EditSVG />
    </div>

    <div class="history__stats">
      <div class="history__stat">
        <div class="history__stat_value">{{ statsData?.wins_count || 0 }}</div>
        <div class="history__stat_name">🏆 Wins</div>
      </div>
      <div class="history__stat">
        <div class="history__stat_value">
          <USDCSVG />
          <span>{{ statsData?.earned_amount || 0 }}</span>
        </div>
        <div class="history__stat_name">Earnings</div>
      </div>
      <div class="history__stat">
        <div class="history__stat_value">{{ statsData?.participated || 0 }}</div>
        <div class="history__stat_name">Participated</div>
      </div>
    </div>

    <div class="history__tabs">
      <div
          class="history__tab"
          v-for="(tab, index) in tabs"
          :key="index"
          @click="activeTab = tab"
          :class="{ 'active': activeTab === tab }"
      >
        {{ tab }}
      </div>
    </div>

    <EmptyState
        v-if="!duelsData?.length"
        :title="emptyData.title"
        :text="emptyData.text"
        :button="emptyData.button"
        :redirect="emptyData.redirect"
    />

    <div v-else class="history__duels">
      <DuelCard
          :tab="activeTab"
          v-for="duel in duelsData"
          :duel="duel"
      />
    </div>

    <ProfileModal v-model="isOpenModal" />
  </div>
</template>

<script setup>
import {useUserStore} from "../store/userStore.js";
import USDCSVG from "../components/SVG/USDCSVG.vue";
import {computed, ref, watch} from "vue";
import UploadSVG from "../components/SVG/UploadSVG.vue";
import EditSVG from "../components/SVG/EditSVG.vue";
import {
  GetMyDuelsAsParticipant,
  GetUserStats,
  UploadAvatar
} from "../api/index.js";
import Avatar from "../components/Avatar.vue";
import ProfileModal from "../components/ProfileModal.vue";
import EmptyState from "../components/EmptyState.vue";
import {useToken} from "../composables/useToken.js";
import DuelCard from "../components/Duel/DuelCard.vue";

const userStore = useUserStore();
const { getToken } = useToken();

const duelsArr = ref([]);
const isOpenModal = ref(false);
const statsData = ref(null);

const fileInput = ref(null);

const tabs = ref(['Playing', 'Results', 'Created', 'Resolve']);
const activeTab = ref('Playing');

const emptyData = computed(() => {
  if (activeTab.value === "Created" || activeTab.value === "Resolve") {
    return {
      title: 'No created duels',
      text: `You didn’t create any duels yet`,
      button: "Create a duel",
      redirect: '/create',
    }
  } else if (activeTab.value === "Playing" || activeTab.value === "Results") {
    return {
      title: 'No active duels',
      text: `Currently there are no duels in play`,
      button: "Browse Duels",
      redirect: '/',
    }
  }
})

const duelsData = computed(() => {
  if (activeTab.value === "Created" || activeTab.value === "Resolve") {
    return duelsArr.value.filter((duel) => duel?.owner_id === userStore.userData?.id) || [];
  } else if (activeTab.value === "Playing") {
    return duelsArr.value.filter((duel) => duel?.status === 4) || [];
  } else if (activeTab.value === "Results") {
    return duelsArr.value.filter((duel) => duel?.status === 5 || duel?.status === 6) || [];
  }
})

const triggerFileUpload = () => {
  if (userStore.userData) {
    fileInput.value?.click();
  }
}

const handleFileUpload = (event) => {
  const file = event.target.files[0];
  if (!file) return;

  if (file.size > 5 * 1024 * 1024) {
    return;
  }

  const reader = new FileReader();
  reader.readAsDataURL(file);
  reader.onload = async () => {
    const formData = new FormData();
    formData.append('image', file);
    const resp = await UploadAvatar(formData);

    userStore.userData.image_url = resp.data.image_url;
  };

  event.target.value = '';
}

const getData = async () => {
  const token = await getToken();

  if (token) {
    const resp = await GetUserStats();
    statsData.value = resp.data;

    const res = await GetMyDuelsAsParticipant();
    duelsArr.value = res.data;
  }
}

watch(() => userStore.userData, async (value) => {
  if (value) {
    await getData();
  } else {
    statsData.value = null;
    duelsArr.value = [];
  }
}, { immediate: true });
</script>

<style scoped lang="scss">
.history {
  &__duels {
    display: flex;
    flex-direction: column;
    gap: 20px;
    margin-top: 16px;
  }
  &__avatar {
    &_inner {
      display: flex;
      justify-content: center;
      align-items: center;
    }
    &_wrap {
      width: 56px;
      height: 56px;
      border-radius: 50%;
      border: 2px solid #292929;
    }
    position: relative;
    cursor: pointer;
    &_icon {
      position: absolute;
      bottom: -0px;
      right: -6px;
      height: 24px;
      width: 24px;
      cursor: pointer;
    }
  }
  &__name {
    cursor: pointer;
    display: flex;
    align-items: center;
    justify-content: center;
    margin-top: 12px;
    gap: 8px;

    span {
      color: #F9F8F8;
      font-size: 14px;
      font-weight: 400;
      line-height: 100%;
    }
  }
  &__stats {
    display: flex;
    align-items: center;
    margin-top: 16px;
  }
  &__stat {
    width: calc(100% / 3);
    display: flex;
    flex-direction: column;
    gap: 5px;
    align-items: center;
    justify-content: center;
    border-left: 1px solid #292929;
    height: 40px;
    &:first-child {
      border-left: none;
    }
    &_value {
      display: flex;
      align-items: center;
      gap: 4px;
      &, span {
        color: #FFF;
        font-size: 16px;
        font-weight: 500;
        line-height: 100%;
      }
    }
    &_name {
      color: #6B6F89;
      font-size: 12px;
      font-weight: 400;
      line-height: 16px;
      letter-spacing: 0.175px;
    }
  }
  &__tabs {
    display: flex;
    align-items: center;
    justify-content: space-between;
    margin-top: 24px;
  }
  &__tab {
    color: #6B6F89;
    font-size: 14px;
    font-weight: 400;
    line-height: 22px;
    letter-spacing: 0.2px;
    border-bottom: 1px solid transparent;
    transition: 0.3s;
    cursor: pointer;
    padding-bottom: 4px;
    &.active {
      color: #D0F267;
      border-bottom: 1px solid #D0F267;
    }
  }
}
</style>
