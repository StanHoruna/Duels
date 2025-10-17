<template>
  <div v-if="isOpen" class="modal-overlay" v-on-click-outside="closeModal">
    <div v-on-click-outside="closeModal" class="modal-content">
      <div class="modal-title">Set Up Your Profile</div>

      <label>
        <span>Username</span>
        <span>{{ username?.length }}/20</span>
      </label>
      <Input
          v-model="username"
          maxlength="20"
          @onEnter="saveModal"
      />

      <div class="modal-buttons">
        <Button @click="closeModal" name="Skip" variant="transparent" />
        <Button @click="saveModal" name="Save" variant="green" />
      </div>
    </div>
  </div>
</template>

<script setup>
import { vOnClickOutside } from "@vueuse/components";
import {useUserStore} from "../store/userStore.js";
import Input from "./UI/Input.vue";
import Button from "./UI/Button.vue";
import {UpdateUsername} from "../api/index.js";
import {ref, watch} from "vue";
import {cloneDeep} from "lodash";

const isOpen = defineModel({default: false})

const username = ref('');

const userStore = useUserStore();

const closeModal = () => {
  isOpen.value = false;
};

const saveModal = async () => {
  await UpdateUsername(username.value);

  userStore.userData.username = cloneDeep(username.value);

  closeModal();
}

watch(() => isOpen.value, () => {
  if (isOpen.value) {
    username.value = cloneDeep(userStore.userData.username);
  }
})
</script>

<style scoped lang="scss">
.modal {
  &-overlay {
    backdrop-filter: blur(5px);
    position: absolute;
    top: 0;
    left: 0;
    width: 100%;
    height: 100%;
    background: rgba(0, 0, 0, 0.5);
    display: flex;
    align-items: flex-end;
    justify-content: center;
    z-index: 1000;
    transition: 0.3s;
  }
  &-content {
    width: 100%;
    background: #1C1C1C;
    padding: 16px;
    display: flex;
    flex-direction: column;
    border-radius: 10px 10px 0 0;
    label {
      display: flex;
      align-items: center;
      justify-content: space-between;
      color: #C9CCD8;
      font-size: 12px;
      font-weight: 300;
      line-height: 18px;
      letter-spacing: 0.2px;
      margin-bottom: 8px;
    }
  }
  &-title {
    color: #A5AABE;
    text-align: center;
    font-size: 20px;
    font-weight: 500;
    line-height: 24px;
    margin-bottom: 24px;
  }
  &-buttons {
    margin-top: 24px;
    display: flex;
    align-items: center;
    gap: 12px;
  }
}
</style>
