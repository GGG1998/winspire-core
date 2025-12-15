import { type ReactNode } from 'react';
import { useNavigate } from 'react-router';
import { Dialog, DialogTitle, DialogBody, DialogActions } from '../../../shared/components/ui/dialog';
import { Button } from '../../../shared/components/ui/button';
import type { PostMatchOutcome } from '../types';

interface PostMatchModalProps {
  outcome: PostMatchOutcome | null;
  onClose: () => void;
}

interface OutcomeConfig {
  emoji: string;
  title: string;
  titleColor: string;
  primaryButtonLabel: string;
  primaryButtonColor: 'cyan' | 'zinc' | 'yellow';
  showSecondaryButton: boolean;
  getExtraContent: (outcome: PostMatchOutcome) => ReactNode;
  getPrimaryAction: (outcome: PostMatchOutcome) => string;
}

const OUTCOME_CONFIG: Record<PostMatchOutcome['type'], OutcomeConfig> = {
  winner: {
    emoji: '🎉',
    title: 'Wygrałeś!',
    titleColor: 'text-green-600 dark:text-green-400',
    primaryButtonLabel: 'Przejdź do pre-lobby',
    primaryButtonColor: 'cyan',
    showSecondaryButton: false,
    getExtraContent: (outcome) => {
      if (outcome.type !== 'winner') return null;
      return (
        <div className="mt-4 p-3 bg-cyan-50 dark:bg-cyan-900/30 rounded-lg">
          <p className="text-center text-cyan-800 dark:text-cyan-200 font-semibold">
            Następna runda: {outcome.payload.next_round_number}
          </p>
        </div>
      );
    },
    getPrimaryAction: (outcome) =>
      outcome.type === 'winner' ? outcome.payload.prelobby_url : '/',
  },

  eliminated: {
    emoji: '💔',
    title: 'Koniec przygody',
    titleColor: 'text-gray-600 dark:text-gray-400',
    primaryButtonLabel: 'Zobacz turniej',
    primaryButtonColor: 'zinc',
    showSecondaryButton: true,
    getExtraContent: (outcome) => {
      if (outcome.type !== 'eliminated' || outcome.payload.final_position <= 0) return null;
      return (
        <p className="mt-2 text-sm text-center text-gray-500 dark:text-gray-500">
          Twoja pozycja: {outcome.payload.final_position}. miejsce
        </p>
      );
    },
    getPrimaryAction: (outcome) =>
      outcome.type === 'eliminated' ? `/tournaments/${outcome.payload.tournament_id}` : '/',
  },

  champion: {
    emoji: '🏆',
    title: 'Jesteś mistrzem!',
    titleColor: 'text-yellow-500 dark:text-yellow-400',
    primaryButtonLabel: 'Zobacz wyniki turnieju',
    primaryButtonColor: 'yellow',
    showSecondaryButton: true,
    getExtraContent: (outcome) => {
      if (outcome.type !== 'champion' || !outcome.payload.prize_summary) return null;
      return (
        <div className="mt-4 p-4 bg-yellow-50 dark:bg-yellow-900/20 rounded-lg">
          <p className="text-sm font-medium text-center text-yellow-800 dark:text-yellow-200">
            {outcome.payload.prize_summary}
          </p>
        </div>
      );
    },
    getPrimaryAction: (outcome) =>
      outcome.type === 'champion' ? `/tournaments/${outcome.payload.tournament_id}/results` : '/',
  },
};

/**
 * PostMatchModal Component
 *
 * Modal shown after match completion based on player outcome:
 * - Winner: Return to pre-lobby for next round
 * - Eliminated: Thanks for playing, eliminated from tournament
 * - Champion: Congratulations, tournament winner!
 */
export function PostMatchModal({ outcome, onClose }: PostMatchModalProps) {
  const navigate = useNavigate();
  
  if (!outcome) return null;

  const config = OUTCOME_CONFIG[outcome.type];

  const handlePrimaryAction = () => {
    navigate(config.getPrimaryAction(outcome));
    onClose();
  };

  const handleSecondaryAction = () => {
    navigate('/');
    onClose();
  };

  return (
    <Dialog open={!!outcome} onClose={onClose} size="lg">
      <div className="text-center">
        <div className="text-7xl mb-4">{config.emoji}</div>
        <DialogTitle className={`text-3xl font-bold mb-2 ${config.titleColor}`}>
          {config.title}
        </DialogTitle>
      </div>

      <DialogBody>
        <p className="text-lg text-center text-gray-600 dark:text-gray-400">
          {outcome.payload.message}
        </p>
        {config.getExtraContent(outcome)}
      </DialogBody>

      <DialogActions>
        {config.showSecondaryButton && (
          <Button outline onClick={handleSecondaryAction}>
            Wróć do strony głównej
          </Button>
        )}
        <Button
          color={config.primaryButtonColor}
          onClick={handlePrimaryAction}
          className={config.showSecondaryButton ? '' : 'w-full sm:w-auto'}
        >
          {config.primaryButtonLabel}
        </Button>
      </DialogActions>
    </Dialog>
  );
}
