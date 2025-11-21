-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION public.update_updated_at()
RETURNS TRIGGER
LANGUAGE plpgsql
AS $$
BEGIN
  NEW.updated_at = NOW();
  RETURN NEW;
END;
$$;

-- Attach to streamer_profiles
DROP TRIGGER IF EXISTS update_streamer_profiles_updated_at ON public.streamer_profiles;
CREATE TRIGGER update_streamer_profiles_updated_at
  BEFORE UPDATE ON public.streamer_profiles
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at();

-- Attach to user_profiles
DROP TRIGGER IF EXISTS update_user_profiles_updated_at ON public.user_profiles;
CREATE TRIGGER update_user_profiles_updated_at
  BEFORE UPDATE ON public.user_profiles
  FOR EACH ROW
  EXECUTE FUNCTION public.update_updated_at();

