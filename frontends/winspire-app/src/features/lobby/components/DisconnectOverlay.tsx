import { useEffect, useState } from 'react';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';

interface DisconnectOverlayProps {
  disconnectedPlayerName: string;
  isCurrentUserDisconnected: boolean;
  remainingSeconds: number;
  onTimeExpired?: () => void;
}

/**
 * DisconnectOverlay Component
 * 
 * Full-screen overlay shown when a player disconnects
 * Features:
 * - 30-second countdown timer
 * - Different messages for current user vs opponent disconnect
 * - Pulsing animation
 * - "Opponent Disqualified" message when timer expires
 * - Auto-callback on expiration
 */
export function DisconnectOverlay({ 
  disconnectedPlayerName,
  isCurrentUserDisconnected,
  remainingSeconds: initialSeconds,
  onTimeExpired
}: DisconnectOverlayProps) {
  const [seconds, setSeconds] = useState(initialSeconds);
  const [hasExpired, setHasExpired] = useState(false);

  useEffect(() => {
    setSeconds(initialSeconds);
  }, [initialSeconds]);

  useEffect(() => {
    if (seconds <= 0 && !hasExpired) {
      setHasExpired(true);
      onTimeExpired?.();
      return;
    }

    const timer = setInterval(() => {
      setSeconds((prev) => {
        if (prev <= 1) {
          setHasExpired(true);
          onTimeExpired?.();
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [seconds, hasExpired, onTimeExpired]);

  // Color based on remaining time
  const getColorClass = () => {
    if (hasExpired) return 'text-red-500 dark:text-red-400';
    if (seconds <= 10) return 'text-red-500 dark:text-red-400';
    if (seconds <= 20) return 'text-amber-500 dark:text-amber-400';
    return 'text-blue-500 dark:text-blue-400';
  };

  const getBgClass = () => {
    if (hasExpired) return 'bg-red-500/10 dark:bg-red-500/20';
    if (seconds <= 10) return 'bg-red-500/10 dark:bg-red-500/20';
    if (seconds <= 20) return 'bg-amber-500/10 dark:bg-amber-500/20';
    return 'bg-blue-500/10 dark:bg-blue-500/20';
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
      <div className="max-w-md w-full mx-4">
        <div className={`rounded-2xl ${getBgClass()} p-8 text-center`}>
          {/* Icon */}
          {hasExpired ? (
            <div className="text-8xl mb-4">❌</div>
          ) : isCurrentUserDisconnected ? (
            <div className="text-8xl mb-4 animate-bounce">⚠️</div>
          ) : (
            <div className="text-8xl mb-4 animate-pulse">⏳</div>
          )}

          {/* Title */}
          {hasExpired ? (
            <h2 className="text-3xl font-bold text-red-600 dark:text-red-400 mb-4">
              Dyskwalifikacja
            </h2>
          ) : isCurrentUserDisconnected ? (
            <h2 className="text-3xl font-bold text-amber-600 dark:text-amber-400 mb-4">
              Utracono połączenie
            </h2>
          ) : (
            <h2 className="text-3xl font-bold text-blue-600 dark:text-blue-400 mb-4">
              Przeciwnik się rozłączył
            </h2>
          )}

          {/* Message */}
          {hasExpired ? (
            <div className="mb-6">
              <p className="text-xl text-red-700 dark:text-red-300 font-semibold mb-2">
                {isCurrentUserDisconnected 
                  ? 'Zostałeś zdyskwalifikowany'
                  : `${disconnectedPlayerName} został zdyskwalifikowany`
                }
              </p>
              <p className="text-sm text-red-600 dark:text-red-400">
                Czas na ponowne połączenie minął
              </p>
            </div>
          ) : isCurrentUserDisconnected ? (
            <div className="mb-6">
              <p className="text-lg text-amber-700 dark:text-amber-300 mb-2">
                Próba ponownego połączenia...
              </p>
              <p className="text-sm text-amber-600 dark:text-amber-400">
                Połącz się ponownie w ciągu {seconds} sekund lub zostaniesz zdyskwalifikowany
              </p>
            </div>
          ) : (
            <div className="mb-6">
              <p className="text-lg text-blue-700 dark:text-blue-300 mb-2">
                <span className="font-semibold">{disconnectedPlayerName}</span> stracił połączenie
              </p>
              <p className="text-sm text-blue-600 dark:text-blue-400">
                Czekamy na ponowne połączenie...
              </p>
            </div>
          )}

          {/* Countdown */}
          {!hasExpired && (
            <div className="mb-6">
              <div className={`inline-flex items-center justify-center size-32 rounded-full ${getBgClass()}`}>
                <span className={`text-6xl font-bold ${getColorClass()}`}>
                  {seconds}
                </span>
              </div>
              <p className="mt-2 text-sm text-gray-600 dark:text-gray-400">
                sekund do dyskwalifikacji
              </p>
            </div>
          )}

          {/* Loading Spinner */}
          {!hasExpired && (
            <div className="flex items-center justify-center gap-2">
              <LoadingSpinner size="sm" />
              <span className="text-sm text-gray-600 dark:text-gray-400">
                Oczekiwanie na ponowne połączenie...
              </span>
            </div>
          )}

          {/* Expired Actions */}
          {hasExpired && !isCurrentUserDisconnected && (
            <div className="mt-6">
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Otrzymujesz zwycięstwo przez walkower
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}














