/**
 * BracketView Stories
 * Feature: 003-matchmaking-lobby-frontend
 * 
 * Interactive tournament bracket visualization with different sizes
 */

import type { Meta, StoryObj } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';
import { BracketView } from './BracketView';
import type { Tournament, Bracket } from '../types';
import { tournamentApi } from '../api/tournamentApi';

const meta = {
  title: 'Host/BracketView',
  component: BracketView,
  parameters: {
    layout: 'fullscreen',
    backgrounds: {
      default: 'dark',
    },
  },
  decorators: [
    (Story) => (
      <MemoryRouter>
        <div className="min-h-screen bg-gray-50 dark:bg-gray-900 p-4">
          <Story />
        </div>
      </MemoryRouter>
    ),
  ],
} satisfies Meta<typeof BracketView>;

export default meta;
type Story = StoryObj<typeof meta>;

// Helper to generate UUIDs
const generateId = (index: number) => `00000000-0000-0000-0000-${String(index).padStart(12, '0')}`;

// Helper to generate bracket data
function generateBracket(numRounds: number, tournamentId: string): Bracket {
  const rounds = [];
  let matchCounter = 0;
  let participantCounter = 0;

  for (let roundNum = numRounds; roundNum >= 1; roundNum--) {
    const matchesInRound = Math.pow(2, numRounds - roundNum);
    const matches = [];
    
    const roundNames = ['Final', 'Semi-final', 'Quarter-final', 'Round of 16', 'Round of 32', 'Round of 64', 'Round of 128', 'Round of 256', 'Round of 512', 'Round of 1024'];
    const roundName = roundNames[numRounds - roundNum] || `Round ${roundNum}`;

    for (let matchNum = 0; matchNum < matchesInRound; matchNum++) {
      const matchId = generateId(++matchCounter);
      const participant1Id = generateId(1000 + (++participantCounter));
      const participant2Id = matchesInRound > 256 && matchNum % 10 !== 0 ? null : generateId(1000 + (++participantCounter));
      
      let status: 'ready' | 'started' | 'completed' = 'ready';
      let winnerId = null;
      let scorePlayer1 = null;
      let scorePlayer2 = null;
      let startedAt = null;
      let completedAt = null;

      const completionChance = (numRounds - roundNum) / numRounds;
      const random = Math.random();

      if (random < completionChance * 0.8) {
        status = 'completed';
        winnerId = Math.random() > 0.5 ? participant1Id : (participant2Id || participant1Id);
        scorePlayer1 = Math.floor(Math.random() * 3) + 13;
        scorePlayer2 = winnerId === participant1Id ? scorePlayer1 - Math.floor(Math.random() * 5) - 3 : scorePlayer1 + Math.floor(Math.random() * 5) + 3;
        startedAt = new Date(Date.now() - Math.random() * 3600000).toISOString();
        completedAt = new Date(Date.now() - Math.random() * 1800000).toISOString();
      } else if (random < completionChance * 0.9) {
        status = 'started';
        startedAt = new Date(Date.now() - Math.random() * 600000).toISOString();
      }

      matches.push({
        id: matchId,
        roundId: generateId(5000 + roundNum),
        matchNumber: matchNum + 1,
        nextMatchId: roundNum < numRounds ? generateId(matchCounter + Math.floor(matchesInRound / 2)) : null,
        participant1Id: participant1Id,
        participant2Id: participant2Id,
        status,
        participant1Ready: status !== 'ready',
        participant2Ready: status !== 'ready' && participant2Id !== null,
        winnerId,
        scorePlayer1,
        scorePlayer2,
        resultSource: status === 'completed' ? 'game_api' as const : null,
        disconnectedPlayerId: null,
        disconnectedAt: null,
        gameApiMatchId: status !== 'ready' ? `game-${matchId}` : null,
        createdAt: new Date(Date.now() - 86400000).toISOString(),
        startedAt,
        completedAt,
        updatedAt: completedAt || startedAt || new Date(Date.now() - 86400000).toISOString(),
      });
    }

    const roundStatus = matches.every(m => m.status === 'completed') ? 'completed' : matches.some(m => m.status === 'started') ? 'in_progress' : 'pending';
    rounds.unshift({
      id: generateId(5000 + roundNum),
      roundNumber: roundNum,
      roundName,
      matchesCount: matchesInRound,
      status: roundStatus as 'pending' | 'in_progress' | 'completed',
      matches,
    });
  }

  const totalMatches = rounds.reduce((sum, r) => sum + r.matchesCount, 0);
  const byesCount = rounds[0]?.matches.filter(m => !m.participant2Id).length || 0;

  return {
    id: generateId(9000),
    tournamentId,
    totalRounds: numRounds,
    totalMatches,
    byesCount,
    generatedAt: new Date().toISOString(),
    rounds,
  };
}

