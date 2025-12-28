import { useEffect, useState } from 'react';

interface MatchStartCountdownProps {
  initialSeconds: number;
  onComplete?: () => void;
}

/**
 * MatchStartCountdown Component
 * 
 * Large animated countdown (3, 2, 1) before match starts
 * Features:
 * - Large centered numbers
 * - Pulse animation
 * - Color transition (green → yellow → red)
 * - Auto-completion callback
 */
export function MatchStartCountdown({ initialSeconds, onComplete }: MatchStartCountdownProps) {
  const [seconds, setSeconds] = useState(initialSeconds);

  useEffect(() => {
    setSeconds(initialSeconds);
  }, [initialSeconds]);

  useEffect(() => {
    if (seconds <= 0) {
      onComplete?.();
      return;
    }

    const timer = setTimeout(() => {
      setSeconds((prev) => prev - 1);
    }, 1000);

    return () => clearTimeout(timer);
  }, [seconds, onComplete]);

  // Determine color based on remaining time
  const getColorClass = () => {
    if (seconds === 3) return 'text-green-500 dark:text-green-400';
    if (seconds === 2) return 'text-yellow-500 dark:text-yellow-400';
    if (seconds === 1) return 'text-red-500 dark:text-red-400';
    return 'text-gray-500';
  };

  const getBgClass = () => {
    if (seconds === 3) return 'bg-green-500/10 dark:bg-green-500/20';
    if (seconds === 2) return 'bg-yellow-500/10 dark:bg-yellow-500/20';
    if (seconds === 1) return 'bg-red-500/10 dark:bg-red-500/20';
    return 'bg-gray-500/10';
  };

  if (seconds <= 0) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/80 backdrop-blur-sm">
      <div className="text-center">
        {/* Countdown Number */}
        <div className={`
          inline-flex items-center justify-center
          size-64 rounded-full
          ${getBgClass()}
          animate-pulse
        `}>
          <span className={`
            text-9xl font-bold
            ${getColorClass()}
          `}>
            {seconds}
          </span>
        </div>

        {/* Message */}
        <p className="mt-8 text-2xl font-semibold text-white">
          Mecz rozpoczyna się za {seconds}...
        </p>
      </div>
    </div>
  );
}

