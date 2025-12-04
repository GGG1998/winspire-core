/**
 * GracePeriodIndicator Stories
 * Feature: 003-matchmaking-lobby-frontend
 */

import type { Meta, StoryObj } from '@storybook/react';
import { GracePeriodIndicator } from './GracePeriodIndicator';

const meta = {
  title: 'Lobby/GracePeriodIndicator',
  component: GracePeriodIndicator,
  parameters: {
    layout: 'padded',
  },
  tags: ['autodocs'],
  argTypes: {
    gracePeriodEndsAt: {
      control: 'text',
      description: 'ISO timestamp when grace period ends',
    },
    participantCount: {
      control: 'number',
      description: 'Current number of participants',
    },
    isUpdating: {
      control: 'boolean',
      description: 'Whether participant count is being updated',
    },
  },
} satisfies Meta<typeof GracePeriodIndicator>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * No grace period active
 * Component doesn't render when gracePeriodEndsAt is null
 */
export const NoGracePeriod: Story = {
  args: {
    gracePeriodEndsAt: null,
    participantCount: 8,
    isUpdating: false,
  },
};

/**
 * Grace period with 30 seconds remaining
 * Full countdown at the start
 */
export const FullCountdown: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 30 * 1000).toISOString(),
    participantCount: 8,
    isUpdating: false,
  },
};

/**
 * Grace period with 15 seconds remaining
 * Mid-way through the countdown
 */
export const MidCountdown: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 15 * 1000).toISOString(),
    participantCount: 10,
    isUpdating: false,
  },
};

/**
 * Grace period with 5 seconds remaining
 * Critical time - turns red and pulses
 */
export const LowTime: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 5 * 1000).toISOString(),
    participantCount: 12,
    isUpdating: false,
  },
  parameters: {
    docs: {
      description: {
        story: 'When less than 10 seconds remain, the countdown turns red and pulses to indicate urgency.',
      },
    },
  },
};

/**
 * Participant count updating
 * Shows the pulsing animation on participant count
 */
export const CountUpdating: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 20 * 1000).toISOString(),
    participantCount: 9,
    isUpdating: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'When a new participant joins or leaves, the count badge pulses with a green animation.',
      },
    },
  },
};

/**
 * Small participant count
 */
export const SmallCount: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 25 * 1000).toISOString(),
    participantCount: 2,
    isUpdating: false,
  },
};

/**
 * Large participant count
 */
export const LargeCount: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 25 * 1000).toISOString(),
    participantCount: 64,
    isUpdating: false,
  },
};

/**
 * Almost expired
 * 2 seconds remaining
 */
export const AlmostExpired: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 2 * 1000).toISOString(),
    participantCount: 8,
    isUpdating: false,
  },
};

/**
 * With participant update during low time
 * Combines urgency and update animations
 */
export const LowTimeWithUpdate: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 7 * 1000).toISOString(),
    participantCount: 11,
    isUpdating: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'Shows both the red countdown (low time) and the green pulsing participant badge (updating).',
      },
    },
  },
};

/**
 * Progress bar states
 * Shows different stages of the progress bar
 */
export const ProgressBarStart: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 30 * 1000).toISOString(),
    participantCount: 8,
    isUpdating: false,
  },
  name: 'Progress Bar - Start (0%)',
};

export const ProgressBarMiddle: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 15 * 1000).toISOString(),
    participantCount: 8,
    isUpdating: false,
  },
  name: 'Progress Bar - Middle (50%)',
};

export const ProgressBarAlmostDone: Story = {
  args: {
    gracePeriodEndsAt: new Date(Date.now() + 3 * 1000).toISOString(),
    participantCount: 8,
    isUpdating: false,
  },
  name: 'Progress Bar - Almost Done (90%)',
};

