const AUTH_KEY = 'siem_auth';

export interface AuthCredentials {
  username: string;
  password: string;
}

export const authService = {
  login: (credentials: AuthCredentials): boolean => {
    const token = btoa(`${credentials.username}:${credentials.password}`);
    localStorage.setItem(AUTH_KEY, token);
    return true;
  },

  logout: (): void => {
    localStorage.removeItem(AUTH_KEY);
  },

  getToken: (): string | null => {
    return localStorage.getItem(AUTH_KEY);
  },

  isAuthenticated: (): boolean => {
    return !!localStorage.getItem(AUTH_KEY);
  },

  getAuthHeader: (): string => {
    const token = localStorage.getItem(AUTH_KEY);
    return token ? `Basic ${token}` : '';
  },
};
