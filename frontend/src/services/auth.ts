const AUTH_TOKEN_KEY = 'siem_auth_token';
const USERNAME_KEY = 'siem_username';

const SESSION_TIMEOUT = 30 * 60 * 1000;
const SESSION_TIMESTAMP_KEY = 'siem_session_timestamp';

export interface AuthCredentials {
  username: string;
  password: string;
}

export const authService = {
  login: async (credentials: AuthCredentials): Promise<boolean> => {
    try {
      const token = btoa(`${credentials.username}:${credentials.password}`);
      
      const statsResponse = await fetch('/api/stats', {
        headers: {
          'Authorization': `Basic ${token}`
        }
      });

      if (!statsResponse.ok) {
        console.error('Authentication failed:', statsResponse.status);
        return false;
      }

      sessionStorage.setItem(AUTH_TOKEN_KEY, token);
      sessionStorage.setItem(USERNAME_KEY, credentials.username);
      sessionStorage.setItem(SESSION_TIMESTAMP_KEY, Date.now().toString());
      
      console.log('Authentication successful for user:', credentials.username);
      return true;
    } catch (error) {
      console.error('Login error:', error);
      return false;
    }
  },

  logout: (): void => {
    sessionStorage.removeItem(AUTH_TOKEN_KEY);
    sessionStorage.removeItem(USERNAME_KEY);
    sessionStorage.removeItem(SESSION_TIMESTAMP_KEY);
    console.log('User logged out');
  },

  getToken: (): string | null => {
    if (!authService.isSessionValid()) {
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
    const sessionValid = authService.isSessionValid();
    
    if (hasToken && !sessionValid) {
      // Сессия истекла - очищаем
      authService.logout();
      return false;
    }
    
    return hasToken && sessionValid;
  },

  isSessionValid: (): boolean => {
    const timestamp = sessionStorage.getItem(SESSION_TIMESTAMP_KEY);
    if (!timestamp) {
      return false;
    }

    const sessionAge = Date.now() - parseInt(timestamp, 10);
    return sessionAge < SESSION_TIMEOUT;
  },

  refreshSession: (): void => {
    if (authService.isAuthenticated()) {
      sessionStorage.setItem(SESSION_TIMESTAMP_KEY, Date.now().toString());
    }
  },

  getAuthHeader: (): string => {
    const token = authService.getToken();
    return token ? `Basic ${token}` : '';
  },

  getSessionTimeRemaining: (): number => {
    const timestamp = sessionStorage.getItem(SESSION_TIMESTAMP_KEY);
    if (!timestamp) {
      return 0;
    }

    const sessionAge = Date.now() - parseInt(timestamp, 10);
    const remaining = SESSION_TIMEOUT - sessionAge;
    return Math.max(0, Math.floor(remaining / 1000));
  },
};
