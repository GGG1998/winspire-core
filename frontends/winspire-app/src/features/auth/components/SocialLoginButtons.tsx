import { useState } from 'react';
import { useAuth } from '../hooks/useAuth';

export type OAuthProvider = 'google' | 'twitch' | 'discord';

interface SocialLoginButtonsProps {
  providers: OAuthProvider[];
}

export function SocialLoginButtons({ providers }: SocialLoginButtonsProps) {
  const { loginWithGoogle, loginWithTwitch, loginWithDiscord } = useAuth();
  const [loadingProvider, setLoadingProvider] = useState<OAuthProvider | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleOAuthLogin = async (provider: OAuthProvider) => {
    setError(null);
    setLoadingProvider(provider);
    
    let result;
    switch (provider) {
      case 'google':
        result = await loginWithGoogle();
        break;
      case 'twitch':
        result = await loginWithTwitch();
        break;
      case 'discord':
        result = await loginWithDiscord();
        break;
    }

    if (!result.success) {
      setError(result.error || 'Failed to sign in');
      setLoadingProvider(null);
    }
    // If successful, the browser will redirect to OAuth provider
  };

  const getProviderIcon = (provider: OAuthProvider) => {
    switch (provider) {
      case 'google':
        return (
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
            <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
            <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l2.85-2.22.81-.62z" fill="#FBBC05"/>
            <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
          </svg>
        );
      case 'twitch':
        return (
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M11.571 4.714h1.715v5.143H11.57zm4.715 0H18v5.143h-1.714zM6 0L1.714 4.286v15.428h5.143V24l4.286-4.286h3.428L22.286 12V0zm14.571 11.143l-3.428 3.428h-3.429l-3 3v-3H6.857V1.714h13.714z"/>
          </svg>
        );
      case 'discord':
        return (
          <svg className="h-5 w-5" viewBox="0 0 24 24" fill="currentColor">
            <path d="M20.317 4.37a19.791 19.791 0 0 0-4.885-1.515a.074.074 0 0 0-.079.037c-.21.375-.444.864-.608 1.25a18.27 18.27 0 0 0-5.487 0a12.64 12.64 0 0 0-.617-1.25a.077.077 0 0 0-.079-.037A19.736 19.736 0 0 0 3.677 4.37a.07.07 0 0 0-.032.027C.533 9.046-.32 13.58.099 18.057a.082.082 0 0 0 .031.057a19.9 19.9 0 0 0 5.993 3.03a.078.078 0 0 0 .084-.028a14.09 14.09 0 0 0 1.226-1.994a.076.076 0 0 0-.041-.106a13.107 13.107 0 0 1-1.872-.892a.077.077 0 0 1-.008-.128a10.2 10.2 0 0 0 .372-.292a.074.074 0 0 1 .077-.01c3.928 1.793 8.18 1.793 12.062 0a.074.074 0 0 1 .078.01c.12.098.246.198.373.292a.077.077 0 0 1-.006.127a12.299 12.299 0 0 1-1.873.892a.077.077 0 0 0-.041.107c.36.698.772 1.362 1.225 1.993a.076.076 0 0 0 .084.028a19.839 19.839 0 0 0 6.002-3.03a.077.077 0 0 0 .032-.054c.5-5.177-.838-9.674-3.549-13.66a.061.061 0 0 0-.031-.03zM8.02 15.33c-1.183 0-2.157-1.085-2.157-2.419c0-1.333.956-2.419 2.157-2.419c1.21 0 2.176 1.096 2.157 2.42c0 1.333-.956 2.418-2.157 2.418zm7.975 0c-1.183 0-2.157-1.085-2.157-2.419c0-1.333.955-2.419 2.157-2.419c1.21 0 2.176 1.096 2.157 2.42c0 1.333-.946 2.418-2.157 2.418z"/>
          </svg>
        );
    }
  };

  const getProviderName = (provider: OAuthProvider) => {
    switch (provider) {
      case 'google':
        return 'Google';
      case 'twitch':
        return 'Twitch';
      case 'discord':
        return 'Discord';
    }
  };

  const getProviderColor = (provider: OAuthProvider) => {
    switch (provider) {
      case 'google':
        return 'bg-white text-gray-700 hover:bg-gray-50 ring-1 ring-inset ring-gray-300 dark:bg-gray-800 dark:text-white dark:ring-gray-700 dark:hover:bg-gray-700';
      case 'twitch':
        return 'bg-[#9146FF] text-white hover:bg-[#7c3adb] dark:bg-[#9146FF] dark:hover:bg-[#7c3adb]';
      case 'discord':
        return 'bg-[#5865F2] text-white hover:bg-[#4752c4] dark:bg-[#5865F2] dark:hover:bg-[#4752c4]';
    }
  };

  return (
    <div className="space-y-3">
      {error && (
        <div className="rounded-lg bg-red-50 p-3 text-sm text-red-800 ring-1 ring-red-200 dark:bg-red-900/20 dark:text-red-200 dark:ring-red-800">
          {error}
        </div>
      )}
      
      {providers.map((provider) => (
        <button
          key={provider}
          type="button"
          onClick={() => handleOAuthLogin(provider)}
          disabled={loadingProvider !== null}
          className={`
            flex w-full items-center justify-center gap-3 rounded-lg px-4 py-2.5 text-sm font-semibold
            shadow-sm transition-colors focus-visible:outline focus-visible:outline-2 
            focus-visible:outline-offset-2 disabled:cursor-not-allowed disabled:opacity-50
            ${getProviderColor(provider)}
          `}
        >
          {loadingProvider === provider ? (
            <svg className="h-5 w-5 animate-spin" viewBox="0 0 24 24" fill="none">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"/>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"/>
            </svg>
          ) : (
            getProviderIcon(provider)
          )}
          <span>
            {loadingProvider === provider 
              ? 'Redirecting...' 
              : `Continue with ${getProviderName(provider)}`
            }
          </span>
        </button>
      ))}
    </div>
  );
}




