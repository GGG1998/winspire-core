import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { handleOAuthCallback } from '../api/supabaseAuth';
import { useAuth } from '../hooks/useAuth';
import { supabase } from '../../../shared/api/supabase';
import { AuthLayout } from '../../../shared/components/ui/auth-layout';
import { Heading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';

/**
 * Returns the appropriate login URL based on OAuth provider or profile type.
 * - Twitch/Discord providers → streamer login
 * - Google/unknown providers → user login
 */
function getLoginUrl(providerOrProfileType: string | undefined): string {
  if (providerOrProfileType === 'twitch' || providerOrProfileType === 'discord' || providerOrProfileType === 'streamer') {
    return '/auth/streamer/login';
  }
  return '/auth/user/login';
}

export function OAuthCallbackPage() {
  const navigate = useNavigate();
  const { refreshUser } = useAuth();
  const [error, setError] = useState<string | null>(null);
  const handledRef = useRef(false);

  useEffect(() => {
    // Prevent double execution in React StrictMode
    if (handledRef.current) {
      return;
    }

    let timeoutId: ReturnType<typeof setTimeout>;
    let unsubscribe: (() => void) | null = null;

    const handleCallback = async () => {
      try {
        // Set a timeout in case Supabase never fires the auth event
        timeoutId = setTimeout(async () => {
          console.error('[OAuthCallbackPage] Timeout waiting for session establishment');
          setError('Authentication timeout. Please try again.');
          // Try to determine the login URL from any existing session
          const { data: { session } } = await supabase.auth.getSession();
          const provider = session?.user?.app_metadata?.provider;
          setTimeout(() => {
            navigate(getLoginUrl(provider));
          }, 3000);
        }, 15000); // 15 second timeout

        // Listen for auth state changes from Supabase
        // Set up listener BEFORE exchanging code so we don't miss the SIGNED_IN event
        const { data: { subscription } } = supabase.auth.onAuthStateChange(async (event, session) => {
          // Only process SIGNED_IN event (when OAuth callback is processed)
          if (event === 'SIGNED_IN' && session && !handledRef.current) {
            handledRef.current = true;
            clearTimeout(timeoutId);
            
            try {
              const result = await handleOAuthCallback();

              if (result.error || !result.user) {
                console.error('[OAuthCallbackPage] Profile processing failed:', result.error);
                setError(result.error?.message || 'Authentication failed');
                const loginUrl = getLoginUrl(result.user?.profileType || session?.user?.app_metadata?.provider);
                setTimeout(() => {
                  navigate(loginUrl);
                }, 3000);
                return;
              }

              // Refresh the auth context with the new user
              await refreshUser();

              // Check if user has completed their profile (nickname is required)
              if (!result.user.profile.nickname || result.user.profile.nickname.trim() === '') {
                navigate('/auth/complete-profile');
                return;
              }

              // Profile complete, redirect to home page
              navigate('/');
            } catch (err) {
              console.error('[OAuthCallbackPage] Error during profile processing:', err);
              setError(err instanceof Error ? err.message : 'An unexpected error occurred');
              const loginUrl = getLoginUrl(session?.user?.app_metadata?.provider);
              setTimeout(() => {
                navigate(loginUrl);
              }, 3000);
            }
          }
        });

        unsubscribe = subscription.unsubscribe;

        // PKCE flow: Extract auth code from URL query parameter and exchange it for a session
        const urlParams = new URLSearchParams(window.location.search);
        const code = urlParams.get('code');

        if (code) {
          // Clean up URL to prevent issues on refresh (code can only be used once)
          window.history.replaceState({}, '', window.location.pathname);

          const { error: exchangeError } = await supabase.auth.exchangeCodeForSession(code);
          if (exchangeError) {
            console.error('[OAuthCallbackPage] Failed to exchange code for session:', exchangeError);
            clearTimeout(timeoutId);
            setError('Authentication failed. Please try again.');
            // When exchange fails, we don't have session data to determine provider
            // Default to user login (most common case for Google OAuth)
            setTimeout(() => {
              navigate(getLoginUrl(undefined));
            }, 3000);
            return;
          }
          // After successful exchange, the onAuthStateChange listener will fire SIGNED_IN
          // and handle the rest of the flow, so we can return here
          return;
        }

        // Check if session already exists (in case event already fired, e.g., hash fragment flow)
        const { data: { session } } = await supabase.auth.getSession();
        if (session && !handledRef.current) {
          handledRef.current = true;
          clearTimeout(timeoutId);
          
          const result = await handleOAuthCallback();

          if (result.error || !result.user) {
            console.error('[OAuthCallbackPage] Profile processing failed:', result.error);
            setError(result.error?.message || 'Authentication failed');
            const loginUrl = getLoginUrl(result.user?.profileType || session?.user?.app_metadata?.provider);
            setTimeout(() => {
              navigate(loginUrl);
            }, 3000);
            return;
          }

          await refreshUser();

          if (!result.user.profile.nickname || result.user.profile.nickname.trim() === '') {
            navigate('/auth/complete-profile');
            return;
          }

          navigate('/');
        }
      } catch (err) {
        console.error('[OAuthCallbackPage] Unexpected error:', err);
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
        // Try to get session to determine the right login URL
        const { data: { session: currentSession } } = await supabase.auth.getSession();
        const loginUrl = getLoginUrl(currentSession?.user?.app_metadata?.provider);
        setTimeout(() => {
          navigate(loginUrl);
        }, 3000);
      }
    };

    handleCallback();

    // Cleanup
    return () => {
      if (timeoutId) {
        clearTimeout(timeoutId);
      }
      if (unsubscribe) {
        unsubscribe();
      }
    };
  }, [navigate, refreshUser]);

  if (error) {
    return (
      <AuthLayout>
        <div className="space-y-6 text-center">
          <Heading>Authentication Error</Heading>
          <div className="rounded-lg bg-red-50 p-4 ring-1 ring-red-200 dark:bg-red-900/20 dark:ring-red-800">
            <Text className="text-red-800 dark:text-red-200">{error}</Text>
          </div>
          <Text className="text-sm text-zinc-500 dark:text-zinc-400">
            Redirecting to login page...
          </Text>
        </div>
      </AuthLayout>
    );
  }

  return (
    <AuthLayout>
      <div className="space-y-6 text-center">
        <Heading>Completing sign in...</Heading>
        <div className="flex justify-center">
          <svg
            className="h-12 w-12 animate-spin text-zinc-950 dark:text-white"
            viewBox="0 0 24 24"
            fill="none"
          >
            <circle
              className="opacity-25"
              cx="12"
              cy="12"
              r="10"
              stroke="currentColor"
              strokeWidth="4"
            />
            <path
              className="opacity-75"
              fill="currentColor"
              d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"
            />
          </svg>
        </div>
        <Text className="text-sm text-zinc-500 dark:text-zinc-400">
          Please wait while we complete your authentication...
        </Text>
      </div>
    </AuthLayout>
  );
}




