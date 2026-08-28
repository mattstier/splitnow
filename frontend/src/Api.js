import axios from "axios";
import { useAuthStore } from "./stores/AuthStore.js";

const api = axios.create({
  baseURL: import.meta.env.VITE_API_ENDPOINT || '',
  headers: {
    'Content-Type': 'application/json'
  }
})

// attach bearer with token to every outgoing request
api.interceptors.request.use((config) => {
  const token = useAuthStore.getState().token
  if (token) config.headers.Authorization = `Bearer ${token}`;
  return config;
});

// handle response from the backend
// logout on 401 (token expired), re-throw everything else
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      useAuthStore.getState().logout();
    }
    return Promise.reject(err);
  }
);

export default api;
