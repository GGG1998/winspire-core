import { render, screen } from '@testing-library/react'
import { TournamentHeader } from '../components/TournamentHeader'
import type { Tournament } from '../types'

function buildTournament(overrides: Partial<Tournament> = {}): Tournament {
  const base: Tournament = {
    id: 't-1',
    name: 'Test',
    status: 'started',
    startTime: new Date(),
    game: 'Packman',
    creatorId: 'host-1',
    roomLink: '',
    isCompleted: false,
    createdAt: new Date(),
    updatedAt: new Date(),
    me: undefined,
  }
  return { ...base, ...overrides }
}

describe('TournamentHeader return to match', () => {
  it('shows return button when me.match exists', () => {
    const tournament = buildTournament({
      me: {
        participationStatus: 'checked_in',
        match: {
          matchId: 'm-123',
          tournamentId: 't-1',
          status: 'started',
          round: 1,
          table: 2,
          startedAt: new Date(),
          updatedAt: new Date(),
        },
      },
    })

    render(
      <TournamentHeader
        tournament={tournament}
        status={tournament.status}
        isOwner
      />
    )

    expect(screen.getByText(/Wróć do meczu/i)).toBeInTheDocument()
  })

  it('hides return button when me.match missing', () => {
    const tournament = buildTournament({
      me: {
        participationStatus: 'checked_in',
      },
    })

    render(
      <TournamentHeader
        tournament={tournament}
        status={tournament.status}
        isOwner
      />
    )

    expect(screen.queryByText(/Wróć do meczu/i)).toBeNull()
  })
})


