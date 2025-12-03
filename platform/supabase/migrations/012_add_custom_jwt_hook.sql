-- Function to add custom claims to JWT tokens
-- This function enriches JWT with user nickname for use in microservices

CREATE OR REPLACE FUNCTION public.custom_access_token_hook(event jsonb)
RETURNS jsonb
LANGUAGE plpgsql
STABLE
SECURITY DEFINER
AS $$
DECLARE
  claims jsonb;
  user_id uuid;
  user_nickname text;
  user_type text;
BEGIN
  -- Extract user ID from event
  user_id := (event->>'user_id')::uuid;
  
  -- Get the existing claims
  claims := event->'claims';
  
  -- Try to get nickname from user_profiles first
  SELECT nickname INTO user_nickname
  FROM public.user_profiles
  WHERE id = user_id;
  
  -- If not found, try streamer_profiles
  IF user_nickname IS NULL THEN
    SELECT nickname INTO user_nickname
    FROM public.streamer_profiles
    WHERE id = user_id;
    
    IF user_nickname IS NOT NULL THEN
      user_type := 'streamer';
    END IF;
  ELSE
    user_type := 'user';
  END IF;
  
  -- Add nickname to claims if found
  IF user_nickname IS NOT NULL AND user_nickname != '' THEN
    claims := jsonb_set(claims, '{nickname}', to_jsonb(user_nickname));
  END IF;
  
  -- Add user_type to claims if found
  IF user_type IS NOT NULL THEN
    claims := jsonb_set(claims, '{user_type}', to_jsonb(user_type));
  END IF;
  
  -- Return the modified event
  event := jsonb_set(event, '{claims}', claims);
  
  RETURN event;
END;
$$;

-- Grant read access to tables needed by the hook
GRANT SELECT ON public.user_profiles TO supabase_auth_admin;
GRANT SELECT ON public.streamer_profiles TO supabase_auth_admin;

-- Grant execute permission to supabase_auth_admin
GRANT EXECUTE ON FUNCTION public.custom_access_token_hook TO supabase_auth_admin;

-- Revoke execute from public for security
REVOKE EXECUTE ON FUNCTION public.custom_access_token_hook FROM PUBLIC;
REVOKE EXECUTE ON FUNCTION public.custom_access_token_hook FROM anon;
REVOKE EXECUTE ON FUNCTION public.custom_access_token_hook FROM authenticated;

-- Comment
COMMENT ON FUNCTION public.custom_access_token_hook IS 'Custom access token hook that adds user nickname and type to JWT claims';

