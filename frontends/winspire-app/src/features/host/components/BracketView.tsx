/**
 * BracketView Component
 * Feature: 002-streamer-tournament-creation
 * 
 * Displays tournament bracket/elimination tree
 * Currently a placeholder - will be implemented when bracket data is available
 */

import { UI_LABELS } from '../constants'
import type { Tournament } from '../types'

interface BracketViewProps {
  tournament: Tournament
}

export function BracketView({ tournament }: BracketViewProps) {
  const labels = UI_LABELS.comingSoon
  const formatType = tournament.format?.type || 'single_elimination'
  
  // Format type label
  const formatLabels: Record<string, string> = {
    single_elimination: 'Single Elimination',
    double_elimination: 'Double Elimination',
    round_robin: 'Round Robin',
    swiss: 'System szwajcarski'
  }

  return (
    <div className="py-8">
      {/* Placeholder content */}
      <div className="text-center max-w-md mx-auto">
        <div className="w-20 h-20 mx-auto mb-6 rounded-2xl bg-zinc-100 dark:bg-zinc-800 flex items-center justify-center">
          <svg className="w-10 h-10 text-zinc-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
            <path
              strokeLinecap="round"
              strokeLinejoin="round"
              strokeWidth={1.5}
              d="M4 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1V5zM4 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1H5a1 1 0 01-1-1v-4zM14 5a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1V5zM14 15a1 1 0 011-1h4a1 1 0 011 1v4a1 1 0 01-1 1h-4a1 1 0 01-1-1v-4z"
            />
          </svg>
        </div>
        
        <h3 className="text-xl font-semibold text-zinc-950 dark:text-white mb-2">
          Drabinka turniejowa
        </h3>
        
        <p className="text-zinc-500 dark:text-zinc-400 mb-6">
          {labels.description}
        </p>

        {/* Format info card */}
        <div className="rounded-xl border border-zinc-200 dark:border-zinc-700/50 bg-zinc-50 dark:bg-zinc-800/50 p-4">
          <div className="text-sm text-zinc-500 dark:text-zinc-400 mb-1">Format turnieju</div>
          <div className="text-lg font-medium text-zinc-950 dark:text-white">
            {formatLabels[formatType] || formatType}
          </div>
          {tournament.format?.maxSlots && (
            <div className="text-sm text-zinc-500 dark:text-zinc-400 mt-1">
              {tournament.format.maxSlots} uczestników
            </div>
          )}
        </div>

        {/* Bracket visualization placeholder */}
        <div className="mt-8 grid grid-cols-3 gap-4 opacity-30">
          {/* Round 1 */}
          <div className="space-y-4">
            <div className="text-xs text-zinc-400 uppercase tracking-wide mb-2">Runda 1</div>
            {[1, 2, 3].map(i => (
              <div key={i} className="h-16 rounded-lg border-2 border-dashed border-zinc-300 dark:border-zinc-600" />
            ))}
          </div>
          
          {/* Round 2 */}
          <div className="space-y-4 pt-10">
            <div className="text-xs text-zinc-400 uppercase tracking-wide mb-2">Półfinały</div>
            {[1, 2].map(i => (
              <div key={i} className="h-16 rounded-lg border-2 border-dashed border-zinc-300 dark:border-zinc-600" />
            ))}
          </div>
          
          {/* Finals */}
          <div className="space-y-4 pt-20">
            <div className="text-xs text-zinc-400 uppercase tracking-wide mb-2">Finał</div>
            <div className="h-16 rounded-lg border-2 border-dashed border-zinc-300 dark:border-zinc-600" />
          </div>
        </div>
      </div>
    </div>
  )
}

