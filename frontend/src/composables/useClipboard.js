import { useNotificationStore } from "../store/notificationStore.js";

export function useClipboard() {
  const notificationStore = useNotificationStore();

  const copyShareLink = async (duel_id) => {
    const link = window.location.origin + '/duel/' + duel_id;

    await copyToClipboard(link, 'Duel link copied to clipboard');
  }

  const copyToClipboard = async (text, message) => {
    try {
      await navigator.clipboard.writeText(text);
      notificationStore.addNotification({
        text: message,
        type: "success",
      });
    } catch (error) {
      notificationStore.addNotification({
        text: error.message,
        type: "error",
      });
    }
  };

  return { copyToClipboard, copyShareLink };
}
