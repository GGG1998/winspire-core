/**
 * TournamentTabs Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Navigation tabs for tournament detail views (Polish labels)
 */

import clsx from 'clsx'
import type { TournamentTab } from '../types'
import { UI_LABELS } from '../constants'

interface TournamentTabsProps {
  activeTab: TournamentTab
  onTabChange: (tab: TournamentTab) => void
}

const TABS: Array<{ id: TournamentTab; labelKey: keyof typeof UI_LABELS.tabs }> = [
  { id: 'overview', labelKey: 'overview' },
  { id: 'bracket', labelKey: 'bracket' },
  { id: 'matches', labelKey: 'matches' },
  { id: 'players', labelKey: 'players' },
  { id: 'results', labelKey: 'results' }
]

export function TournamentTabs({ activeTab, onTabChange }: TournamentTabsProps) {
  return (
    <div className="border-b border-zinc-950/5 dark:border-white/5">
      <nav className="-mb-px flex gap-6" aria-label="Tabs">
        {TABS.map((tab) => (
          <button
            key={tab.id}
            onClick={() => onTabChange(tab.id)}
            className={clsx(
              'py-4 px-1 text-sm font-medium border-b-2 transition-colors whitespace-nowrap',
              activeTab === tab.id
                ? 'border-cyan-500 text-cyan-600 dark:text-cyan-400'
                : 'border-transparent text-zinc-500 hover:text-zinc-700 hover:border-zinc-300 dark:text-zinc-400 dark:hover:text-zinc-200 dark:hover:border-zinc-600'
            )}
          >
            {UI_LABELS.tabs[tab.labelKey]}
          </button>
        ))}
      </nav>
    </div>
  )
}

