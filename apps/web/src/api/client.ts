import axios, { AxiosError } from 'axios';
import { ApiError } from './types';

// Базовый URL для локальной разработки (в будущем можно брать из ENV)
const BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const apiClient = axios.create({
  baseURL: BASE_URL,
  headers: {
    'Content-Type': 'application/json',
  },
});

// Перехватчик для обработки единого формата ошибок
apiClient.interceptors.response.use(
  (response) => {
    return response;
  },
  (error: AxiosError<{ error: string; code: string; message: string }>) => {
    const defaultApiError: ApiError = {
      status: 500,
      code: 'INTERNAL_ERROR',
      message: 'Произошла непредвиденная ошибка при обращении к серверу',
    };

    if (error.response) {
      // Сервер ответил с кодом, отличным от 2xx
      const data = error.response.data;
      if (data && data.code) {
        return Promise.reject({
          status: error.response.status,
          code: data.code,
          message: data.message || defaultApiError.message,
        } as ApiError);
      }
      return Promise.reject({
         status: error.response.status,
         code: 'UNKNOWN_HTTP_ERROR',
         message: error.message,
      } as ApiError);
    } else if (error.request) {
      // Запрос был сделан, но ответ не получен
      return Promise.reject({
        ...defaultApiError,
        code: 'NETWORK_ERROR',
        message: 'Не удалось связаться с сервером. Проверьте подключение к интернету.',
      } as ApiError);
    } 

    return Promise.reject(defaultApiError);
  }
);
