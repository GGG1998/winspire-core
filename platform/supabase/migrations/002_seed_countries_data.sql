-- Seed common countries
INSERT INTO public.countries (name, iso_code) VALUES
  ('United States', 'US'),
  ('Poland', 'PL'),
  ('United Kingdom', 'GB'),
  ('Germany', 'DE'),
  ('France', 'FR'),
  ('Canada', 'CA'),
  ('Australia', 'AU'),
  ('Japan', 'JP'),
  ('Brazil', 'BR'),
  ('India', 'IN'),
  ('Mexico', 'MX'),
  ('Spain', 'ES'),
  ('Italy', 'IT'),
  ('Netherlands', 'NL'),
  ('Sweden', 'SE'),
  ('Norway', 'NO'),
  ('Denmark', 'DK'),
  ('Finland', 'FI'),
  ('Switzerland', 'CH'),
  ('Austria', 'AT')
ON CONFLICT (iso_code) DO NOTHING;

