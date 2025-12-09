import { supabase } from '../../../shared/api/supabase';
import type { User, UserProfileType, UserProfile, StreamerProfile, UserRegisterData, StreamerRegisterData, LoginCredentials } from '../types';

export interface AuthError {
  message: string;
  code?: string;
}

async function fetchUserProfile(userId: string, profileType: UserProfileType): Promise<UserProfile | StreamerProfile | null> {
  console.log('fetchUserProfile', userId, profileType);
  try {
    if (profileType === 'user') {
      const { data, error } = await supabase
        .from('user_profiles')
        .select('*')
        .eq('id', userId)
        .single();

      if (error) {
        console.error('Error fetching user profile:', error);
        return null;
      }

      return data as UserProfile;
    } else {
      const { data, error } = await supabase
        .from('streamer_profiles')
        .select('*')
        .eq('id', userId)
        .single();

      if (error) {
        console.error('Error fetching streamer profile:', error);
        return null;
      }

      return data as StreamerProfile;
    }
  } catch (error) {
    console.error('Error fetching profile:', error);
    return null;
  }
}

async function createUserProfile(userId: string, data: UserRegisterData): Promise<UserProfile | null> {
  try {
    // Profile is created automatically by database trigger, so we update it with additional fields
    const updateData: any = {};
    
    // Only add fields if they are provided
    if (data.city !== undefined) {
      updateData.city = data.city || null;
    }
    if (data.country_id !== undefined) {
      updateData.country_id = data.country_id || null;
    }
    
    // If there's nothing to update, just fetch the profile
    if (Object.keys(updateData).length === 0) {
      const { data: existingProfile, error: fetchError } = await supabase
        .from('user_profiles')
        .select('*')
        .eq('id', userId)
        .single();
      
      if (fetchError || !existingProfile) {
        console.error('Error fetching user profile:', fetchError);
        return null;
      }
      return existingProfile as UserProfile;
    }

    const { data: profile, error } = await supabase
      .from('user_profiles')
      .update(updateData)
      .eq('id', userId)
      .select()
      .single();

    if (error) {
      console.error('Error updating user profile:', error);
      // If update fails, try to fetch the existing profile (created by trigger)
      const { data: existingProfile, error: fetchError } = await supabase
        .from('user_profiles')
        .select('*')
        .eq('id', userId)
        .single();
      
      if (fetchError || !existingProfile) {
        console.error('Error fetching user profile:', fetchError);
        return null;
      }
      return existingProfile as UserProfile;
    }

    return profile as UserProfile;
  } catch (error) {
    console.error('Error updating user profile:', error);
    return null;
  }
}

async function createStreamerProfile(userId: string, data: StreamerRegisterData): Promise<StreamerProfile | null> {
  try {
    // Profile is created automatically by database trigger, so we update it with additional fields
    const updateData: any = {};
    
    // Only add fields if they are provided
    if (data.city !== undefined) {
      updateData.city = data.city || null;
    }
    if (data.country_id !== undefined) {
      updateData.country_id = data.country_id || null;
    }
    
    // If there's nothing to update, just fetch the profile
    if (Object.keys(updateData).length === 0) {
      const { data: existingProfile, error: fetchError } = await supabase
        .from('streamer_profiles')
        .select('*')
        .eq('id', userId)
        .single();
      
      if (fetchError || !existingProfile) {
        console.error('Error fetching streamer profile:', fetchError);
        return null;
      }
      return existingProfile as StreamerProfile;
    }

    const { data: profile, error } = await supabase
      .from('streamer_profiles')
      .update(updateData)
      .eq('id', userId)
      .select()
      .single();

    if (error) {
      console.error('Error updating streamer profile:', error);
      // If update fails, try to fetch the existing profile (created by trigger)
      const { data: existingProfile, error: fetchError } = await supabase
        .from('streamer_profiles')
        .select('*')
        .eq('id', userId)
        .single();
      
      if (fetchError || !existingProfile) {
        console.error('Error fetching streamer profile:', fetchError);
        return null;
      }
      return existingProfile as StreamerProfile;
    }

    return profile as StreamerProfile;
  } catch (error) {
    console.error('Error updating streamer profile:', error);
    return null;
  }
}

