import { defineStore } from "pinia";
import CookieManager from "../helpers/cookieManager.js";
import {useNotificationStore} from "./notificationStore.js";
import {useWalletStore} from "./walletStore.js";
import {useUserStore} from "./userStore.js";

export const useLogOutStore = defineStore("logOutStore", () => {
  const notificationStore = useNotificationStore();
  const walletStore = useWalletStore();
  const userStore = useUserStore();

  const handleLogOut = async () => {
    CookieManager.removeItem("access_token");
    CookieManager.removeItem("refresh_token");

    userStore.clearStore();
    notificationStore.clearStore();
    await walletStore.clearStore();
  }

  return {
    handleLogOut
  };
});