// Base tournament
const baseTournament: Tournament = {
  id: '702a5025-3059-4704-b184-f980c71555db',
  creatorId: 'host-123',
  name: 'Championship Tournament',
  game: 'CS:GO',
  gameLogoUrl: 'https://cdn.example.com/games/csgo.png',
  status: 'started',
  startTime: new Date(Date.now() + 3600000),
  roomLink: 'https://tournament.example.com/room/702a5025-3059-4704-b184-f980c71555db',
  isCompleted: false,
  createdAt: new Date(Date.now() - 86400000),
  updatedAt: new Date(),
  format: {
    type: 'single_elimination',
    teamSize: 1,
    maxSlots: 64,
    bestOf: 1,
  },
  prize: {
    type: 'cash',
    description: '$1000 Prize Pool',
    value: 1000,
    currency: 'USD',
  },
  participantCount: 32,
};

// Mock tournamentApi.getBracket to return our mock data
function setupMockBracket(bracket: Bracket) {
  const originalGetBracket = tournamentApi.getBracket;
  
  tournamentApi.getBracket = async () => {
    await new Promise(resolve => setTimeout(resolve, 100));
    return bracket;
  };
  
  return () => {
    tournamentApi.getBracket = originalGetBracket;
  };
}

/**
 * Small 2-round bracket (4 participants)
 */
export const SmallBracket: Story = {
  args: {
    tournament: {
      ...baseTournament,
      name: 'Small Tournament (4 players)',
      format: {
        ...baseTournament.format!,
        maxSlots: 4,
      },
    },
  },
  play: async () => {
    const bracket = generateBracket(2, baseTournament.id);
    setupMockBracket(bracket);
  },
};

/**
 * Medium 3-round bracket (8 participants)
 */
export const MediumBracket: Story = {
  args: {
    tournament: {
      ...baseTournament,
      name: 'Medium Tournament (8 players)',
      format: {
        ...baseTournament.format!,
        maxSlots: 8,
      },
    },
  },
  play: async () => {
    const bracket = generateBracket(3, baseTournament.id);
    setupMockBracket(bracket);
  },
};

/**
 * Large 5-round bracket (32 participants)
 */
export const LargeBracket: Story = {
  args: {
    tournament: {
      ...baseTournament,
      name: 'Large Tournament (32 players)',
      format: {
        ...baseTournament.format!,
        maxSlots: 32,
      },
    },
  },
  play: async () => {
    const bracket = generateBracket(5, baseTournament.id);
    setupMockBracket(bracket);
  },
};

/**
 * Extra Large 7-round bracket (128 participants)
 */
export const ExtraLargeBracket: Story = {
  args: {
    tournament: {
      ...baseTournament,
      name: 'Extra Large Tournament (128 players)',
      format: {
        ...baseTournament.format!,
        maxSlots: 128,
      },
    },
  },
  play: async () => {
    const bracket = generateBracket(7, baseTournament.id);
    setupMockBracket(bracket);
  },
};

/**
 * Huge 10-round bracket (1024 participants)
 * Stress test for horizontal scrolling
 */
export const HugeBracket: Story = {
  args: {
    tournament: {
      ...baseTournament,
      name: 'Huge Tournament (1024 players)',
      format: {
        ...baseTournament.format!,
        maxSlots: 1024,
      },
    },
  },
  play: async () => {
    const bracket = generateBracket(10, baseTournament.id);
    setupMockBracket(bracket);
  },
};

/**
 * Not Started Yet
 */
export const NotStarted: Story = {
  args: {
    tournament: {
      ...baseTournament,
      status: 'draft',
    },
  },
};
