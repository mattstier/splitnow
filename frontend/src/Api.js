import axios from "axios";

const getToken = () => localStorage.getItem('token')
const api = axios.create();

// attach bearer with token to every outgoing request
api.interceptors.request.use((config) => {
    const token = getToken()
    if (token) config.headers.Authorization = `Bearer ${token}`;
    return config;
});

// handle response from the backend 
// logout on 401 (token expired), re-throw everything else
api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token');
      window.location.reload();
    }
    return Promise.reject(err);
  }
);

export default api;
