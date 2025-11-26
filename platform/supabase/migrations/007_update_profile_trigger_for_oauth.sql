-- Update the profile trigger to handle OAuth users without user_type metadata
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  user_type TEXT;
  app_id TEXT;
  provider TEXT;
BEGIN
  -- Extract metadata
  user_type := NEW.raw_user_meta_data->>'user_type';
  app_id := NEW.raw_user_meta_data->>'app_id';
  provider := NEW.raw_app_meta_data->>'provider';
  
  -- For OAuth users without explicit user_type, determine from provider
  IF user_type IS NULL AND provider IS NOT NULL THEN
    IF provider IN ('twitch', 'discord') THEN
      user_type := 'streamer';
    ELSIF provider = 'google' THEN
      user_type := 'user';
    ELSE
      -- Default to user for other providers
      user_type := 'user';
    END IF;
  END IF;
  
  -- Validate app_id matches user_type (only for non-OAuth users)
  IF app_id IS NOT NULL THEN
    IF (user_type = 'streamer' AND app_id != 'streamer-frontend-v1') THEN
      RAISE EXCEPTION 'Invalid app_id for streamer registration';
    ELSIF (user_type = 'user' AND app_id != 'user-frontend-v1') THEN
      RAISE EXCEPTION 'Invalid app_id for user registration';
    END IF;
  END IF;
  
  -- Create appropriate profile
  IF user_type = 'streamer' THEN
    INSERT INTO public.streamer_profiles (
      id, first_name, last_name, nickname
    ) VALUES (
      NEW.id,
      COALESCE(NEW.raw_user_meta_data->>'first_name', ''),
      COALESCE(NEW.raw_user_meta_data->>'last_name', ''),
      COALESCE(NEW.raw_user_meta_data->>'nickname', '')
    );
  ELSIF user_type = 'user' THEN
    INSERT INTO public.user_profiles (
      id, first_name, last_name, nickname
    ) VALUES (
      NEW.id,
      COALESCE(NEW.raw_user_meta_data->>'first_name', ''),
      COALESCE(NEW.raw_user_meta_data->>'last_name', ''),
      COALESCE(NEW.raw_user_meta_data->>'nickname', '')
    );
  ELSE
    RAISE EXCEPTION 'Invalid user_type: %. Must be streamer or user', user_type;
  END IF;
  
  RETURN NEW;
EXCEPTION
  WHEN OTHERS THEN
    RAISE WARNING 'Profile creation failed for user %: %', NEW.id, SQLERRM;
    RAISE;
END;
$$;

