/**
 * PostMatchModal Stories
 * Feature: RoundManager post-match navigation
 */

import type { Meta, StoryObj } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';
import { PostMatchModal } from './PostMatchModal';
import type { PostMatchOutcome } from '../types';

const meta = {
  title: 'Lobby/PostMatchModal',
  component: PostMatchModal,
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
  decorators: [
    (Story) => (
      <MemoryRouter>
        <div className="min-h-screen bg-gray-100 dark:bg-gray-900">
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
  argTypes: {
    outcome: {
      control: 'object',
      description: 'Post-match outcome determining modal content',
    },
    onClose: {
      action: 'closed',
      description: 'Called when modal is closed',
    },
  },
} satisfies Meta<typeof PostMatchModal>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * No outcome - modal not shown
 */
export const NoOutcome: Story = {
  args: {
    outcome: null,
    onClose: () => {},
  },
  parameters: {
    docs: {
      description: {
        story: 'When outcome is null, no modal is rendered.',
      },
    },
  },
};

/**
 * Winner outcome - advance to next round
 */
export const Winner: Story = {
  args: {
    outcome: {
      type: 'winner',
      payload: {
        tournament_id: 'tournament-123',
        match_id: 'match-456',
        next_round_number: 2,
        message: 'Świetna gra! Przechodzisz do następnej rundy.',
        prelobby_url: '/tournaments/tournament-123/lobby',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  parameters: {
    docs: {
      description: {
        story: 'Shown to the winner of a match. Provides navigation to the pre-lobby for the next round.',
      },
    },
  },
};

/**
 * Winner - Semifinal advance
 */
export const WinnerSemifinal: Story = {
  args: {
    outcome: {
      type: 'winner',
      payload: {
        tournament_id: 'tournament-123',
        match_id: 'match-789',
        next_round_number: 3,
        message: 'Niesamowite! Awansujesz do półfinału!',
        prelobby_url: '/tournaments/tournament-123/prelobby',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  name: 'Winner - Semifinal',
};

/**
 * Eliminated outcome - tournament over for player
 */
export const Eliminated: Story = {
  args: {
    outcome: {
      type: 'eliminated',
      payload: {
        tournament_id: 'tournament-123',
        match_id: 'match-456',
        final_position: 5,
        message: 'Dziękujemy za udział w turnieju!',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  parameters: {
    docs: {
      description: {
        story: 'Shown to players who lost their match and are eliminated from the tournament.',
      },
    },
  },
};

/**
 * Eliminated - Early round
 */
export const EliminatedEarlyRound: Story = {
  args: {
    outcome: {
      type: 'eliminated',
      payload: {
        tournament_id: 'tournament-123',
        match_id: 'match-101',
        final_position: 9,
        message: 'Nie tym razem, ale spróbuj ponownie!',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  name: 'Eliminated - Early Round (9th place)',
};

/**
 * Eliminated - Semifinal loss (3rd-4th place)
 */
export const EliminatedSemifinal: Story = {
  args: {
    outcome: {
      type: 'eliminated',
      payload: {
        tournament_id: 'tournament-123',
        match_id: 'match-semifinal',
        final_position: 3,
        message: 'Świetny wynik! Zajmujesz miejsce na podium!',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  name: 'Eliminated - Semifinal (3rd place)',
};

/**
 * Eliminated - No position shown
 */
export const EliminatedNoPosition: Story = {
  args: {
    outcome: {
      type: 'eliminated',
      payload: {
        tournament_id: 'tournament-123',
        match_id: 'match-456',
        final_position: 0,
        message: 'Dziękujemy za grę!',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  name: 'Eliminated - No Position',
  parameters: {
    docs: {
      description: {
        story: 'When final_position is 0, the position line is not shown.',
      },
    },
  },
};

/**
 * Champion outcome - tournament winner!
 */
export const Champion: Story = {
  args: {
    outcome: {
      type: 'champion',
      payload: {
        tournament_id: 'tournament-123',
        champion_id: 'user-winner',
        message: 'Zdobywasz tytuł mistrza turnieju!',
        prize_summary: 'Nagroda: 500 PLN + Ekskluzywna skórka',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  parameters: {
    docs: {
      description: {
        story: 'Shown to the tournament champion. Displays prize information and celebration.',
      },
    },
  },
};

/**
 * Champion - No prize
 */
export const ChampionNoPrize: Story = {
  args: {
    outcome: {
      type: 'champion',
      payload: {
        tournament_id: 'tournament-123',
        champion_id: 'user-winner',
        message: 'Gratulacje! Wygrywasz turniej!',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  name: 'Champion - No Prize',
  parameters: {
    docs: {
      description: {
        story: 'Champion modal without prize summary displayed.',
      },
    },
  },
};

/**
 * Champion - Large prize
 */
export const ChampionLargePrize: Story = {
  args: {
    outcome: {
      type: 'champion',
      payload: {
        tournament_id: 'tournament-123',
        champion_id: 'user-winner',
        message: 'NIESAMOWITE! Jesteś absolutnym mistrzem!',
        prize_summary: 'Nagroda główna: 10,000 PLN + Sprzęt gamingowy o wartości 5,000 PLN + Roczna subskrypcja premium',
      },
    } as PostMatchOutcome,
    onClose: () => {},
  },
  name: 'Champion - Large Prize',
};
