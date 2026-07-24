import { create } from "zustand";
import { getMe, logout as apiLogout } from "@/lib/auth-api";

export interface User {
  id: string;
  email: string;
  display_name: string;
  roles: string[];
  is_admin?: boolean;
}

interface AuthState {
  user: User | null;
  isAuthenticated: boolean;
  isHydrated: boolean;
  setAuth: (accessToken: string, refreshToken: string) => void;
  setUser: (user: User | null) => void;
  logout: () => void;
  hydrate: () => void;
  loadUser: () => Promise<void>;
}

export const useAuthStore = create<AuthState>((set, get) => ({
  user: null,
  isAuthenticated: false,
  isHydrated: false,

  setAuth: (accessToken, refreshToken) => {
    localStorage.setItem("access_token", accessToken);
    localStorage.setItem("refresh_token", refreshToken);
    set({ isAuthenticated: true });
  },

  setUser: (user) => set({ user }),

  logout: () => {
    apiLogout().finally(() => {
      set({ user: null, isAuthenticated: false });
    });
  },

  hydrate: () => {
    const token = localStorage.getItem("access_token");
    set({ isAuthenticated: !!token, isHydrated: true });
  },

  loadUser: async () => {
    if (!get().isAuthenticated) return;
    try {
      const user = await getMe();
      set({ user });
    } catch {
      localStorage.removeItem("access_token");
      localStorage.removeItem("refresh_token");
      set({ user: null, isAuthenticated: false });
    }
  },
}));
