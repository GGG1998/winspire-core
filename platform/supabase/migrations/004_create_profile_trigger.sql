-- Function to create profile based on user metadata
CREATE OR REPLACE FUNCTION public.handle_new_user()
RETURNS TRIGGER
LANGUAGE plpgsql
SECURITY DEFINER
SET search_path = public
AS $$
DECLARE
  user_type TEXT;
  app_id TEXT;
BEGIN
  -- Extract metadata
  user_type := NEW.raw_user_meta_data->>'user_type';
  app_id := NEW.raw_user_meta_data->>'app_id';
  
  -- Validate app_id matches user_type
  IF (user_type = 'streamer' AND app_id != 'streamer-frontend-v1') THEN
    RAISE EXCEPTION 'Invalid app_id for streamer registration';
  ELSIF (user_type = 'user' AND app_id != 'user-frontend-v1') THEN
    RAISE EXCEPTION 'Invalid app_id for user registration';
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

-- Attach trigger to auth.users
DROP TRIGGER IF EXISTS on_auth_user_created ON auth.users;
CREATE TRIGGER on_auth_user_created
  AFTER INSERT ON auth.users
  FOR EACH ROW
  EXECUTE FUNCTION public.handle_new_user();

