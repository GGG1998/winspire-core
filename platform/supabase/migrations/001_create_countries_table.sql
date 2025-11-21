-- Ensure uuid generation support is available locally
CREATE EXTENSION IF NOT EXISTS "uuid-ossp" WITH SCHEMA extensions;

-- Create countries lookup table
CREATE TABLE IF NOT EXISTS public.countries (
  id UUID PRIMARY KEY DEFAULT extensions.uuid_generate_v4(),
  name TEXT NOT NULL,
  iso_code VARCHAR(2) NOT NULL UNIQUE,
  created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Add indexes
CREATE INDEX idx_countries_iso_code ON public.countries(iso_code);
CREATE INDEX idx_countries_name ON public.countries(name);

-- Enable RLS
ALTER TABLE public.countries ENABLE ROW LEVEL SECURITY;

-- Public read policy
CREATE POLICY "Anyone can view countries"
  ON public.countries
  FOR SELECT
  TO authenticated
  USING (true);

