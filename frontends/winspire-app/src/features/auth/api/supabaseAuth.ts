import { supabase } from '../../../shared/api/supabase';
import type { User, UserProfileType, UserProfile, StreamerProfile, UserRegisterData, StreamerRegisterData, LoginCredentials } from '../types';

export interface AuthError {
  message: string;
  code?: string;
}

async function fetchUserProfile(userId: string, profileType: UserProfileType): Promise<UserProfile | StreamerProfile | null> {
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
    const { data: profile, error } = await supabase
      .from('user_profiles')
      .insert({
        id: userId,
        first_name: data.first_name,
        last_name: data.last_name,
        nickname: data.nickname,
        street: data.street || null,
        city: data.city || null,
        postal_code: data.postal_code || null,
        country_id: data.country_id || null,
      })
      .select()
      .single();

    if (error) {
      console.error('Error creating user profile:', error);
      return null;
    }

    return profile as UserProfile;
  } catch (error) {
    console.error('Error creating user profile:', error);
    return null;
  }
}

async function createStreamerProfile(userId: string, data: StreamerRegisterData): Promise<StreamerProfile | null> {
  try {
    const { data: profile, error } = await supabase
      .from('streamer_profiles')
      .insert({
        id: userId,
        first_name: data.first_name,
        last_name: data.last_name,
        nickname: data.nickname,
        street: data.street || null,
        city: data.city || null,
        postal_code: data.postal_code || null,
        country_id: data.country_id || null,
      })
      .select()
      .single();

    if (error) {
      console.error('Error creating streamer profile:', error);
      return null;
    }

    return profile as StreamerProfile;
  } catch (error) {
    console.error('Error creating streamer profile:', error);
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
    const { data: { user: authUser }, error } = await supabase.auth.getUser();

    if (error || !authUser) {
      return null;
    }

    // Try user profile first
    let profile = await fetchUserProfile(authUser.id, 'user');
    let profileType: UserProfileType = 'user';

    if (!profile) {
      // Try streamer profile
      profile = await fetchUserProfile(authUser.id, 'streamer');
      profileType = 'streamer';
    }

    if (!profile) {
      return null;
    }

    return {
      id: authUser.id,
      email: authUser.email!,
      profileType,
      profile,
    };
  } catch (error) {
    console.error('Error getting current user:', error);
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

