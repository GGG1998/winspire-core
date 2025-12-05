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

  const refreshUser = useCallback(async (isInitial = false) => {
    try {
      const currentUser = await getCurrentUser();
      setUser(currentUser);
      setIsLoading(false);
    } catch (error) {
      console.error('[AuthContext] Error refreshing user:', error);
      setUser(null);
      setIsLoading(false);
      
      // If this is a token refresh (not initial load) and it failed,
      // clear the session to prevent stuck state
      if (!isInitial) {
        console.warn('[AuthContext] Token refresh failed, clearing session');
        try {
          await supabase.auth.signOut();
        } catch (signOutError) {
          console.error('[AuthContext] Error signing out after failed refresh:', signOutError);
        }
      }
    }
  }, []);

  useEffect(() => {
    let didReceiveInitialSession = false;
    
    // Safety timeout: if INITIAL_SESSION doesn't fire within 10 seconds,
    // something is seriously wrong - stop loading
    const safetyTimeout = setTimeout(() => {
      if (!didReceiveInitialSession) {
        console.error('[AuthContext] INITIAL_SESSION timeout - auth state not received');
        setUser(null);
        setIsLoading(false);
      }
    }, 10000);

    // Listen for auth state changes - this fires INITIAL_SESSION on mount
    const {
      data: { subscription },
    } = supabase.auth.onAuthStateChange(async (event, session) => {
      if (event === 'INITIAL_SESSION') {
        didReceiveInitialSession = true;
        clearTimeout(safetyTimeout);
        
        // This is the first event fired when the component mounts
        // It contains the session from localStorage if available
        if (session?.user) {
          await refreshUser(true);
        } else {
          setUser(null);
          setIsLoading(false);
        }
      } else if (event === 'TOKEN_REFRESHED') {
        // Token was automatically refreshed by Supabase
        // Re-fetch user profile to ensure we have latest data
        console.log('[AuthContext] Token refreshed, updating user data');
        await refreshUser(false);
      } else if (event === 'SIGNED_OUT') {
        console.log('[AuthContext] User signed out');
        setUser(null);
        setIsLoading(false);
      } else if (event === 'USER_UPDATED') {
        // User metadata was updated
        console.log('[AuthContext] User updated');
        await refreshUser(false);
      }
    });

    return () => {
      clearTimeout(safetyTimeout);
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