export async function loginUser(credentials: LoginCredentials): Promise<{ user: User | null; error: AuthError | null }> {
  try {
    const { data: authData, error: authError } = await supabase.auth.signInWithPassword({
      email: credentials.email,
      password: credentials.password,
    });

    if (authError || !authData.user) {
      return {
        user: null,
        error: {
          message: authError?.message || 'Login failed',
          code: authError?.status?.toString(),
        },
      };
    }

    // Try to fetch user profile first
    let profile = await fetchUserProfile(authData.user.id, 'user');
    
    // If user profile doesn't exist, try streamer profile
    if (!profile) {
      profile = await fetchUserProfile(authData.user.id, 'streamer');
      if (profile) {
        return {
          user: {
            id: authData.user.id,
            email: authData.user.email!,
            profileType: 'streamer',
            profile: profile as StreamerProfile,
          },
          error: null,
        };
      }
    } else {
      return {
        user: {
          id: authData.user.id,
          email: authData.user.email!,
          profileType: 'user',
          profile: profile as UserProfile,
        },
        error: null,
      };
    }

    // If no profile found, return error
    return {
      user: null,
      error: {
        message: 'User profile not found',
        code: 'PROFILE_NOT_FOUND',
      },
    };
  } catch (error) {
    return {
      user: null,
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

export async function loginStreamer(credentials: LoginCredentials): Promise<{ user: User | null; error: AuthError | null }> {
  try {
    const { data: authData, error: authError } = await supabase.auth.signInWithPassword({
      email: credentials.email,
      password: credentials.password,
    });

    if (authError || !authData.user) {
      return {
        user: null,
        error: {
          message: authError?.message || 'Login failed',
          code: authError?.status?.toString(),
        },
      };
    }

    const profile = await fetchUserProfile(authData.user.id, 'streamer');

    if (!profile) {
      return {
        user: null,
        error: {
          message: 'Streamer profile not found',
          code: 'PROFILE_NOT_FOUND',
        },
      };
    }

    return {
      user: {
        id: authData.user.id,
        email: authData.user.email!,
        profileType: 'streamer',
        profile: profile as StreamerProfile,
      },
      error: null,
    };
  } catch (error) {
    return {
      user: null,
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

export async function registerUser(data: UserRegisterData): Promise<{ user: User | null; error: AuthError | null }> {
  try {
    const { data: authData, error: authError } = await supabase.auth.signUp({
      email: data.email,
      password: data.password,
      options: {
        data: {
          user_type: 'user',
          app_id: 'user-frontend-v1',
          first_name: data.first_name,
          last_name: data.last_name,
          nickname: data.nickname,
        },
      },
    });

    if (authError || !authData.user) {
      return {
        user: null,
        error: {
          message: authError?.message || 'Registration failed',
          code: authError?.status?.toString(),
        },
      };
    }

    const profile = await createUserProfile(authData.user.id, data);

    if (!profile) {
      // Note: Cannot delete user from client-side, this should be handled by a backend function
      // or database trigger. For now, we'll just return an error.
      return {
        user: null,
        error: {
          message: 'Failed to create user profile. Please contact support.',
          code: 'PROFILE_CREATION_FAILED',
        },
      };
    }

    return {
      user: {
        id: authData.user.id,
        email: authData.user.email!,
        profileType: 'user',
        profile: profile,
      },
      error: null,
    };
  } catch (error) {
    return {
      user: null,
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
    }
}

export async function registerStreamer(data: StreamerRegisterData): Promise<{ user: User | null; error: AuthError | null }> {
  try {
    const { data: authData, error: authError } = await supabase.auth.signUp({
      email: data.email,
      password: data.password,
      options: {
        data: {
          user_type: 'streamer',
          app_id: 'streamer-frontend-v1',
          first_name: data.first_name,
          last_name: data.last_name,
          nickname: data.nickname,
        },
      },
    });

    if (authError || !authData.user) {
      return {
        user: null,
        error: {
          message: authError?.message || 'Registration failed',
          code: authError?.status?.toString(),
        },
      };
    }

    const profile = await createStreamerProfile(authData.user.id, data);

    if (!profile) {
      // Note: Cannot delete user from client-side, this should be handled by a backend function
      // or database trigger. For now, we'll just return an error.
      return {
        user: null,
        error: {
          message: 'Failed to create streamer profile. Please contact support.',
          code: 'PROFILE_CREATION_FAILED',
        },
      };
    }

    return {
      user: {
        id: authData.user.id,
        email: authData.user.email!,
        profileType: 'streamer',
        profile: profile,
      },
      error: null,
    };
  } catch (error) {
    return {
      user: null,
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

export async function getCurrentUser(): Promise<User | null> {
  try {
    console.log('[getCurrentUser] Fetching current user session');
    
    // Use getSession() first - it's synchronous from localStorage cache
    // This prevents the hanging issue with getUser() which makes a network request
    const { data: { session }, error: sessionError } = await supabase.auth.getSession();
    
    if (sessionError) {
      console.error('[getCurrentUser] Session error:', sessionError);
      return null;
    }
    
    if (!session?.user) {
      console.log('[getCurrentUser] No active session found');
      return null;
    }

    const authUser = session.user;
    console.log('[getCurrentUser] Active session found for user:', authUser.id);

    // Try user profile first
    let profile = await fetchUserProfile(authUser.id, 'user');
    let profileType: UserProfileType = 'user';

    if (!profile) {
      console.log('[getCurrentUser] User profile not found, trying streamer profile');
      // Try streamer profile
      profile = await fetchUserProfile(authUser.id, 'streamer');
      profileType = 'streamer';
    }

    if (!profile) {
      console.error('[getCurrentUser] No profile found for user:', authUser.id);
      return null;
    }

    console.log('[getCurrentUser] User profile loaded:', { profileType, userId: authUser.id });
    return {
      id: authUser.id,
      email: authUser.email!,
      profileType,
      profile,
    };
  } catch (error) {
    console.error('[getCurrentUser] Error getting current user:', error);
    return null;
  }
}

export async function logout(): Promise<{ error: AuthError | null }> {
  try {
    const { error } = await supabase.auth.signOut();
    return {
      error: error ? {
        message: error.message,
        code: error.status?.toString(),
      } : null,
    };
  } catch (error) {
    return {
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

// OAuth Sign-in Functions
export async function signInWithGoogle(): Promise<{ error: AuthError | null }> {
  try {
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'google',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
        queryParams: {
          access_type: 'offline',
          prompt: 'consent',
        },
      },
    });

    if (error) {
      return {
        error: {
          message: error.message,
          code: error.status?.toString(),
        },
      };
    }

    return { error: null };
  } catch (error) {
    return {
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

export async function signInWithTwitch(): Promise<{ error: AuthError | null }> {
  try {
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'twitch',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    });

    if (error) {
      return {
        error: {
          message: error.message,
          code: error.status?.toString(),
        },
      };
    }

    return { error: null };
  } catch (error) {
    return {
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

export async function signInWithDiscord(): Promise<{ error: AuthError | null }> {
  try {
    const { error } = await supabase.auth.signInWithOAuth({
      provider: 'discord',
      options: {
        redirectTo: `${window.location.origin}/auth/callback`,
      },
    });

    if (error) {
      return {
        error: {
          message: error.message,
          code: error.status?.toString(),
        },
      };
    }

    return { error: null };
  } catch (error) {
    return {
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

// Handle OAuth callback and create/fetch profile
export async function handleOAuthCallback(): Promise<{ user: User | null; error: AuthError | null }> {
  try {
    console.log('[handleOAuthCallback] Starting OAuth callback processing');
    
    const { data: { user: authUser }, error: authError } = await supabase.auth.getUser();

    if (authError || !authUser) {
      console.error('[handleOAuthCallback] Failed to get authenticated user:', authError);
      return {
        user: null,
        error: {
          message: authError?.message || 'Failed to authenticate',
          code: authError?.status?.toString(),
        },
      };
    }

    console.log('[handleOAuthCallback] Authenticated user found:', {
      id: authUser.id,
      email: authUser.email,
      provider: authUser.app_metadata.provider,
    });

    // Determine profile type based on OAuth provider
    const provider = authUser.app_metadata.provider;
    let profileType: UserProfileType;
    
    if (provider === 'twitch' || provider === 'discord') {
      profileType = 'streamer';
      console.log('[handleOAuthCallback] Provider is', provider, '- using streamer profile');
    } else if (provider === 'google') {
      profileType = 'user';
      console.log('[handleOAuthCallback] Provider is Google - using user profile');
    } else {
      // Default to user for unknown providers
      profileType = 'user';
      console.log('[handleOAuthCallback] Unknown provider', provider, '- defaulting to user profile');
    }

    // Try to fetch existing profile
    console.log('[handleOAuthCallback] Checking for existing profile...');
    let profile = await fetchUserProfile(authUser.id, profileType);

    // If profile doesn't exist, create it with OAuth data
    if (!profile) {
      console.log('[handleOAuthCallback] Profile not found, creating new profile');
      const oauthMetadata = authUser.user_metadata;
      console.log('[handleOAuthCallback] OAuth metadata:', {
        full_name: oauthMetadata.full_name,
        name: oauthMetadata.name,
        preferred_username: oauthMetadata.preferred_username,
        user_name: oauthMetadata.user_name,
      });
      
      const firstName = oauthMetadata.full_name?.split(' ')[0] || oauthMetadata.name?.split(' ')[0] || 'User';
      const lastName = oauthMetadata.full_name?.split(' ').slice(1).join(' ') || oauthMetadata.name?.split(' ').slice(1).join(' ') || '';
      const nickname = oauthMetadata.preferred_username || oauthMetadata.user_name || oauthMetadata.name || authUser.email?.split('@')[0] || 'user';

      console.log('[handleOAuthCallback] Creating profile with:', { firstName, lastName, nickname, profileType });

      if (profileType === 'streamer') {
        profile = await createStreamerProfile(authUser.id, {
          email: authUser.email!,
          password: '', // Not used for OAuth
          first_name: firstName,
          last_name: lastName,
          nickname: nickname,
          profileType: 'streamer',
        });
      } else {
        profile = await createUserProfile(authUser.id, {
          email: authUser.email!,
          password: '', // Not used for OAuth
          first_name: firstName,
          last_name: lastName,
          nickname: nickname,
          profileType: 'user',
        });
      }

      if (!profile) {
        console.error('[handleOAuthCallback] Failed to create profile');
        return {
          user: null,
          error: {
            message: 'Failed to create user profile',
            code: 'PROFILE_CREATION_FAILED',
          },
        };
      }
      
      console.log('[handleOAuthCallback] Profile created successfully');
    } else {
      console.log('[handleOAuthCallback] Existing profile found');
    }

    console.log('[handleOAuthCallback] OAuth callback completed successfully');
    return {
      user: {
        id: authUser.id,
        email: authUser.email!,
        profileType,
        profile,
      },
      error: null,
    };
  } catch (error) {
    console.error('[handleOAuthCallback] Unexpected error:', error);
    return {
      user: null,
      error: {
        message: error instanceof Error ? error.message : 'An unexpected error occurred',
      },
    };
  }
}

