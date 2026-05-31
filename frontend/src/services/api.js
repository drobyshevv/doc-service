// frontend/src/services/api.js
import axios from 'axios';

const api = axios.create({
  baseURL: '/',
  headers: { 'Content-Type': 'application/json' },
});

//  Интерцептор запроса: добавляем токен и X-User-ID
api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token');
  const userStr = localStorage.getItem('user');
  
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  
  if (userStr) {
    try {
      const user = JSON.parse(userStr);
      if (user?.id) {
        config.headers['X-User-ID'] = user.id;
      }
    } catch (e) {
      console.error('Failed to parse user:', e);
    }
  }
  
  return config;
});

//  Интерцептор ответа: очистка токена при 401
api.interceptors.response.use(
  (res) => res,
  (error) => {
    if (error.response?.status === 401) {
      localStorage.removeItem('token');
      localStorage.removeItem('user');
    }
    return Promise.reject(error);
  }
);

// ============================================================================
//  API: Аутентификация
// ============================================================================
export const authAPI = {
  /** Регистрация нового пользователя */
  register: (email, password) => 
    api.post('/auth/register', { email, password }),
  
  /** Вход в аккаунт */
  login: (email, password) => 
    api.post('/auth/login', { email, password }),
};

// ============================================================================
//  API: Документы
// ============================================================================
export const documentsAPI = {
  list: (params = {}) => api.get('/documents', { params }),
  get: (id) => api.get(`/documents/${id}`),
  getMeta: (id) => api.get(`/documents/${id}/meta`),
  getVersions: (id) => api.get(`/documents/${id}/versions`),
  
  /** Загрузка новой версии документа */
  uploadVersion: (id, formData) => {
    return api.post(`/documents/${id}/versions`, formData, {
      headers: {
        'Content-Type': undefined,
      },
    });
  },
  
  /** Скачивание документа (текущей или конкретной версии) */
  download: (id, version = null) => {
    const url = version 
      ? `/documents/${id}/versions/${version}` 
      : `/documents/${id}`;
    return api.get(url, { responseType: 'blob' });
  },
  
  /** Откат к конкретной версии */
  rollback: (id, version) => 
    api.post(`/documents/${id}/versions/${version}/rollback`),
  
  delete: (id) => api.delete(`/documents/${id}`),
  update: (id, data) => api.put(`/documents/${id}`, data),
};

// ============================================================================
//  API: Поиск
// ============================================================================
export const searchAPI = {
  search: (query, params = {}) => 
    api.get('/search/', { params: { query, ...params } }),
  phrase: (phrase, params = {}) => 
    api.get('/search/phrase', { params: { phrase, ...params } }),
  suggest: (query, params = {}) => 
    api.get('/search/suggest', { params: { query, ...params } }),
  owner: (params = {}) => 
    api.get('/search/owner', { params }),
};

export default api;