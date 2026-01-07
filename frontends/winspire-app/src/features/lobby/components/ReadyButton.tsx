import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';

interface ReadyButtonProps {
  isReady: boolean;
  isLoading?: boolean;
  disabled?: boolean;
  onClick?: () => void;
}

/**
 * ReadyButton Component
 * 
 * Button to mark player as ready for match start
 * Features:
 * - Toggle between "Gotowy" and "Anuluj gotowość"
 * - Loading state during API call
 * - Disabled state when opponent not present
 * - Visual feedback with colors
 */
export function ReadyButton({ isReady, isLoading = false, disabled = false, onClick }: ReadyButtonProps) {
  return (
    <button
      onClick={onClick}
      disabled={disabled || isLoading}
      data-testid="ready-button"
      className={`
        relative px-8 py-4 rounded-lg font-semibold text-lg
        transition-all duration-200 transform
        disabled:opacity-50 disabled:cursor-not-allowed
        ${!disabled && !isLoading ? 'hover:scale-105 active:scale-95' : ''}
        ${isReady 
          ? 'bg-amber-500 hover:bg-amber-600 text-white shadow-lg shadow-amber-500/50' 
          : 'bg-green-500 hover:bg-green-600 text-white shadow-lg shadow-green-500/50'
        }
      `}
    >
      {isLoading ? (
        <span className="flex items-center gap-3">
          <LoadingSpinner size="sm" />
          <span>Aktualizacja...</span>
        </span>
      ) : (
        <span className="flex items-center gap-2">
          {isReady ? (
            <>
              <svg className="size-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
              <span>Anuluj gotowość</span>
            </>
          ) : (
            <>
              <svg className="size-6" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
              <span>Jestem gotowy!</span>
            </>
          )}
        </span>
      )}
    </button>
  );
}

