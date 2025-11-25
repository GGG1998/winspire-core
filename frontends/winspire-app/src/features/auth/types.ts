export type UserProfileType = 'user' | 'streamer';

export interface BaseProfile {
  id: string;
  first_name: string;
  last_name: string;
  nickname: string;
  city?: string | null;
  country_id?: string | null;
  created_at: string;
  updated_at: string;
}

export interface UserProfile extends BaseProfile {
  // User-specific fields can be added here in the future
}

export interface StreamerProfile extends BaseProfile {
  // Streamer-specific fields can be added here in the future
}

export interface User {
  id: string;
  email: string;
  profileType: UserProfileType;
  profile: UserProfile | StreamerProfile;
}

export interface LoginCredentials {
  email: string;
  password: string;
}

export interface BaseRegisterData {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  nickname: string;
  city?: string;
  country_id?: string;
}

export interface UserRegisterData extends BaseRegisterData {
  profileType: 'user';
}

export interface StreamerRegisterData extends BaseRegisterData {
  profileType: 'streamer';
}

export type RegisterData = UserRegisterData | StreamerRegisterData;

// Legacy type for backward compatibility
export interface AuthResponse {
  user: User;
  token: string;
  refreshToken?: string;
}

