import { useEffect, useState } from 'react';
import { GRACE_PERIOD_DURATION } from '../constants';

interface GracePeriodIndicatorProps {
  gracePeriodEndsAt: string | null;
  participantCount: number;
  isUpdating?: boolean;
}

/**
 * Display grace period countdown and participant count updates
 * Shows 30-second countdown with real-time participant count
 */
export function GracePeriodIndicator({
  gracePeriodEndsAt,
  participantCount,
  isUpdating = false,
}: GracePeriodIndicatorProps) {
  const [remainingSeconds, setRemainingSeconds] = useState(0);

  useEffect(() => {
    if (!gracePeriodEndsAt) {
      setRemainingSeconds(0);
      return;
    }

    const calculateRemaining = () => {
      const end = new Date(gracePeriodEndsAt).getTime();
      const now = Date.now();
      const diff = end - now;
      return Math.max(0, Math.floor(diff / 1000));
    };

    setRemainingSeconds(calculateRemaining());

    const interval = setInterval(() => {
      const remaining = calculateRemaining();
      setRemainingSeconds(remaining);

      if (remaining <= 0) {
        clearInterval(interval);
      }
    }, 1000);

    return () => clearInterval(interval);
  }, [gracePeriodEndsAt]);

  if (!gracePeriodEndsAt || remainingSeconds <= 0) {
    return null;
  }

  const progress = ((GRACE_PERIOD_DURATION - remainingSeconds) / GRACE_PERIOD_DURATION) * 100;
  const isLowTime = remainingSeconds <= 10;

  return (
    <div className="rounded-lg border-2 border-amber-300 bg-amber-50 p-4">
      <div className="flex items-center gap-3">
        <div className={`text-3xl ${isLowTime ? 'animate-pulse' : ''}`}>⏳</div>
        
        <div className="flex-1">
          <div className="flex items-center justify-between mb-2">
            <h3 className="text-sm font-semibold text-amber-900">
              Finalizacja drabinki...
            </h3>
            <span
              className={`text-2xl font-bold ${
                isLowTime ? 'text-red-600' : 'text-amber-700'
              }`}
            >
              {remainingSeconds}s
            </span>
          </div>

          {/* Progress bar */}
          <div className="h-2 rounded-full bg-amber-200 overflow-hidden">
            <div
              className={`h-full transition-all duration-1000 ease-linear ${
                isLowTime ? 'bg-red-500' : 'bg-amber-500'
              }`}
              style={{ width: `${progress}%` }}
            />
          </div>

          {/* Participant count */}
          <div className="mt-2 flex items-center gap-2">
            <span className="text-sm text-amber-800">
              Liczba graczy:
            </span>
            <span
              className={`inline-flex items-center rounded-full px-2.5 py-0.5 text-sm font-semibold ${
                isUpdating
                  ? 'bg-green-100 text-green-800 animate-pulse'
                  : 'bg-amber-100 text-amber-800'
              }`}
            >
              {participantCount}
              {isUpdating && (
                <svg
                  className="ml-1 size-3 animate-spin"
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
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
              )}
            </span>
          </div>

          <p className="mt-2 text-xs text-amber-700">
            Spóźnieni gracze mogą jeszcze dołączyć. Drabinka zostanie wygenerowana po zakończeniu okresu łaski.
          </p>
        </div>
      </div>
    </div>
  );
}

