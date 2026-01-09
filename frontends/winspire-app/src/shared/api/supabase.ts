import { createClient } from '@supabase/supabase-js';

const supabaseUrl = import.meta.env.VITE_SUPABASE_URL;
const supabaseAnonKey = import.meta.env.VITE_SUPABASE_ANON_KEY;

if (!supabaseUrl || !supabaseAnonKey) {
  throw new Error('Missing Supabase environment variables. Please set VITE_SUPABASE_URL and VITE_SUPABASE_ANON_KEY');
}

export const supabase = createClient(supabaseUrl, supabaseAnonKey, {
  auth: {
    // Automatically detect session from URL hash fragment (OAuth callback)
    detectSessionInUrl: true,
    // Use PKCE flow for OAuth (more secure)
    flowType: 'pkce',
    // Store session in localStorage
    storage: window.localStorage,
    // Custom storage key for easier debugging
    storageKey: 'winspire-auth',
    // Auto-refresh tokens before they expire
    autoRefreshToken: true,
    // Persist session across page refreshes
    persistSession: true,
    // Debug logging disabled in production (enable for troubleshooting auth issues)
    debug: false,
    // Disable navigator.locks to prevent deadlock issues
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    lock: async (name: string, acquireTimeout: number, fn: () => Promise<any>) => {
      return await fn();
    },
  },
});






