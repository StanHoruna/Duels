import axios from "axios";
import CookieManager from "../helpers/cookieManager.js";

let refreshPromise = null;

export function useToken() {
  const getToken = async () => {
    let token = CookieManager.getItem("access_token");

    if (token) return token;

    if (refreshPromise) return refreshPromise;

    let rt = CookieManager.getItem("refresh_token");

    if (rt) {
      return await refreshToken();
    } else {
      return null;
    }
  };

  const refreshToken = async () => {
    if (refreshPromise) return refreshPromise;

    const rt = CookieManager.getItem("refresh_token");

    refreshPromise = axios
      .post(`${import.meta.env.VITE_API_URL}/auth/refresh`, {}, {
        headers: { Authorization: rt },
      })
      .then(response => {
        const { access_token, refresh_token } = response.data?.jwt_info;

        setToken(access_token, refresh_token);

        return access_token;
      })
      .catch(error => {
        throw error;
      })
      .finally(() => {
        refreshPromise = null;
      });

    return refreshPromise;
  };

  const setToken = (access_token, refresh_token) => {
    CookieManager.setItem("access_token", access_token, Date.now() + 3600000); // 1h
    CookieManager.setItem("refresh_token", refresh_token, Date.now() + 2592000000); // 720h
  };

  return { getToken, setToken, refreshToken };
}
