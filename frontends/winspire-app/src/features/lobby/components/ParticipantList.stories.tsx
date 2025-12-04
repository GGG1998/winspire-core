/**
 * ParticipantList Stories
 * Feature: 003-matchmaking-lobby-frontend
 */

import type { Meta, StoryObj } from '@storybook/react';
import { ParticipantList } from './ParticipantList';
import type { PreLobbyParticipant } from '../types';

// Mock participants data
const mockParticipants: PreLobbyParticipant[] = [
  {
    id: '123e4567-e89b-12d3-a456-426614174000',
    displayName: 'ProGamer2024',
    avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ProGamer2024',
    joinedAt: new Date(Date.now() - 5 * 60 * 1000).toISOString(), // 5 minutes ago
  },
  {
    id: '223e4567-e89b-12d3-a456-426614174001',
    displayName: 'CyberNinja',
    avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=CyberNinja',
    joinedAt: new Date(Date.now() - 4 * 60 * 1000).toISOString(), // 4 minutes ago
  },
  {
    id: '323e4567-e89b-12d3-a456-426614174002',
    displayName: 'ElitePlayer',
    avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=ElitePlayer',
    joinedAt: new Date(Date.now() - 3 * 60 * 1000).toISOString(), // 3 minutes ago
  },
  {
    id: '423e4567-e89b-12d3-a456-426614174003',
    displayName: 'GameMaster',
    avatarUrl: null, // No avatar
    joinedAt: new Date(Date.now() - 2 * 60 * 1000).toISOString(), // 2 minutes ago
  },
  {
    id: '523e4567-e89b-12d3-a456-426614174004',
    displayName: 'TheLegend27',
    avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=TheLegend27',
    joinedAt: new Date(Date.now() - 1 * 60 * 1000).toISOString(), // 1 minute ago
  },
];

const meta = {
  title: 'Lobby/ParticipantList',
  component: ParticipantList,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    currentUserId: {
      control: 'text',
      description: 'ID of the current user (will be highlighted)',
    },
  },
} satisfies Meta<typeof ParticipantList>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Empty state - no participants yet
 */
export const Empty: Story = {
  args: {
    participants: [],
  },
};

/**
 * Single participant
 */
export const SingleParticipant: Story = {
  args: {
    participants: [mockParticipants[0]],
  },
};

/**
 * Multiple participants
 */
export const MultipleParticipants: Story = {
  args: {
    participants: mockParticipants,
  },
};

/**
 * Current user highlighted
 * The second participant is the current user
 */
export const WithCurrentUser: Story = {
  args: {
    participants: mockParticipants,
    currentUserId: mockParticipants[1].id,
  },
};

/**
 * Participant without avatar
 * Shows how the component handles missing avatars
 */
export const ParticipantWithoutAvatar: Story = {
  args: {
    participants: [mockParticipants[3]],
  },
};

/**
 * Large list of participants
 * Tests scrolling behavior
 */
export const LargeList: Story = {
  args: {
    participants: Array.from({ length: 20 }, (_, i) => ({
      id: `participant-${i}`,
      displayName: `Player${i + 1}`,
      avatarUrl: `https://api.dicebear.com/7.x/avataaars/svg?seed=Player${i + 1}`,
      joinedAt: new Date(Date.now() - i * 30 * 1000).toISOString(),
    })),
    currentUserId: 'participant-5',
  },
};

/**
 * Long display names
 * Tests text truncation
 */
export const LongDisplayNames: Story = {
  args: {
    participants: [
      {
        id: '1',
        displayName: 'VeryLongDisplayNameThatShouldBeTruncated123456',
        avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=1',
        joinedAt: new Date().toISOString(),
      },
      {
        id: '2',
        displayName: 'AnotherExtremelyLongNameForTesting',
        avatarUrl: 'https://api.dicebear.com/7.x/avataaars/svg?seed=2',
        joinedAt: new Date().toISOString(),
      },
    ],
  },
};

