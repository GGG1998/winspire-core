import { Avatar } from '../../../shared/components/ui/avatar';

interface ByeWaitingStateProps {
  playerName: string;
  playerAvatarUrl?: string | null;
  nextMatchSlot?: string; // e.g., "Winner of Match 3"
  roundName?: string; // e.g., "Półfinały"
  onRefresh?: () => void;
}

/**
 * ByeWaitingState Component
 * 
 * Displays waiting state for players who received a BYE
 * Features:
 * - "Auto-advanced" congratulations message
 * - Next match slot display ("Winner of Match X")
 * - Visual indication of waiting status
 * - Round progression info
 */
export function ByeWaitingState({ 
  playerName, 
  playerAvatarUrl,
  nextMatchSlot = 'będzie znany wkrótce',
  roundName = 'następnej rundy',
  onRefresh
}: ByeWaitingStateProps) {
  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      {/* Success Card */}
      <div className="rounded-lg border-2 border-green-200 dark:border-green-800 bg-green-50 dark:bg-green-900/20 p-8 mb-6">
        <div className="text-center">
          {/* Trophy Icon */}
          <div className="text-8xl mb-4">🏆</div>
          
          {/* Title */}
          <h2 className="text-3xl font-bold text-green-800 dark:text-green-200 mb-3">
            Automatyczny awans!
          </h2>
          
          {/* Description */}
          <p className="text-lg text-green-700 dark:text-green-300 mb-6">
            Zostałeś automatycznie awansowany do {roundName} z powodu BYE
          </p>

          {/* Player Info */}
          <div className="flex items-center justify-center gap-4 mb-6">
            {playerAvatarUrl ? (
              <img 
                src={playerAvatarUrl} 
                alt={playerName}
                className="size-16 rounded-full border-4 border-green-300 dark:border-green-700"
              />
            ) : (
              <Avatar
                className="size-16 text-xl border-4 border-green-300 dark:border-green-700"
                initials={playerName.slice(0, 2).toUpperCase()}
              />
            )}
            <div className="text-left">
              <p className="text-sm text-green-600 dark:text-green-400 font-medium">
                Gracz
              </p>
              <p className="text-xl font-bold text-green-900 dark:text-green-100">
                {playerName}
              </p>
            </div>
          </div>
        </div>
      </div>

      {/* Next Match Info Card */}
      <div className="rounded-lg border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-6 mb-6">
        <div className="text-center">
          <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-3">
            Twój następny mecz
          </h3>
          
          <div className="flex items-center justify-center gap-4 py-4">
            {/* You */}
            <div className="flex flex-col items-center">
              {playerAvatarUrl ? (
                <img 
                  src={playerAvatarUrl} 
                  alt={playerName}
                  className="size-12 rounded-full border-2 border-cyan-400 dark:border-cyan-600 mb-2"
                />
              ) : (
                <Avatar
                  className="size-12 border-2 border-cyan-400 dark:border-cyan-600 mb-2"
                  initials={playerName.slice(0, 2).toUpperCase()}
                />
              )}
              <p className="text-sm font-medium text-gray-900 dark:text-white">
                {playerName}
              </p>
              <p className="text-xs text-cyan-600 dark:text-cyan-400">
                (Ty)
              </p>
            </div>

            {/* VS */}
            <div className="px-4">
              <span className="text-2xl font-bold text-gray-400 dark:text-gray-500">
                VS
              </span>
            </div>

            {/* Opponent (TBD) */}
            <div className="flex flex-col items-center">
              <div className="size-12 rounded-full border-2 border-dashed border-gray-300 dark:border-gray-600 bg-gray-100 dark:bg-gray-700 flex items-center justify-center mb-2">
                <span className="text-gray-400 dark:text-gray-500 text-xl">?</span>
              </div>
              <p className="text-sm font-medium text-gray-500 dark:text-gray-400 text-center max-w-32">
                {nextMatchSlot}
              </p>
            </div>
          </div>

          <p className="text-sm text-gray-600 dark:text-gray-400 mt-4">
            Twój przeciwnik zostanie wyłoniony z aktualnie rozgrywanych meczów
          </p>
        </div>
      </div>

      {/* Info Card */}
      <div className="rounded-lg border border-blue-200 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4">
        <div className="flex items-start gap-3">
          <div className="text-2xl">ℹ️</div>
          <div className="text-sm text-blue-800 dark:text-blue-200">
            <p className="font-medium mb-2">Co to jest BYE?</p>
            <p className="mb-2">
              W przypadku nieparzystej liczby uczestników, jeden gracz w każdej rundzie automatycznie awansuje 
              do następnej rundy bez konieczności rozgrywania meczu.
            </p>
            <p>
              Zostaniesz automatycznie powiadomiony, gdy Twój następny mecz będzie gotowy do rozpoczęcia.
            </p>
          </div>
        </div>
      </div>

      {/* Refresh Button (Optional) */}
      {onRefresh && (
        <div className="mt-6 text-center">
          <button
            onClick={onRefresh}
            className="px-6 py-2 text-sm font-medium text-gray-700 dark:text-gray-300 hover:text-gray-900 dark:hover:text-white transition-colors"
          >
            Odśwież status
          </button>
        </div>
      )}
    </div>
  );
}

