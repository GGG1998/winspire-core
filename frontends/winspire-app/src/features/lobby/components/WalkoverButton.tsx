import { useState } from 'react';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';

interface WalkoverButtonProps {
  onClaim: () => Promise<void>;
  disabled?: boolean;
}

/**
 * WalkoverButton Component
 * 
 * Button to claim walkover after 2-minute opponent no-show
 * Features:
 * - Prominent call-to-action styling
 * - Loading state during API call
 * - Disabled state
 * - Confirmation modal (optional)
 */
export function WalkoverButton({ onClaim, disabled = false }: WalkoverButtonProps) {
  const [isLoading, setIsLoading] = useState(false);
  const [showConfirm, setShowConfirm] = useState(false);

  const handleClick = () => {
    setShowConfirm(true);
  };

  const handleConfirm = async () => {
    try {
      setIsLoading(true);
      await onClaim();
      setShowConfirm(false);
    } catch (err) {
      console.error('[WalkoverButton] Claim failed:', err);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <>
      <button
        onClick={handleClick}
        disabled={disabled || isLoading}
        className="w-full px-6 py-4 rounded-lg font-semibold text-lg bg-amber-500 hover:bg-amber-600 text-white shadow-lg transition-all disabled:opacity-50 disabled:cursor-not-allowed"
      >
        {isLoading ? (
          <span className="flex items-center justify-center gap-2">
            <LoadingSpinner size="sm" />
            Zgłaszanie...
          </span>
        ) : (
          'Zgłoś walkower (przeciwnik nie stawił się)'
        )}
      </button>

      {/* Confirmation Modal */}
      {showConfirm && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm">
          <div className="bg-white dark:bg-gray-800 rounded-lg p-6 max-w-md mx-4 shadow-xl">
            <h3 className="text-xl font-bold text-gray-900 dark:text-white mb-3">
              Potwierdź walkower
            </h3>
            <p className="text-gray-600 dark:text-gray-400 mb-6">
              Czy na pewno chcesz zgłosić walkower? Ta akcja spowoduje automatyczną wygraną i awans do następnej rundy.
            </p>
            <div className="flex items-center gap-3">
              <button
                onClick={handleConfirm}
                disabled={isLoading}
                className="flex-1 px-4 py-2 bg-amber-600 hover:bg-amber-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50"
              >
                {isLoading ? (
                  <span className="flex items-center justify-center gap-2">
                    <LoadingSpinner size="sm" />
                    Zgłaszanie...
                  </span>
                ) : (
                  'Tak, zgłoś walkower'
                )}
              </button>
              <button
                onClick={() => setShowConfirm(false)}
                disabled={isLoading}
                className="px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded-lg font-medium transition-colors disabled:opacity-50"
              >
                Anuluj
              </button>
            </div>
          </div>
        </div>
      )}
    </>
  );
}






