/**
 * CreateTournamentModal Stories
 * Feature: 002-streamer-tournament-creation
 */

import type { Meta, StoryObj } from '@storybook/react'
import { fn } from '@storybook/test'
import { CreateTournamentModal } from './CreateTournamentModal'
import { TournamentsProvider } from '../contexts/TournamentsContext'

const meta = {
  title: 'Host/CreateTournamentModal',
  component: CreateTournamentModal,
  decorators: [
    (Story) => (
      <TournamentsProvider>
        <Story />
      </TournamentsProvider>
    ),
  ],
  parameters: {
    layout: 'centered',
    // Mock API calls at the fetch level to prevent errors
    mockData: [
      {
        url: '/api/tournaments',
        method: 'GET',
        status: 200,
        response: { tournaments: [] },
      },
    ],
  },
  tags: ['autodocs'],
  argTypes: {
    isOpen: {
      control: 'boolean',
      description: 'Controls modal visibility',
    },
  },
  args: {
    isOpen: false,
    onClose: fn(),
    onSuccess: fn(),
  },
} satisfies Meta<typeof CreateTournamentModal>

export default meta
type Story = StoryObj<typeof meta>

/**
 * Default state - modal closed
 */
export const Closed: Story = {
  args: {
    isOpen: false,
  },
}

/**
 * Modal open with empty form
 */
export const Open: Story = {
  args: {
    isOpen: true,
  },
}

/**
 * Modal with validation errors
 * 
 * To see validation:
 * 1. Click "Utwórz turniej" button without filling the form
 * 2. Try entering a name shorter than 3 characters
 */
export const WithValidationErrors: Story = {
  args: {
    isOpen: true,
  },
  play: async ({ canvasElement: _canvasElement }) => {
    // This story is interactive - user can trigger validation errors
    // by interacting with the form
  },
}

/**
 * Modal in submitting state (simulated)
 * 
 * Note: The actual submitting state is managed internally by the component.
 * This story demonstrates the visual appearance when the form is being submitted.
 */
export const Submitting: Story = {
  args: {
    isOpen: true,
  },
  parameters: {
    docs: {
      description: {
        story: 'When the form is being submitted, the submit button shows a loading spinner and is disabled.',
      },
    },
  },
}
