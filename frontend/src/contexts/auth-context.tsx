import { createContext, useContext, useEffect, useState, type ReactNode } from 'react';
import { api, type User, type LoginResponse } from '@/lib/api';

interface AuthContextType {
  user: User | null;
  token: string | null;
  loading: boolean;
  login: (username: string, password: string) => Promise<LoginResponse>;
  verify2FALogin: (tempToken: string, code: string) => Promise<void>;
  logout: () => void;
  refreshUser: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [token, setTokenState] = useState<string | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const storedToken = api.getToken();
    if (storedToken) {
      setTokenState(storedToken);
      api.getProfile()
        .then(setUser)
        .catch(() => {
          api.setToken(null);
          setTokenState(null);
        })
        .finally(() => setLoading(false));
    } else {
      // Use microtask to avoid synchronous setState in effect
      queueMicrotask(() => setLoading(false));
    }
  }, []);

  const login = async (username: string, password: string): Promise<LoginResponse> => {
    const result = await api.login(username, password);
    if (result.user) {
      setUser(result.user);
      setTokenState(api.getToken());
    }
    return result;
  };

  const verify2FALogin = async (tempToken: string, code: string) => {
    const result = await api.verify2FALogin(tempToken, code);
    setUser(result.user);
    setTokenState(api.getToken());
  };

  const logout = () => {
    api.logout();
    setUser(null);
    setTokenState(null);
  };

  const refreshUser = async () => {
    const user = await api.getProfile();
    setUser(user);
  };

  return (
    <AuthContext.Provider value={{ user, token, loading, login, verify2FALogin, logout, refreshUser }}>
      {children}
    </AuthContext.Provider>
  );
}

// eslint-disable-next-line react-refresh/only-export-components
export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
