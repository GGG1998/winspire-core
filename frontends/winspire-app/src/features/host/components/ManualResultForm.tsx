import { useState } from 'react';
import { tournamentApi } from '../api/tournamentApi';
import { LoadingSpinner } from '../../../shared/components/common/LoadingSpinner';
import type { MatchPlayer } from '../types';

interface ManualResultFormProps {
  matchId: string;
  player1: MatchPlayer | null;
  player2: MatchPlayer | null;
  onSuccess?: () => void;
  onCancel?: () => void;
}

/**
 * ManualResultForm Component
 * 
 * Allows host to manually enter match results (fallback for game API failures)
 * Features:
 * - Winner selection (radio buttons)
 * - Optional score inputs
 * - Form validation
 * - Loading state during submission
 * - Success/error feedback
 */
export function ManualResultForm({ 
  matchId, 
  player1, 
  player2,
  onSuccess,
  onCancel
}: ManualResultFormProps) {
  const [selectedWinnerId, setSelectedWinnerId] = useState<string>('');
  const [scorePlayer1, setScorePlayer1] = useState<string>('');
  const [scorePlayer2, setScorePlayer2] = useState<string>('');
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedWinnerId) {
      setError('Wybierz zwycięzcę');
      return;
    }

    try {
      setIsSubmitting(true);
      setError(null);

      const score1 = scorePlayer1 ? parseInt(scorePlayer1, 10) : null;
      const score2 = scorePlayer2 ? parseInt(scorePlayer2, 10) : null;

      await tournamentApi.submitManualResult(
        matchId,
        selectedWinnerId,
        score1,
        score2
      );

      console.log('[ManualResultForm] Result submitted successfully');
      onSuccess?.();
    } catch (err) {
      const errorMessage = err instanceof Error ? err.message : 'Nie udało się zapisać wyniku';
      setError(errorMessage);
      console.error('[ManualResultForm] Submission failed:', errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  if (!player1 || !player2) {
    return (
      <div className="text-center text-gray-500 dark:text-gray-400">
        Brak informacji o zawodnikach
      </div>
    );
  }

  return (
    <div className="rounded-lg border-2 border-amber-300 dark:border-amber-700 bg-amber-50 dark:bg-amber-900/20 p-6">
      <div className="mb-4 flex items-center gap-3">
        <div className="text-3xl">⚠️</div>
        <div>
          <h3 className="text-lg font-semibold text-amber-900 dark:text-amber-200">
            Wymagane ręczne wprowadzenie wyniku
          </h3>
          <p className="text-sm text-amber-700 dark:text-amber-300">
            API gry niedostępne - wprowadź wynik ręcznie
          </p>
        </div>
      </div>

      <form onSubmit={handleSubmit} className="space-y-4">
        {/* Winner Selection */}
        <div>
          <label className="block text-sm font-medium text-gray-900 dark:text-white mb-3">
            Kto wygrał mecz?
          </label>
          <div className="space-y-2">
            {/* Player 1 */}
            <label className={`
              flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all
              ${selectedWinnerId === player1.id 
                ? 'border-green-500 dark:border-green-600 bg-green-50 dark:bg-green-900/20' 
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
              }
            `}>
              <input
                type="radio"
                name="winner"
                value={player1.id}
                checked={selectedWinnerId === player1.id}
                onChange={(e) => setSelectedWinnerId(e.target.value)}
                className="size-4"
              />
              <div className="flex items-center gap-2 flex-1">
                {player1.avatarUrl ? (
                  <img 
                    src={player1.avatarUrl} 
                    alt={player1.displayName}
                    className="size-8 rounded-full"
                  />
                ) : (
                  <div className="size-8 rounded-full bg-gray-200 dark:bg-gray-600 flex items-center justify-center text-xs font-medium">
                    {player1.displayName.slice(0, 2).toUpperCase()}
                  </div>
                )}
                <span className="font-medium text-gray-900 dark:text-white">
                  {player1.displayName}
                </span>
              </div>
            </label>

            {/* Player 2 */}
            <label className={`
              flex items-center gap-3 p-3 rounded-lg border-2 cursor-pointer transition-all
              ${selectedWinnerId === player2.id 
                ? 'border-green-500 dark:border-green-600 bg-green-50 dark:bg-green-900/20' 
                : 'border-gray-200 dark:border-gray-700 hover:border-gray-300 dark:hover:border-gray-600'
              }
            `}>
              <input
                type="radio"
                name="winner"
                value={player2.id}
                checked={selectedWinnerId === player2.id}
                onChange={(e) => setSelectedWinnerId(e.target.value)}
                className="size-4"
              />
              <div className="flex items-center gap-2 flex-1">
                {player2.avatarUrl ? (
                  <img 
                    src={player2.avatarUrl} 
                    alt={player2.displayName}
                    className="size-8 rounded-full"
                  />
                ) : (
                  <div className="size-8 rounded-full bg-gray-200 dark:bg-gray-600 flex items-center justify-center text-xs font-medium">
                    {player2.displayName.slice(0, 2).toUpperCase()}
                  </div>
                )}
                <span className="font-medium text-gray-900 dark:text-white">
                  {player2.displayName}
                </span>
              </div>
            </label>
          </div>
        </div>

        {/* Score Inputs (Optional) */}
        <div>
          <label className="block text-sm font-medium text-gray-900 dark:text-white mb-2">
            Wynik (opcjonalnie)
          </label>
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-xs text-gray-600 dark:text-gray-400 mb-1">
                {player1.displayName}
              </label>
              <input
                type="number"
                min="0"
                value={scorePlayer1}
                onChange={(e) => setScorePlayer1(e.target.value)}
                placeholder="0"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-cyan-500"
              />
            </div>
            <div>
              <label className="block text-xs text-gray-600 dark:text-gray-400 mb-1">
                {player2.displayName}
              </label>
              <input
                type="number"
                min="0"
                value={scorePlayer2}
                onChange={(e) => setScorePlayer2(e.target.value)}
                placeholder="0"
                className="w-full px-3 py-2 rounded-lg border border-gray-300 dark:border-gray-600 bg-white dark:bg-gray-800 text-gray-900 dark:text-white focus:ring-2 focus:ring-cyan-500"
              />
            </div>
          </div>
        </div>

        {/* Error Message */}
        {error && (
          <div className="rounded-lg bg-red-50 dark:bg-red-900/20 border border-red-200 dark:border-red-800 p-3">
            <p className="text-sm text-red-700 dark:text-red-300">
              {error}
            </p>
          </div>
        )}

        {/* Actions */}
        <div className="flex items-center gap-3 pt-2">
          <button
            type="submit"
            disabled={isSubmitting || !selectedWinnerId}
            className="flex-1 px-4 py-2 bg-green-600 hover:bg-green-700 disabled:bg-gray-400 text-white rounded-lg font-medium transition-colors disabled:cursor-not-allowed"
          >
            {isSubmitting ? (
              <span className="flex items-center justify-center gap-2">
                <LoadingSpinner size="sm" />
                Zapisywanie...
              </span>
            ) : (
              'Zapisz wynik'
            )}
          </button>
          
          {onCancel && (
            <button
              type="button"
              onClick={onCancel}
              disabled={isSubmitting}
              className="px-4 py-2 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600 text-gray-900 dark:text-white rounded-lg font-medium transition-colors disabled:cursor-not-allowed"
            >
              Anuluj
            </button>
          )}
        </div>
      </form>
    </div>
  );
}

