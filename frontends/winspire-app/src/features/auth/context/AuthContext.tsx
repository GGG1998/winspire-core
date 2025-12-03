import { createContext, useContext, useState, useEffect, useCallback, type ReactNode } from 'react';
import { supabase } from '../../../shared/api/supabase';
import { getCurrentUser, logout as supabaseLogout, loginUser, loginStreamer, registerUser, registerStreamer, signInWithGoogle, signInWithTwitch, signInWithDiscord } from '../api/supabaseAuth';
import type { User, LoginCredentials, UserRegisterData, StreamerRegisterData, UserProfileType } from '../types';

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (credentials: LoginCredentials, profileType: UserProfileType) => Promise<{ success: boolean; error?: string }>;
  register: (data: UserRegisterData | StreamerRegisterData) => Promise<{ success: boolean; error?: string }>;
  logout: () => Promise<void>;
  refreshUser: () => Promise<void>;
  loginWithGoogle: () => Promise<{ success: boolean; error?: string }>;
  loginWithTwitch: () => Promise<{ success: boolean; error?: string }>;
  loginWithDiscord: () => Promise<{ success: boolean; error?: string }>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
  const [user, setUser] = useState<User | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  const refreshUser = useCallback(async () => {
    try {
      const currentUser = await getCurrentUser();
      setUser(currentUser);
    } catch (error) {
      console.error('Error refreshing user:', error);
      setUser(null);
    } finally {
      setIsLoading(false);
    }
  }, []);

  useEffect(() => {
    // Listen for auth state changes - this fires INITIAL_SESSION on mount
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange(async (event, session) => {
      if (event === 'INITIAL_SESSION') {
        // This is the first event fired when the component mounts
        // It contains the session from localStorage if available
        if (session?.user) {
          await refreshUser();
        } else {
          setUser(null);
          setIsLoading(false);
        }
      } else if (event === 'TOKEN_REFRESHED') {
        // Only refresh on token refresh, not on SIGNED_IN
        // because login() already sets the user directly
        await refreshUser();
      } else if (event === 'SIGNED_OUT') {
        setUser(null);
        setIsLoading(false);
      }
    });

    return () => {
      subscription.unsubscribe();
    };
  }, [refreshUser]);

  const login = async (
    credentials: LoginCredentials,
    profileType: UserProfileType
  ): Promise<{ success: boolean; error?: string }> => {
    try {
      setIsLoading(true);
      const result = profileType === 'user' 
        ? await loginUser(credentials)
        : await loginStreamer(credentials);

      if (result.error || !result.user) {
        return {
          success: false,
          error: result.error?.message || 'Login failed',
        };
      }

      setUser(result.user);
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'An unexpected error occurred',
      };
    } finally {
      setIsLoading(false);
    }
  };

  const register = async (
    data: UserRegisterData | StreamerRegisterData
  ): Promise<{ success: boolean; error?: string }> => {
    try {
      setIsLoading(true);
      const result = data.profileType === 'user'
        ? await registerUser(data as UserRegisterData)
        : await registerStreamer(data as StreamerRegisterData);

      if (result.error || !result.user) {
        return {
          success: false,
          error: result.error?.message || 'Registration failed',
        };
      }

      setUser(result.user);
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'An unexpected error occurred',
      };
    } finally {
      setIsLoading(false);
    }
  };

  const logout = async () => {
    const currentProfileType = user?.profileType;
    setIsLoading(true);
    // Optimistically clear the user so guarded routes stop redirecting back
    setUser(null);
    try {
      const { error } = await supabaseLogout();
      if (error) {
        console.error('Error logging out:', error);
        // Rehydrate user state if Supabase sign-out fails
        await refreshUser();
        return;
      }
      
      // Redirect to appropriate login page after successful logout
      const loginRoute = currentProfileType === 'streamer' 
        ? '/auth/streamer/login'
        : '/auth/user/login';
      window.location.href = loginRoute;
    } catch (error) {
      console.error('Error logging out:', error);
      await refreshUser();
    } finally {
      setIsLoading(false);
    }
  };

  const loginWithGoogle = async (): Promise<{ success: boolean; error?: string }> => {
    try {
      const result = await signInWithGoogle();
      if (result.error) {
        return {
          success: false,
          error: result.error.message,
        };
      }
      // OAuth will redirect, so we don't need to update state here
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'An unexpected error occurred',
      };
    }
  };

  const loginWithTwitch = async (): Promise<{ success: boolean; error?: string }> => {
    try {
      const result = await signInWithTwitch();
      if (result.error) {
        return {
          success: false,
          error: result.error.message,
        };
      }
      // OAuth will redirect, so we don't need to update state here
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'An unexpected error occurred',
      };
    }
  };

  const loginWithDiscord = async (): Promise<{ success: boolean; error?: string }> => {
    try {
      const result = await signInWithDiscord();
      if (result.error) {
        return {
          success: false,
          error: result.error.message,
        };
      }
      // OAuth will redirect, so we don't need to update state here
      return { success: true };
    } catch (error) {
      return {
        success: false,
        error: error instanceof Error ? error.message : 'An unexpected error occurred',
      };
    }
  };

  return (
    <AuthContext.Provider
      value={{
        user,
        isAuthenticated: !!user,
        isLoading,
        login,
        register,
        logout,
        refreshUser,
        loginWithGoogle,
        loginWithTwitch,
        loginWithDiscord,
      }}
    >
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

