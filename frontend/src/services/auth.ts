// Используем sessionStorage для хранения JWT токена
const AUTH_TOKEN_KEY = 'siem_jwt_token';
const USERNAME_KEY = 'siem_username';
const TOKEN_EXPIRY_KEY = 'siem_token_expiry';

export interface AuthCredentials {
  username: string;
  password: string;
}

export interface LoginResponse {
  status: string;
  message: string;
  token: string;
  user: string;
}

export const authService = {
  login: async (credentials: AuthCredentials): Promise<boolean> => {
    try {
      const response = await fetch('/api/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(credentials),
      });

      if (!response.ok) {
        console.error('Login failed:', response.status);
        return false;
      }

      const data: LoginResponse = await response.json();

      // Сохраняем JWT токен и username
      sessionStorage.setItem(AUTH_TOKEN_KEY, data.token);
      sessionStorage.setItem(USERNAME_KEY, data.user);
      
      // Сохраняем время истечения (токен живет 24 часа)
      const expiryTime = Date.now() + (24 * 60 * 60 * 1000);
      sessionStorage.setItem(TOKEN_EXPIRY_KEY, expiryTime.toString());
      
      console.log('JWT authentication successful for user:', data.user);
      return true;
    } catch (error) {
      console.error('Login error:', error);
      return false;
    }
  },

  logout: (): void => {
    // Очищаем все данные аутентификации
    sessionStorage.removeItem(AUTH_TOKEN_KEY);
    sessionStorage.removeItem(USERNAME_KEY);
    sessionStorage.removeItem(TOKEN_EXPIRY_KEY);
    console.log('User logged out');
  },

  getToken: (): string | null => {
    // Проверяем истечение токена
    if (!authService.isTokenValid()) {
      authService.logout();
      return null;
    }
    return sessionStorage.getItem(AUTH_TOKEN_KEY);
  },

  getUsername: (): string | null => {
    return sessionStorage.getItem(USERNAME_KEY);
  },

  isAuthenticated: (): boolean => {
    const hasToken = !!sessionStorage.getItem(AUTH_TOKEN_KEY);
    const tokenValid = authService.isTokenValid();
    
    if (hasToken && !tokenValid) {
      // Токен истек - очищаем
      authService.logout();
      return false;
    }
    
    return hasToken && tokenValid;
  },

  isTokenValid: (): boolean => {
    const expiryTime = sessionStorage.getItem(TOKEN_EXPIRY_KEY);
    if (!expiryTime) {
      return false;
    }

    const now = Date.now();
    const expiry = parseInt(expiryTime, 10);
    return now < expiry;
  },

  getAuthHeader: (): string => {
    const token = authService.getToken();
    return token ? `Bearer ${token}` : '';
  },

  getTokenTimeRemaining: (): number => {
    const expiryTime = sessionStorage.getItem(TOKEN_EXPIRY_KEY);
    if (!expiryTime) {
      return 0;
    }

    const remaining = parseInt(expiryTime, 10) - Date.now();
    return Math.max(0, Math.floor(remaining / 1000));
  },
};
