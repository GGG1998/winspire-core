/**
 * Tournament Feature Constants
 * Feature: 002-streamer-tournament-creation
 */

import type { TournamentStatus } from './types'

// ============================================================================
// Status Configuration
// ============================================================================

/**
 * Status display configuration with Polish labels and Tailwind colors
 */
export const TOURNAMENT_STATUS_CONFIG: Record<TournamentStatus, {
  label: string
  bgColor: string
  textColor: string
  borderColor: string
  rowBgColor: string
  badgeColor: 'cyan' | 'emerald' | 'zinc' | 'orange' | 'red'
}> = {
  scheduled: {
    label: 'Zaplanowany',
    bgColor: 'bg-orange-100 dark:bg-orange-900/30',
    textColor: 'text-orange-800 dark:text-orange-300',
    borderColor: 'border-orange-400 dark:border-orange-700',
    rowBgColor: 'bg-orange-50 dark:bg-orange-900/20',
    badgeColor: 'orange'
  },  
  draft: {
    label: 'Szkic',
    bgColor: 'bg-zinc-100 dark:bg-zinc-800',
    textColor: 'text-zinc-700 dark:text-zinc-300',
    borderColor: 'border-zinc-300 dark:border-zinc-700',
    rowBgColor: 'bg-zinc-50 dark:bg-zinc-800/50',
    badgeColor: 'zinc'
  },
  registration_open: {
    label: 'Otwarty',
    bgColor: 'bg-cyan-100 dark:bg-cyan-900/30',
    textColor: 'text-cyan-800 dark:text-cyan-300',
    borderColor: 'border-cyan-400 dark:border-cyan-700',
    rowBgColor: 'bg-cyan-50 dark:bg-cyan-900/20',
    badgeColor: 'cyan'
  },
  registration_closed: {
    label: 'Przed startem',
    bgColor: 'bg-orange-100 dark:bg-orange-900/30',
    textColor: 'text-orange-800 dark:text-orange-300',
    borderColor: 'border-orange-400 dark:border-orange-700',
    rowBgColor: 'bg-orange-50 dark:bg-orange-900/20',
    badgeColor: 'orange'
  },
  started: {
    label: 'Aktywny',
    bgColor: 'bg-emerald-100 dark:bg-emerald-900/30',
    textColor: 'text-emerald-800 dark:text-emerald-300',
    borderColor: 'border-emerald-400 dark:border-emerald-700',
    rowBgColor: 'bg-emerald-50 dark:bg-emerald-900/20',
    badgeColor: 'emerald'
  },
  completed: {
    label: 'Zakończony',
    bgColor: 'bg-zinc-100 dark:bg-zinc-800',
    textColor: 'text-zinc-700 dark:text-zinc-300',
    borderColor: 'border-zinc-300 dark:border-zinc-700',
    rowBgColor: 'bg-zinc-50 dark:bg-zinc-800/50',
    badgeColor: 'zinc'
  },
  cancelled: {
    label: 'Anulowany',
    bgColor: 'bg-red-100 dark:bg-red-900/30',
    textColor: 'text-red-800 dark:text-red-300',
    borderColor: 'border-red-400 dark:border-red-700',
    rowBgColor: 'bg-red-50 dark:bg-red-900/20',
    badgeColor: 'red'
  }
}

// ============================================================================
// Validation Constants
// ============================================================================

/**
 * Form validation rules
 */
export const TOURNAMENT_VALIDATION = {
  name: {
    minLength: 3,
    maxLength: 100
  },
  /** Inactivity timeout in hours */
  inactivityTimeoutHours: 1,
  /** Status update interval in milliseconds */
  statusUpdateIntervalMs: 30000
} as const

// ============================================================================
// Error Messages (Polish)
// ============================================================================

/**
 * Error messages for form validation
 */
export const VALIDATION_MESSAGES = {
  name: {
    required: 'Nazwa jest wymagana',
    minLength: `Nazwa musi mieć co najmniej ${TOURNAMENT_VALIDATION.name.minLength} znaki`,
    maxLength: `Nazwa nie może przekraczać ${TOURNAMENT_VALIDATION.name.maxLength} znaków`
  },
  startTime: {
    required: 'Czas rozpoczęcia jest wymagany',
    futureRequired: 'Czas rozpoczęcia musi być w przyszłości'
  },
  general: {
    network: 'Brak połączenia z internetem',
    validation: 'Dane formularza są nieprawidłowe',
    authorization: 'Nie masz uprawnień do tej operacji',
    server: 'Wystąpił błąd serwera. Spróbuj ponownie.',
    unknown: 'Wystąpił nieoczekiwany błąd'
  }
} as const

// ============================================================================
// UI Constants
// ============================================================================

/**
 * Default game value (currently fixed)
 */
export const DEFAULT_GAME = 'Packman'

/**
 * Default team mode value
 */
export const DEFAULT_TEAM_MODE = '1v1'

/**
 * Default max players value
 */
export const DEFAULT_MAX_PLAYERS = 250

/**
 * Available team mode options
 */
export const TEAM_MODE_OPTIONS = [
  { value: '1v1', label: '1v1' }
] as const

/**
 * UI text labels (Polish)
 */
