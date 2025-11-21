-- Create streamer_profiles table
CREATE TABLE IF NOT EXISTS public.streamer_profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  nickname TEXT NOT NULL UNIQUE,
  street TEXT,
  city TEXT,
  postal_code TEXT,
  country_id UUID REFERENCES public.countries(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Create user_profiles table
CREATE TABLE IF NOT EXISTS public.user_profiles (
  id UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
  first_name TEXT NOT NULL,
  last_name TEXT NOT NULL,
  nickname TEXT NOT NULL UNIQUE,
  street TEXT,
  city TEXT,
  postal_code TEXT,
  country_id UUID REFERENCES public.countries(id),
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add indexes
CREATE INDEX idx_streamer_profiles_nickname ON public.streamer_profiles(nickname);
CREATE INDEX idx_streamer_profiles_country_id ON public.streamer_profiles(country_id);
CREATE INDEX idx_user_profiles_nickname ON public.user_profiles(nickname);
CREATE INDEX idx_user_profiles_country_id ON public.user_profiles(country_id);

-- Enable RLS
ALTER TABLE public.streamer_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.user_profiles ENABLE ROW LEVEL SECURITY;

