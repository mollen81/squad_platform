import { create } from 'zustand';
import type { UserPrincipal } from '@/shared/types/auth';

interface AuthState {
  token: string | null;
  currentUser: UserPrincipal | null;
  isAuthenticated: boolean;
  setSession: (token: string, user: UserPrincipal) => void;
  clearSession: () => void;
}

const STORAGE_TOKEN_KEY = 'squadbit_jwt';
const STORAGE_USER_KEY = 'squadbit_principal';

const getInitialUser = (): UserPrincipal | null => {
  try {
    const raw = localStorage.getItem(STORAGE_USER_KEY);
    return raw ? (JSON.parse(raw) as UserPrincipal) : null;
  } catch {
    return null;
  }
};

export const useAuthStore = create<AuthState>((set) => ({
  token: localStorage.getItem(STORAGE_TOKEN_KEY),
  currentUser: getInitialUser(),
  isAuthenticated: Boolean(localStorage.getItem(STORAGE_TOKEN_KEY)),

  setSession: (token, user) => {
    localStorage.setItem(STORAGE_TOKEN_KEY, token);
    localStorage.setItem(STORAGE_USER_KEY, JSON.stringify(user));
    set({ token, currentUser: user, isAuthenticated: true });
  },

  clearSession: () => {
    localStorage.removeItem(STORAGE_TOKEN_KEY);
    localStorage.removeItem(STORAGE_USER_KEY);
    set({ token: null, currentUser: null, isAuthenticated: false });
  },
}));
