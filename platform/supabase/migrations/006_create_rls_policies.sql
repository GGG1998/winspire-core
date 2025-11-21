-- Streamer profile policies
CREATE POLICY "Streamers can view own profile"
  ON public.streamer_profiles
  FOR SELECT
  USING (auth.uid() = id);

CREATE POLICY "Streamers can update own profile"
  ON public.streamer_profiles
  FOR UPDATE
  USING (auth.uid() = id);

CREATE POLICY "System can insert streamer profiles"
  ON public.streamer_profiles
  FOR INSERT
  WITH CHECK (true);

-- User profile policies
CREATE POLICY "Users can view own profile"
  ON public.user_profiles
  FOR SELECT
  USING (auth.uid() = id);

CREATE POLICY "Users can update own profile"
  ON public.user_profiles
  FOR UPDATE
  USING (auth.uid() = id);

CREATE POLICY "System can insert user profiles"
  ON public.user_profiles
  FOR INSERT
  WITH CHECK (true);

