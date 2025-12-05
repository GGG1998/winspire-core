/**
 * LobbyLayout Stories
 * Feature: 003-matchmaking-lobby-frontend
 */

import type { Meta, StoryObj } from '@storybook/react';
import { MemoryRouter } from 'react-router-dom';
import { LobbyLayout } from './LobbyLayout';
import { AuthProvider } from '../../auth/context/AuthContext';

const meta = {
  title: 'Lobby/LobbyLayout',
  component: LobbyLayout,
  decorators: [
    (Story) => (
      <MemoryRouter initialEntries={['/tournaments/123/lobby']}>
        <AuthProvider>
          <Story />
        </AuthProvider>
      </MemoryRouter>
    ),
  ],
  parameters: {
    layout: 'fullscreen',
  },
  tags: ['autodocs'],
} satisfies Meta<typeof LobbyLayout>;

export default meta;
type Story = StoryObj<typeof meta>;

/**
 * Default layout (without tournament context)
 * Shows the lobby layout with standard navigation
 */
export const Default: Story = {
  render: () => (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
          Strona z Layout (standardowa nawigacja)
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          To jest przykładowa zawartość w LobbyLayout. Layout zawiera sidebar z nawigacją dla graczy:
          Strona główna, Dostępne turnieje, Moje mecze, Rankingi.
        </p>
        
        <div className="mt-8 grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Karta 1
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Przykładowa zawartość karty
            </p>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Karta 2
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Przykładowa zawartość karty
            </p>
          </div>
          <div className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6">
            <h3 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
              Karta 3
            </h3>
            <p className="text-sm text-gray-600 dark:text-gray-400">
              Przykładowa zawartość karty
            </p>
          </div>
        </div>
      </div>
    </LobbyLayout>
  ),
};

/**
 * With tournament context
 * Shows "Powrót do Podglądu turnieju" link in sidebar
 */
export const WithTournamentContext: Story = {
  render: () => (
    <LobbyLayout tournamentId="test-tournament-123" streamerId="streamer-456">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
          Strona z Layout (kontekst turnieju)
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mb-4">
          Gdy podamy <code className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded">tournamentId</code> i{' '}
          <code className="px-2 py-1 bg-gray-100 dark:bg-gray-800 rounded">streamerId</code>,
          sidebar pokazuje link "Powrót do Podglądu turnieju" zamiast standardowej nawigacji.
        </p>
        
        <div className="mt-8 bg-cyan-50 dark:bg-cyan-900/20 border-2 border-cyan-200 dark:border-cyan-800 rounded-lg p-6">
          <h2 className="text-lg font-semibold text-cyan-900 dark:text-cyan-100 mb-2">
            ℹ️ Kontekst turnieju aktywny
          </h2>
          <ul className="text-sm text-cyan-700 dark:text-cyan-300 space-y-1">
            <li>• Tournament ID: <code>test-tournament-123</code></li>
            <li>• Streamer ID: <code>streamer-456</code></li>
            <li>• Link powrotu: <code>/h/streamer-456/tournaments/test-tournament-123</code></li>
          </ul>
        </div>
      </div>
    </LobbyLayout>
  ),
};

/**
 * With long content
 * Tests scrolling behavior
 */
export const WithLongContent: Story = {
  render: () => (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
          Długa Zawartość
        </h1>
        {Array.from({ length: 20 }, (_, i) => (
          <div key={i} className="mb-6">
            <h2 className="text-xl font-semibold text-gray-900 dark:text-white mb-2">
              Sekcja {i + 1}
            </h2>
            <p className="text-gray-600 dark:text-gray-400">
              Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor 
              incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud 
              exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.
            </p>
          </div>
        ))}
      </div>
    </LobbyLayout>
  ),
};

/**
 * Empty content
 * Minimal content in layout
 */
export const EmptyContent: Story = {
  render: () => (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <div className="text-center py-12">
          <h1 className="text-2xl font-semibold text-gray-900 dark:text-white mb-2">
            Brak zawartości
          </h1>
          <p className="text-gray-600 dark:text-gray-400">
            Ta strona jest pusta
          </p>
        </div>
      </div>
    </LobbyLayout>
  ),
};

/**
 * With full-width content
 * Content that spans the full width
 */
