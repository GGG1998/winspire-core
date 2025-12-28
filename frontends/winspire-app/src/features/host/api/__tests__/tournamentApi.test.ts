/**
 * Tournament API Tests
 * Feature: 002-streamer-tournament-creation
 */

import { describe, it, expect } from 'vitest'
import { formToApiInput, apiToUiTournament } from '../tournamentApi'
import type { TournamentFormData, TournamentApiData } from '../../types'

describe('formToApiInput', () => {
  it('should convert form data to API input format with all fields', () => {
    const formData: TournamentFormData = {
      name: '  Test Tournament  ',
      startTime: '2024-12-31T20:00',
      game: 'Packman',
      gameSnapshot: { id: 'game-1', name: 'Packman', slug: 'packman', version: '1.0.0' },
      teamMode: '1v1',
      minPlayers: 2,
      maxPlayers: 16
    }

    const result = formToApiInput(formData)

    expect(result.name).toBe('Test Tournament') // trimmed
    expect(result.scheduledStartTimeAt).toBe('2024-12-31T20:00:00.000Z')
    expect(result.competitionParameters.tournamentParameters?.teamSize).toBe(1)
    expect(result.minimumTeamCount).toBe(16)
    expect(result.game?.slug).toBe('packman')
  })

  it('should parse teamMode correctly for different formats', () => {
    const testCases = [
      { teamMode: '1v1', expectedTeamSize: 1 },
      { teamMode: '2v2', expectedTeamSize: 2 },
      { teamMode: '3v3', expectedTeamSize: 3 },
      { teamMode: '5v5', expectedTeamSize: 5 }
    ]

    testCases.forEach(({ teamMode, expectedTeamSize }) => {
      const formData: TournamentFormData = {
        name: 'Test',
        startTime: '2024-12-31T20:00',
        game: 'Packman',
        gameSnapshot: { id: 'game-1', name: 'Packman', slug: 'packman', version: '1.0.0' },
        teamMode,
        minPlayers: 2,
        maxPlayers: 16
      }

      const result = formToApiInput(formData)
      expect(result.competitionParameters.tournamentParameters?.teamSize).toBe(expectedTeamSize)
    })
  })

  it('should handle missing optional fields', () => {
    const formData: TournamentFormData = {
      name: 'Test Tournament',
      startTime: '',
      game: '',
      gameSnapshot: { id: '', name: '', slug: '', version: '' },
      teamMode: '1v1',
      minPlayers: 2,
      maxPlayers: 8
    }

    const result = formToApiInput(formData)

    expect(result.name).toBe('Test Tournament')
    expect(result.scheduledStartTimeAt).toBeUndefined()
    expect(result.game).toBeUndefined()
    expect(result.minimumTeamCount).toBe(8)
  })

  it('should trim whitespace from tournament name', () => {
    const formData: TournamentFormData = {
      name: '   Spaced Tournament   ',
      startTime: '2024-12-31T20:00',
      game: 'Packman',
      gameSnapshot: { id: 'game-1', name: 'Packman', slug: 'packman', version: '1.0.0' },
      teamMode: '1v1',
      minPlayers: 2,
      maxPlayers: 16
    }

    const result = formToApiInput(formData)
    expect(result.name).toBe('Spaced Tournament')
  })

  it('should convert game name to lowercase slug', () => {
    const formData: TournamentFormData = {
      name: 'Test',
      startTime: '2024-12-31T20:00',
      game: 'PACKMAN',
      gameSnapshot: { id: 'game-1', name: 'PACKMAN', slug: 'packman', version: '1.0.0' },
      teamMode: '1v1',
      minPlayers: 2,
      maxPlayers: 16
    }

    const result = formToApiInput(formData)
    expect(result.game?.slug).toBe('packman')
  })
})

