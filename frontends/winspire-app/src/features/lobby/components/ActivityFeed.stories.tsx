/**
 * ActivityFeed Stories
 * Feature: 003-matchmaking-lobby-frontend
 */

import type { Meta, StoryObj } from '@storybook/react';
import { ActivityFeed } from './ActivityFeed';
import type { ActivityFeedItem } from '../types';

// Mock activity feed items
const mockActivityItems: ActivityFeedItem[] = [
  {
    id: '1',
    type: 'participant_joined',
    message: 'ProGamer2024 dołączył do turnieju',
    timestamp: new Date(Date.now() - 5 * 60 * 1000).toISOString(),
    participantName: 'ProGamer2024',
  },
  {
    id: '2',
    type: 'participant_joined',
    message: 'CyberNinja dołączył do turnieju',
    timestamp: new Date(Date.now() - 4 * 60 * 1000).toISOString(),
    participantName: 'CyberNinja',
  },
  {
    id: '3',
    type: 'participant_joined',
    message: 'ElitePlayer dołączył do turnieju',
    timestamp: new Date(Date.now() - 3 * 60 * 1000).toISOString(),
    participantName: 'ElitePlayer',
  },
  {
    id: '4',
    type: 'participant_left',
    message: 'GameMaster opuścił turniej',
    timestamp: new Date(Date.now() - 2 * 60 * 1000).toISOString(),
    participantName: 'GameMaster',
  },
  {
    id: '5',
    type: 'tournament_starting',
    message: 'Turniej rozpoczyna się!',
    timestamp: new Date(Date.now() - 1 * 60 * 1000).toISOString(),
  },
  {
    id: '6',
    type: 'grace_period_started',
    message: 'Rozpoczęto okres łaski - spóźnieni gracze mogą jeszcze dołączyć (30s)',
    timestamp: new Date(Date.now() - 30 * 1000).toISOString(),
  },
];

const meta = {
  title: 'Lobby/ActivityFeed',
  component: ActivityFeed,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    maxItems: {
      control: 'number',
      description: 'Maximum number of items to display',
    },
  },
} satisfies Meta<typeof ActivityFeed>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Empty state - no activity yet
 */
export const Empty: Story = {
  args: {
    items: [],
  },
};

/**
 * Single activity item
 */
export const SingleItem: Story = {
  args: {
    items: [mockActivityItems[0]],
  },
};

/**
 * Multiple activity items
 */
export const MultipleItems: Story = {
  args: {
    items: mockActivityItems,
  },
};

/**
 * Player joined events
 */
export const PlayerJoinedEvents: Story = {
  args: {
    items: mockActivityItems.filter((item) => item.type === 'participant_joined'),
  },
};

/**
 * Player left events
 */
export const PlayerLeftEvents: Story = {
  args: {
    items: [mockActivityItems[3]],
  },
};

/**
 * Tournament events
 */
export const TournamentEvents: Story = {
  args: {
    items: mockActivityItems.filter((item) => 
      item.type === 'tournament_starting' || item.type === 'grace_period_started'
    ),
  },
};

/**
 * Long activity feed
 * Tests scrolling behavior
 */
export const LongFeed: Story = {
  args: {
    items: Array.from({ length: 30 }, (_, i) => ({
      id: `item-${i}`,
      type: i % 3 === 0 ? 'participant_joined' : i % 3 === 1 ? 'participant_left' : 'tournament_starting',
      message: `Test activity message ${i + 1}`,
      timestamp: new Date(Date.now() - i * 10 * 1000).toISOString(),
      participantName: i % 2 === 0 ? `Player${i}` : undefined,
    })) as ActivityFeedItem[],
  },
};

/**
 * With max items limit
 * Only shows the last 5 items
 */
export const WithMaxItems: Story = {
  args: {
    items: mockActivityItems,
    maxItems: 5,
  },
};

/**
 * Real-time simulation
 * Shows how new items appear with animation
 */
export const RealtimeSimulation: Story = {
  args: {
    items: mockActivityItems.slice(0, 3),
  },
  parameters: {
    docs: {
      description: {
        story: 'New items will fade in and slide up from the bottom when added. The feed auto-scrolls to show the latest activity.',
      },
    },
  },
};

/**
 * Mixed activity types
 * Shows all different activity types together
 */
export const AllActivityTypes: Story = {
  args: {
    items: [
      {
        id: '1',
        type: 'participant_joined',
        message: 'ProGamer2024 dołączył do turnieju',
        timestamp: new Date(Date.now() - 5 * 1000).toISOString(),
        participantName: 'ProGamer2024',
      },
      {
        id: '2',
        type: 'participant_left',
        message: 'OldPlayer opuścił turniej',
        timestamp: new Date(Date.now() - 4 * 1000).toISOString(),
        participantName: 'OldPlayer',
      },
      {
        id: '3',
        type: 'tournament_starting',
        message: 'Turniej rozpoczyna się!',
        timestamp: new Date(Date.now() - 3 * 1000).toISOString(),
      },
      {
        id: '4',
        type: 'grace_period_started',
        message: 'Rozpoczęto okres łaski - spóźnieni gracze mogą jeszcze dołączyć (30s)',
        timestamp: new Date(Date.now() - 2 * 1000).toISOString(),
      },
    ],
  },
};

