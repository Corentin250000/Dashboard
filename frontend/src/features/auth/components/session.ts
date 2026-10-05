import { ApiError, api } from '../../shared/api';

export const isAuthenticated = async (): Promise<boolean> => {
  try {
    await api.me();
    return true;
  } catch (error) {
    if (error instanceof ApiError && error.status === 401) return false;
    throw error;
  }
};