export const UI_LABELS = {
  page: {
    title: 'Turnieje',
    subtitle: 'Zarządzaj swoimi turniejami',
    createButton: 'Stwórz turniej'
  },
  form: {
    nameLabel: 'Nazwa turnieju',
    namePlaceholder: 'Wprowadź nazwę turnieju',
    startTimeLabel: 'Czas rozpoczęcia',
    gameLabel: 'Gra',
    teamModeLabel: 'Tryb zespołu',
    maxPlayersLabel: 'Maksymalna liczba graczy',
    createButton: 'Utwórz turniej',
    cancelButton: 'Anuluj'
  },
  success: {
    title: 'Turniej utworzony!',
    goToRoom: 'Przejdź do pokoju',
    close: 'Zamknij',
    copyLink: 'Skopiuj link',
    copied: 'Skopiowano!'
  },
  table: {
    columns: {
      name: 'Nazwa',
      startTime: 'Czas rozpoczęcia',
      game: 'Gra',
      status: 'Status'
    },
    empty: 'Nie masz jeszcze żadnych turniejów. Kliknij \'Stwórz\', aby utworzyć swój pierwszy turniej.'
  },
  loading: 'Ładowanie...',
  error: {
    retry: 'Spróbuj ponownie'
  },
  // Tab labels for tournament detail page
  tabs: {
    overview: 'Przegląd',
    bracket: 'Drabinka',
    matches: 'Mecze',
    players: 'Gracze',
    results: 'Wyniki'
  },
  // Detail page labels
  detail: {
    startsIn: 'Start za',
    joinTournament: 'Dołącz do turnieju',
    tournamentInProgress: 'Turniej w toku',
    tournamentEnded: 'Turniej zakończony',
    backToTournaments: '← Wróć do turniejów',
    tournamentNotFound: 'Nie znaleziono turnieju',
    tournamentNotFoundDesc: 'Turniej, którego szukasz, nie istnieje.',
    loadingTournament: 'Ładowanie turnieju...'
  },
  // Format card labels
  format: {
    title: 'Format',
    game: 'Gra',
    readyWindow: 'Okno gotowości',
    teamSize: 'Rozmiar drużyny',
    customPrize: 'Nagroda',
    formatType: 'Format'
  },
  // Players panel labels
  players: {
    title: 'Gracze',
    registered: 'Zarejestrowani',
    ready: 'Gotowi',
    slots: 'Miejsca',
    slotsLeft: 'wolnych miejsc',
    noParticipants: 'Brak uczestników. Bądź pierwszy!',
    areRegistered: 'są zarejestrowani.',
    and: 'i',
    others: 'innych'
  },
  // Sidebar labels
  sidebar: {
    dashboard: 'Panel główny',
    tournaments: 'Turnieje',
    settings: 'Ustawienia',
    upcomingTournaments: 'Nadchodzące turnieje',
    support: 'Pomoc i wsparcie',
    whatsNew: 'Co nowego',
    search: 'Szukaj turniejów',
    accountSettings: 'Ustawienia konta',
    logout: 'Wyloguj',
    viewProfile: 'Zobacz profil',
    security: 'Bezpieczeństwo',
    sendFeedback: 'Prześlij opinię'
  },
  // Placeholder for upcoming tabs
  comingSoon: {
    title: 'Wkrótce',
    description: 'Ta sekcja jest w przygotowaniu.'
  },
  // Tournament settings/actions
  settings: {
    title: 'Ustawienia',
    edit: 'Edytuj turniej',
    publish: 'Opublikuj turniej',
    openRegistration: 'Otwórz rejestrację',
    start: 'Rozpocznij turniej',
    cancel: 'Anuluj turniej',
    confirmPublish: 'Czy na pewno chcesz opublikować turniej? Turniej będzie widoczny, ale rejestracja nie będzie jeszcze aktywna.',
    confirmOpenRegistration: 'Czy na pewno chcesz otworzyć rejestrację? Gracze będą mogli się zapisywać.',
    confirmStart: 'Czy na pewno chcesz rozpocząć turniej?',
    confirmCancel: 'Czy na pewno chcesz anulować turniej?',
    publishSuccess: 'Turniej został opublikowany',
    openRegistrationSuccess: 'Rejestracja została otwarta',
    startSuccess: 'Turniej został rozpoczęty',
    cancelSuccess: 'Turniej został anulowany',
    publishError: 'Nie udało się opublikować turnieju',
    openRegistrationError: 'Nie udało się otworzyć rejestracji',
    startError: 'Nie udało się rozpocząć turnieju',
    cancelError: 'Nie udało się anulować turnieju'
  }
} as const

// ============================================================================
// Tournament Format Labels (Polish)
// ============================================================================

export const FORMAT_TYPE_LABELS: Record<string, string> = {
  single_elimination: 'Single Elimination',
  double_elimination: 'Double Elimination',
  round_robin: 'Round Robin',
  swiss: 'System szwajcarski'
} as const

// ============================================================================
// Status Badge Configuration (Dark Theme)
// ============================================================================
// Removed: STATUS_BADGE_CONFIG - consolidated into TOURNAMENT_STATUS_CONFIG above

// ============================================================================
// Clipboard Constants
// ============================================================================

/**
 * Copy feedback display duration in milliseconds
 */
export const COPY_FEEDBACK_DURATION_MS = 2000

/**
 * Number of minutes before tournament start when join button becomes available
 */
export const JOIN_BUTTON_MINUTES_BEFORE_START = 5


