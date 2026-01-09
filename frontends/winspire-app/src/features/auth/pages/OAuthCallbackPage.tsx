import { useEffect, useState, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { AuthLayout } from '../../../shared/components/ui/auth-layout';
import { Heading } from '../../../shared/components/ui/heading';
import { Text } from '../../../shared/components/ui/text';

/**
 * Returns the appropriate login URL based on OAuth provider.
 * - Twitch/Discord providers → streamer login
 * - Google/unknown providers → user login
 */
function getLoginUrl(provider: string | undefined): string {
  if (provider === 'twitch' || provider === 'discord') {
    return '/auth/streamer/login';
  }
  return '/auth/user/login';
}

export function OAuthCallbackPage() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const handledRef = useRef(false);

  useEffect(() => {
    // Prevent double execution in React StrictMode
    if (handledRef.current) {
      return;
    }

    const processSession = async () => {
      try {
        // Check localStorage directly to avoid Supabase lock issues
        const storageKey = 'winspire-auth';
        const storedSession = localStorage.getItem(storageKey);

        if (!storedSession) {
          setError('Authentication failed. Please try again.');
          setTimeout(() => navigate(getLoginUrl(undefined)), 3000);
          return;
        }

        const sessionData = JSON.parse(storedSession);
        const session = sessionData?.session || sessionData;

        if (!session?.access_token || !session?.user) {
          setError('Authentication failed. Please try again.');
          setTimeout(() => navigate(getLoginUrl(undefined)), 3000);
          return;
        }

        handledRef.current = true;

        // Simply redirect to profile - AuthContext will handle loading user data
        // The profile page will redirect to complete-profile if needed
        navigate('/auth/profile', { replace: true });
      } catch (err) {
        setError(err instanceof Error ? err.message : 'An unexpected error occurred');
        setTimeout(() => navigate(getLoginUrl(undefined)), 3000);
      }
    };

    // Small delay to let Supabase finish processing the code from URL
    const timeoutId = setTimeout(processSession, 100);

    return () => clearTimeout(timeoutId);
  }, [navigate]);

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