describe('apiToUiTournament', () => {
  it('should convert API data to UI tournament model', () => {
    const apiData: TournamentApiData = {
      id: '123e4567-e89b-12d3-a456-426614174000',
      name: 'Test Tournament',
      status: 'scheduled',
      scheduledStartTimeAt: '2024-12-31T20:00:00Z',
      game: 'Packman',
      gameLogoUrl: 'https://example.com/logo.png',
      bannerUrl: 'https://example.com/banner.png',
      creatorId: '123e4567-e89b-12d3-a456-426614174001',
      roomLink: '/tournament/123',
      isCompleted: false,
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T12:00:00Z'
    }

    const result = apiToUiTournament(apiData)

    expect(result.id).toBe(apiData.id)
    expect(result.name).toBe(apiData.name)
    expect(result.startTime).toBeInstanceOf(Date)
    expect(result.startTime.toISOString()).toBe('2024-12-31T20:00:00.000Z')
    expect(result.game).toBe('Packman')
    expect(result.gameLogoUrl).toBe(apiData.gameLogoUrl)
    expect(result.creatorId).toBe(apiData.creatorId)
    expect(result.roomLink).toBe('/tournament/123')
    expect(result.isCompleted).toBe(false)
    expect(result.createdAt).toBeInstanceOf(Date)
    expect(result.updatedAt).toBeInstanceOf(Date)
  })

  it('should handle optional organizer field', () => {
    const apiData: TournamentApiData = {
      id: '123',
      name: 'Test',
      status: 'scheduled',
      scheduledStartTimeAt: '2024-12-31T20:00:00Z',
      game: 'Packman',
      creatorId: '456',
      roomLink: '/tournament/123',
      isCompleted: false,
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
      organizer: {
        id: 'org-1',
        name: 'Test Org',
        logoUrl: 'https://example.com/org-logo.png'
      }
    }

    const result = apiToUiTournament(apiData)
    expect(result.organizer).toBeDefined()
    expect(result.organizer?.name).toBe('Test Org')
  })

  it('should handle optional format field', () => {
    const apiData: TournamentApiData = {
      id: '123',
      name: 'Test',
      status: 'scheduled',
      scheduledStartTimeAt: '2024-12-31T20:00:00Z',
      game: 'Packman',
      creatorId: '456',
      roomLink: '/tournament/123',
      isCompleted: false,
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
      format: {
        type: 'single_elimination',
        teamSize: 1,
        maxSlots: 16,
        bestOf: 3
      }
    }

    const result = apiToUiTournament(apiData)
    expect(result.format).toBeDefined()
    expect(result.format?.type).toBe('single_elimination')
    expect(result.format?.teamSize).toBe(1)
  })

  it('should convert readyWindow dates to Date objects', () => {
    const apiData: TournamentApiData = {
      id: '123',
      name: 'Test',
      status: 'scheduled',
      scheduledStartTimeAt: '2024-12-31T20:00:00Z',
      game: 'Packman',
      creatorId: '456',
      roomLink: '/tournament/123',
      isCompleted: false,
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
      readyWindow: {
        startsAt: '2024-12-31T19:00:00Z',
        endsAt: '2024-12-31T19:30:00Z'
      }
    }

    const result = apiToUiTournament(apiData)
    expect(result.readyWindow).toBeDefined()
    expect(result.readyWindow?.startsAt).toBeInstanceOf(Date)
    expect(result.readyWindow?.endsAt).toBeInstanceOf(Date)
  })

  it('should handle participantCount field', () => {
    const apiData: TournamentApiData = {
      id: '123',
      name: 'Test',
      status: 'scheduled',
      scheduledStartTimeAt: '2024-12-31T20:00:00Z',
      game: 'Packman',
      creatorId: '456',
      roomLink: '/tournament/123',
      isCompleted: false,
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
      participantCount: 12
    }

    const result = apiToUiTournament(apiData)
    expect(result.participantCount).toBe(12)
  })

  it('should handle lastActivity date conversion', () => {
    const apiData: TournamentApiData = {
      id: '123',
      name: 'Test',
      status: 'scheduled',
      scheduledStartTimeAt: '2024-12-31T20:00:00Z',
      game: 'Packman',
      creatorId: '456',
      roomLink: '/tournament/123',
      isCompleted: false,
      createdAt: '2024-01-01T10:00:00Z',
      updatedAt: '2024-01-01T10:00:00Z',
      lastActivity: '2024-12-31T15:00:00Z'
    }

    const result = apiToUiTournament(apiData)
    expect(result.lastActivity).toBeInstanceOf(Date)
    expect(result.lastActivity?.toISOString()).toBe('2024-12-31T15:00:00.000Z')
  })
})

