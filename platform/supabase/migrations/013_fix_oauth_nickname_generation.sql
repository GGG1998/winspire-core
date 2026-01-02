-- Fix OAuth nickname generation to avoid UNIQUE constraint violations
-- Problem: OAuth users don't have 'nickname' in metadata, trigger defaults to empty string ''
-- This causes UNIQUE constraint violation when second OAuth user tries to sign up
--
-- Additional fix: When using Supabase Admin API, the app_metadata (with provider) is set
-- AFTER the initial INSERT, so the trigger may not see the provider. We add an UPDATE trigger
-- to handle this case by checking if a profile needs to be moved from user to streamer.

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
  base_nickname TEXT;
  final_nickname TEXT;
  suffix_counter INT := 0;
  first_name_val TEXT;
  last_name_val TEXT;
  full_name_val TEXT;
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

  -- Generate nickname with fallback chain (prioritize explicit nickname, then OAuth usernames, then email)
  base_nickname := COALESCE(
    NULLIF(TRIM(NEW.raw_user_meta_data->>'nickname'), ''),
    NULLIF(TRIM(NEW.raw_user_meta_data->>'preferred_username'), ''),
    NULLIF(TRIM(NEW.raw_user_meta_data->>'user_name'), ''),
    NULLIF(TRIM(split_part(NEW.email, '@', 1)), ''),
    'user_' || substr(NEW.id::text, 1, 8)
  );

  -- Extract first and last name from OAuth metadata
  full_name_val := COALESCE(
    NULLIF(TRIM(NEW.raw_user_meta_data->>'full_name'), ''),
    NULLIF(TRIM(NEW.raw_user_meta_data->>'name'), ''),
    ''
  );

  first_name_val := COALESCE(
    NULLIF(TRIM(NEW.raw_user_meta_data->>'first_name'), ''),
    NULLIF(TRIM(split_part(full_name_val, ' ', 1)), ''),
    ''
  );

  last_name_val := COALESCE(
    NULLIF(TRIM(NEW.raw_user_meta_data->>'last_name'), ''),
    CASE
      WHEN position(' ' in full_name_val) > 0
      THEN TRIM(substring(full_name_val from position(' ' in full_name_val) + 1))
      ELSE ''
    END,
    ''
  );

  -- Ensure nickname is unique by appending suffix if needed
  final_nickname := base_nickname;
  LOOP
    BEGIN
      IF user_type = 'streamer' THEN
        INSERT INTO public.streamer_profiles (id, first_name, last_name, nickname)
        VALUES (NEW.id, first_name_val, last_name_val, final_nickname);
      ELSIF user_type = 'user' THEN
        INSERT INTO public.user_profiles (id, first_name, last_name, nickname)
        VALUES (NEW.id, first_name_val, last_name_val, final_nickname);
      ELSE
        RAISE EXCEPTION 'Invalid user_type: %. Must be streamer or user', user_type;
      END IF;
      EXIT; -- Success, exit loop
    EXCEPTION
      WHEN unique_violation THEN
        suffix_counter := suffix_counter + 1;
        IF suffix_counter > 100 THEN
          RAISE EXCEPTION 'Could not generate unique nickname after 100 attempts';
        END IF;
        final_nickname := base_nickname || '_' || suffix_counter;
    END;
  END LOOP;

  RETURN NEW;
EXCEPTION
  WHEN OTHERS THEN
    RAISE WARNING 'Profile creation failed for user %: %', NEW.id, SQLERRM;
    RAISE;
END;
$$;

-- Fix any existing profiles with empty nicknames
UPDATE public.user_profiles
SET nickname = 'user_' || substr(id::text, 1, 8)
WHERE nickname = '' OR nickname IS NULL;

UPDATE public.streamer_profiles
SET nickname = 'streamer_' || substr(id::text, 1, 8)
WHERE nickname = '' OR nickname IS NULL;

-- Handle case where Supabase Admin API sets app_metadata AFTER user creation
-- This trigger moves profile from user_profiles to streamer_profiles if needed
CREATE OR REPLACE FUNCTION public.handle_user_update()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  old_provider TEXT;
  new_provider TEXT;
  expected_type TEXT;
  has_user_profile BOOLEAN;
  has_streamer_profile BOOLEAN;
  profile_data RECORD;
BEGIN
  -- Only care about app_metadata changes
  old_provider := OLD.raw_app_meta_data->>'provider';
  new_provider := NEW.raw_app_meta_data->>'provider';

  -- If provider didn't change, nothing to do
  IF old_provider IS NOT DISTINCT FROM new_provider THEN
    RETURN NEW;
  END IF;

  -- Determine expected profile type based on new provider
  IF new_provider IN ('twitch', 'discord') THEN
    expected_type := 'streamer';
  ELSIF new_provider = 'google' THEN
    expected_type := 'user';
  ELSE
    expected_type := 'user';
  END IF;

  -- Check existing profiles
  SELECT EXISTS(SELECT 1 FROM public.user_profiles WHERE id = NEW.id) INTO has_user_profile;
  SELECT EXISTS(SELECT 1 FROM public.streamer_profiles WHERE id = NEW.id) INTO has_streamer_profile;

  -- If user should be a streamer but has user profile, migrate
  IF expected_type = 'streamer' AND has_user_profile AND NOT has_streamer_profile THEN
    SELECT * INTO profile_data FROM public.user_profiles WHERE id = NEW.id;

    INSERT INTO public.streamer_profiles (id, first_name, last_name, nickname, street, city, postal_code, country_id, created_at, updated_at)
    VALUES (profile_data.id, profile_data.first_name, profile_data.last_name, profile_data.nickname,
            profile_data.street, profile_data.city, profile_data.postal_code, profile_data.country_id,
            profile_data.created_at, profile_data.updated_at);

    DELETE FROM public.user_profiles WHERE id = NEW.id;
  END IF;

  -- If user should be a user but has streamer profile, migrate
  IF expected_type = 'user' AND has_streamer_profile AND NOT has_user_profile THEN
    SELECT * INTO profile_data FROM public.streamer_profiles WHERE id = NEW.id;

    INSERT INTO public.user_profiles (id, first_name, last_name, nickname, street, city, postal_code, country_id, created_at, updated_at)
    VALUES (profile_data.id, profile_data.first_name, profile_data.last_name, profile_data.nickname,
            profile_data.street, profile_data.city, profile_data.postal_code, profile_data.country_id,
            profile_data.created_at, profile_data.updated_at);

    DELETE FROM public.streamer_profiles WHERE id = NEW.id;
  END IF;

  RETURN NEW;
END;
$$;

-- Create the UPDATE trigger
DROP TRIGGER IF EXISTS on_auth_user_updated ON auth.users;
CREATE TRIGGER on_auth_user_updated
  AFTER UPDATE ON auth.users
  FOR EACH ROW
  EXECUTE FUNCTION public.handle_user_update();
