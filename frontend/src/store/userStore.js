import { defineStore } from "pinia";
import {ref} from "vue";
import {GetUser, SignIn} from "../api/index.js";
import {useToken} from "../composables/useToken.js";
import {useNotificationStore} from "./notificationStore.js";
import {useWalletStore} from "./walletStore.js";

export const useUserStore = defineStore("userStore", () => {
  const walletStore = useWalletStore();

  const {setToken, getToken} = useToken();
  const notificationStore = useNotificationStore();

  const userData = ref(null);

  const signIn = async (publicAddress, signedMessage) => {
    try {
      const resp = await SignIn(publicAddress, signedMessage);

      const { jwt_info, user } = resp.data;

      setToken(jwt_info?.access_token, jwt_info?.refresh_token);

      userData.value = user;

      await getUserData();
    } catch (error) {
      if (error?.response?.data) {
        notificationStore.addNotification({ type: "error", text: error?.response?.data });
      }
    }
  }

  const getUserData = async () => {
    try {
      const token = await getToken();

      if (token) {
        if (!userData.value) {
          const userRes = await GetUser();
          userData.value = userRes.data.user;
        }

        await walletStore.getBalance();
      }
    } catch (error) {}
  }

  const clearStore = () => {
    userData.value = null;
  }

  return {
    userData,

    getUserData,
    signIn,
    clearStore,
  };
});
