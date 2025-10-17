import { defineStore } from "pinia";
import { ref } from "vue";

export const useNotificationStore = defineStore("notificationStore", () => {
  const notifications = ref([]); // success, error, warning, loading

  const addNotification = (notification, timeout = 3000, rawID) => {
    const id = rawID || String(Date.now() * Math.random());

    notifications.value.push({id, ...notification});

    if (timeout) setTimeout(() => removeNotification(id), timeout);
  };

  const removeNotification = (id) => {
    notifications.value = notifications.value.filter((item) => item?.id !== id);
  };

  const clearStore = () => {
    notifications.value = [];
  }

  return {
    notifications,
    addNotification,
    removeNotification,
    clearStore,
  };
});
