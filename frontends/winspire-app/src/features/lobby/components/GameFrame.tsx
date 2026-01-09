import React, { useEffect, useState, useRef } from 'react';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import { ReadyButton } from './ReadyButton';
import { GAME_IFRAME_CONFIG } from '../constants';

interface GameFrameProps {
  gameUrl: string;
  matchId: string;
  token?: string;                  // JWT token passed via URL parameter
  onGameLoaded?: () => void;       // Called when game finishes loading
  onGameError?: (error: string) => void;
  // Ready overlay props
  showReadyOverlay?: boolean;      // Show ready button overlay when game loaded
  isPlayerReady?: boolean;         // Current player's ready state
  isOpponentReady?: boolean;       // Opponent's ready state
  onReadyClick?: () => void;       // Callback when ready clicked
  isReadyLoading?: boolean;        // Loading state for ready button
}

/**
 * GameFrame Component
 *
 * Iframe container for loading the game
 * Features:
 * - Centered iframe with proper sizing
 * - Loading state with spinner
 * - Error handling with retry button
 * - Session token passing via URL parameters
 * - 30-second timeout for load failure
 */
export function GameFrame({
  gameUrl,
  matchId,
  token,
  onGameLoaded,
  onGameError,
  showReadyOverlay = false,
  isPlayerReady = false,
  isOpponentReady = false,
  onReadyClick,
  isReadyLoading = false,
}: GameFrameProps) {
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [loadTimeout, setLoadTimeout] = useState(false);
  const [gameLoadedCalled, setGameLoadedCalled] = useState(false);
  const iframeRef = useRef<HTMLIFrameElement>(null);
  const loadTimeoutRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  // Validate gameUrl early
  useEffect(() => {
    if (!gameUrl || gameUrl === '') {
      console.error('[GameFrame] Empty game URL provided');
      setError('Brak URL gry - sprawdź czy gameSnapshot.slug jest dostępny');
      setIsLoading(false);
      onGameError?.('Empty game URL');
    }
  }, [gameUrl, onGameError]);

  // Construct full game URL with session parameters
  const fullGameUrl = React.useMemo(() => {
    // Don't try to construct URL if gameUrl is empty
    if (!gameUrl || gameUrl === '') {
      console.warn('[GameFrame] Empty gameUrl provided, cannot construct URL');
      return '';
    }

    try {
      const url = new URL(gameUrl);
      url.searchParams.set('matchId', matchId);
      if (token) {
        url.searchParams.set('token', token);
      }
      const finalUrl = url.toString();
      return finalUrl;
    } catch (err) {
      setError(`Nieprawidłowy URL gry: ${gameUrl}`);
      setIsLoading(false);
      return '';
    }
  }, [gameUrl, matchId, token]);

  // Set up 30-second timeout for game load
  useEffect(() => {
    // Clear any existing timeout
    if (loadTimeoutRef.current) {
      clearTimeout(loadTimeoutRef.current);
    }

    // Reset states
    // setIsLoading(true);
    setError(null);
    setLoadTimeout(false);
    setGameLoadedCalled(false);

    // Start timeout
    loadTimeoutRef.current = setTimeout(() => {
      if (isLoading) {
        setLoadTimeout(true);
        setError('Gra nie załadowała się w ciągu 30 sekund');
        setIsLoading(false);
        onGameError?.('Game load timeout');
      }
    }, GAME_IFRAME_CONFIG.LOAD_TIMEOUT_MS);

    return () => {
      if (loadTimeoutRef.current) {
        clearTimeout(loadTimeoutRef.current);
      }
    };
  }, [fullGameUrl, onGameError]);

  // Handle iframe load event
  const handleIframeLoad = () => {
    // Clear timeout on successful load
    if (loadTimeoutRef.current) {
      clearTimeout(loadTimeoutRef.current);
    }

    console.log('[GameFrame] Iframe loaded, waiting for game_ready signal from game...');

    // For games that don't send game_ready postMessage, assume loaded after delay
    // Games that send game_ready postMessage will call onGameLoaded sooner
    setTimeout(() => {
      if (isLoading && !gameLoadedCalled) {
        console.log('[GameFrame] Game did not send ready signal, assuming loaded after timeout');
        setIsLoading(false);
        setError(null);

        // Call onGameLoaded callback (only once)
        if (onGameLoaded) {
          setGameLoadedCalled(true);
          onGameLoaded();
        }
      }
    }, 2000); // 2 second buffer for game initialization
  };

  // Handle iframe error
  const handleIframeError = (_e: React.SyntheticEvent<HTMLIFrameElement, Event>) => {
    setError('Nie udało się załadować gry. Sprawdź konsolę przeglądarki.');
    setIsLoading(false);
    onGameError?.('Iframe load error');
  };

  // Retry loading game
  const handleRetry = () => {
    setError(null);
    setIsLoading(true);
    setLoadTimeout(false);

    // Reload iframe by changing key
    if (iframeRef.current) {
      iframeRef.current.src = fullGameUrl;
    }
  };

  return (
    <div className="relative w-full">

      {/* Loading Overlay */}
      {isLoading && !error && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-900 rounded-lg z-10">
          <div className="text-center">
            <LoadingSpinner size="lg" />
            <p className="mt-4 text-sm text-gray-600 dark:text-gray-400">
              {loadTimeout ? 'Gra nie odpowiada...' : 'Ładowanie gry...'}
            </p>
          </div>
        </div>
      )}

      {/* Error State */}
      {error && (
        <div className="absolute inset-0 flex items-center justify-center bg-gray-100 dark:bg-gray-900 rounded-lg z-10">
          <div className="text-center max-w-md">
            <div className="size-16 mx-auto mb-4 rounded-full bg-red-100 dark:bg-red-900/20 flex items-center justify-center">
              <svg className="size-8 text-red-500 dark:text-red-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
              </svg>
            </div>
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Błąd ładowania gry
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400 mb-4">
              {error}
            </p>
            <button
              onClick={handleRetry}
              className="px-4 py-2 bg-cyan-600 text-white rounded-lg hover:bg-cyan-700 transition-colors"
            >
              Spróbuj ponownie
            </button>
          </div>
        </div>
      )}

      {/* Ready Button Overlay - Show when game loaded but player not ready */}
      {!isLoading && !error && fullGameUrl && showReadyOverlay && !isPlayerReady && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/60 rounded-lg z-20">
          <div className="text-center">
            <ReadyButton
              isReady={false}
              isLoading={isReadyLoading}
              disabled={false}
              onClick={onReadyClick}
            />
            <p className="mt-4 text-sm text-white/80">
              Gra załadowana! Kliknij gdy będziesz gotowy.
            </p>
          </div>
        </div>
      )}

      {/* Waiting for opponent overlay - Show when current player ready but opponent not */}
      {!isLoading && !error && fullGameUrl && showReadyOverlay && isPlayerReady && !isOpponentReady && (
        <div className="absolute inset-0 flex items-center justify-center bg-black/60 rounded-lg z-20">
          <div className="text-center">
            <div className="text-6xl mb-4">✅</div>
            <p className="text-xl font-semibold text-white mb-2">Jesteś gotowy!</p>
            <p className="text-sm text-white/80">Oczekiwanie na przeciwnika...</p>
          </div>
        </div>
      )}

      {/* Game Iframe - Only render if we have a valid URL */}
      <div className="relative rounded-lg overflow-hidden bg-black" style={{ aspectRatio: '16/9' }}>
        {fullGameUrl ? (
          <iframe
            ref={iframeRef}
            src={fullGameUrl}
            title="Game"
            className="w-full h-full"
            onLoad={handleIframeLoad}
            onError={handleIframeError}
            allow="fullscreen; gamepad; microphone"
            sandbox="allow-scripts allow-same-origin allow-forms allow-popups"
          />
        ) : (
          <div className="w-full h-full flex items-center justify-center text-white">
            <div className="text-center">
              <p className="text-lg font-semibold mb-2">Brak URL gry</p>
              <p className="text-sm text-gray-400">Oczekiwanie na dane gry...</p>
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
