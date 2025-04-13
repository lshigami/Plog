import axios from 'axios';

// Dynamically determine the API base URL based on the current environment
const getBaseUrl = () => {
  // In production, use the current domain (window.location.origin)
  // In development, use localhost
  return process.env.NODE_ENV === 'production'
    ? `${window.location.origin}/api/v1`
    : 'http://localhost:8080/api/v1';
};

const API_URL = getBaseUrl();

const api = axios.create({
  baseURL: API_URL,
});

// Add a request interceptor to add the auth token to requests
api.interceptors.request.use(
  (config) => {
    const token = localStorage.getItem('token');
    if (token) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error) => {
    return Promise.reject(error);
  }
);

export const register = (username, password) => {
  return api.post('/register', { username, password });
};

export const login = (username, password) => {
  return api.post('/login', { username, password });
};

export const getPosts = (limit = 10, offset = 0) => {
  return api.get(`/posts?limit=${limit}&offset=${offset}`);
};

export const getPost = (id) => {
  return api.get(`/posts/${id}`);
};

export const getMyPosts = () => {
  return api.get('/my-posts');
};

export const createPost = (title, content) => {
  return api.post('/posts', { title, content });
};

export const updatePost = (id, title, content) => {
  return api.put(`/posts/${id}`, { title, content });
};

export default api; 