export const FullWidthContent: Story = {
  render: () => (
    <LobbyLayout>
      <div className="px-4 sm:px-6 lg:px-8 py-8">
        <div className="bg-gradient-to-r from-cyan-500 to-blue-500 rounded-lg p-8 text-white">
          <h1 className="text-3xl font-bold mb-4">
            Pełna Szerokość
          </h1>
          <p className="text-lg">
            Ta zawartość rozciąga się na całą szerokość dostępnego obszaru.
          </p>
        </div>
        
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4 mt-8">
          {Array.from({ length: 8 }, (_, i) => (
            <div 
              key={i} 
              className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
            >
              <h3 className="font-semibold text-gray-900 dark:text-white mb-2">
                Item {i + 1}
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Zawartość
              </p>
            </div>
          ))}
        </div>
      </div>
    </LobbyLayout>
  ),
};

/**
 * Mobile view
 * Layout on mobile devices
 */
export const MobileView: Story = {
  render: () => (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
          Widok Mobilny
        </h1>
        <p className="text-gray-600 dark:text-gray-400 mb-6">
          Layout automatycznie dostosowuje się do mniejszych ekranów. 
          Sidebar staje się hamburger menu.
        </p>
        
        <div className="space-y-4">
          {Array.from({ length: 5 }, (_, i) => (
            <div 
              key={i}
              className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-4"
            >
              <h3 className="font-semibold text-gray-900 dark:text-white mb-1">
                Element {i + 1}
              </h3>
              <p className="text-sm text-gray-600 dark:text-gray-400">
                Responsywna zawartość
              </p>
            </div>
          ))}
        </div>
      </div>
    </LobbyLayout>
  ),
  parameters: {
    viewport: {
      defaultViewport: 'mobile1',
    },
  },
};

/**
 * With loading content
 * Shows loading spinners in layout
 */
export const WithLoadingContent: Story = {
  render: () => (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-4">
          Ładowanie...
        </h1>
        
        <div className="space-y-4">
          {Array.from({ length: 3 }, (_, i) => (
            <div 
              key={i}
              className="bg-white dark:bg-gray-800 rounded-lg border border-gray-200 dark:border-gray-700 p-6"
            >
              <div className="animate-pulse">
                <div className="h-4 bg-gray-200 dark:bg-gray-700 rounded w-3/4 mb-2" />
                <div className="h-3 bg-gray-200 dark:bg-gray-700 rounded w-1/2" />
              </div>
            </div>
          ))}
        </div>
      </div>
    </LobbyLayout>
  ),
};

/**
 * With alerts and notifications
 * Shows various UI states
 */
export const WithAlerts: Story = {
  render: () => (
    <LobbyLayout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        <h1 className="text-3xl font-bold text-gray-900 dark:text-white mb-6">
          Powiadomienia
        </h1>
        
        {/* Success alert */}
        <div className="mb-4 rounded-lg border-2 border-green-300 dark:border-green-800 bg-green-50 dark:bg-green-900/20 p-4">
          <h3 className="text-lg font-semibold text-green-900 dark:text-green-100 mb-1">
            Sukces!
          </h3>
          <p className="text-sm text-green-700 dark:text-green-300">
            Operacja została wykonana pomyślnie.
          </p>
        </div>
        
        {/* Warning alert */}
        <div className="mb-4 rounded-lg border-2 border-yellow-300 dark:border-yellow-800 bg-yellow-50 dark:bg-yellow-900/20 p-4">
          <h3 className="text-lg font-semibold text-yellow-900 dark:text-yellow-100 mb-1">
            Ostrzeżenie
          </h3>
          <p className="text-sm text-yellow-700 dark:text-yellow-300">
            Coś wymaga Twojej uwagi.
          </p>
        </div>
        
        {/* Error alert */}
        <div className="mb-4 rounded-lg border-2 border-red-300 dark:border-red-800 bg-red-50 dark:bg-red-900/20 p-4">
          <h3 className="text-lg font-semibold text-red-900 dark:text-red-100 mb-1">
            Błąd
          </h3>
          <p className="text-sm text-red-700 dark:text-red-300">
            Wystąpił problem z wykonaniem operacji.
          </p>
        </div>
        
        {/* Info alert */}
        <div className="mb-4 rounded-lg border-2 border-blue-300 dark:border-blue-800 bg-blue-50 dark:bg-blue-900/20 p-4">
          <h3 className="text-lg font-semibold text-blue-900 dark:text-blue-100 mb-1">
            Informacja
          </h3>
          <p className="text-sm text-blue-700 dark:text-blue-300">
            Oto przydatna informacja dla Ciebie.
          </p>
        </div>
      </div>
    </LobbyLayout>
  ),
};

