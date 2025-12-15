type DevEventDefinition = {
  name: string;
  eventType: string;
  description: string;
  samplePayload: Record<string, unknown>;
};

type DevCommandHandler = (
  payload?: Record<string, unknown>,
  metadata?: Record<string, string>
) => Promise<unknown>;

interface WinMatchOptions {
  matchId?: string;
  loserId?: string;
  scoreWinner?: number;
  scoreLoser?: number;
}

interface DevConsole {
  endpoint: string;
  commands: Record<string, DevCommandHandler>;
  sendEvent: (
    eventType: string,
    payload?: Record<string, unknown>,
    metadata?: Record<string, string>
  ) => Promise<unknown>;
  listCommands: () => DevEventDefinition[];
  getAccessToken: () => string | null;
  refreshToken: () => string | null;
  getUserId: () => string | null;
  /** Complete current match with yourself as winner. Extracts matchId from URL if not provided. */
  winMatch: (options?: WinMatchOptions) => Promise<unknown>;
  /** Get the current match ID from URL */
  getMatchIdFromUrl: () => string | null;
}

declare global {
  interface Window {
    WinspireDev?: DevConsole;
  }
}

const STORAGE_KEY = 'winspire-auth';
const DEFAULT_EVENTS_PATH = '/v1/matchmaking/dev/events';

const EVENT_DEFINITIONS: DevEventDefinition[] = [
  {
    name: 'tournamentStartRequested',
    eventType: 'TournamentStartRequested',
    description: 'Rozpoczyna turniej i uruchamia grace period.',
    samplePayload: {
      tournament_id: '11111111-1111-1111-1111-111111111111',
      host_id: '22222222-2222-2222-2222-222222222222',
      participants: [
        'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
      ],
      game_id: '33333333-3333-3333-3333-333333333333',
      game_snapshot: {
        id: '44444444-4444-4444-4444-444444444444',
        slug: 'game-slug',
        name: 'Sample Game',
        version: '1.0.0',
      },
      started_at: new Date().toISOString(),
    },
  },
  {
    name: 'bracketGenerated',
    eventType: 'BracketGenerated',
    description: 'Publikuje informacje o wygenerowanej drabince (round 1).',
    samplePayload: {
      tournament_id: '11111111-1111-1111-1111-111111111111',
    },
  },
  {
    name: 'matchCreated',
    eventType: 'MatchCreated',
    description: 'Tworzy mecz w kolejnych rundach i wysyła match_assigned.',
    samplePayload: {
      tournament_id: '11111111-1111-1111-1111-111111111111',
      match_id: '55555555-5555-5555-5555-555555555555',
      round_number: 2,
      match_number: 1,
      participant1_id: 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa',
      participant2_id: 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb',
      is_bye: false,
    },
  },
  {
    name: 'matchStarted',
    eventType: 'MatchStarted',
    description: 'Informuje o starcie meczu i generuje gameUrl.',
    samplePayload: {
      match_id: '55555555-5555-5555-5555-555555555555',
      tournament_id: '11111111-1111-1111-1111-111111111111',
    },
  },
];

const eventEndpoint: string = (() => {
  const env = import.meta.env as Record<string, string | boolean | undefined>;
  const explicit = env.VITE_MATCHMAKING_DEV_EVENTS_ENDPOINT;
  if (typeof explicit === 'string' && explicit) {
    return explicit;
  }

  const base = (env.VITE_API_BASE_URL as string | undefined) || '';
  if (!base) {
    return DEFAULT_EVENTS_PATH;
  }

  const normalizedBase = base.endsWith('/') ? base.slice(0, -1) : base;
  return `${normalizedBase}${DEFAULT_EVENTS_PATH.startsWith('/') ? '' : '/'}${DEFAULT_EVENTS_PATH}`;
})();

const defaultMetadata = {
  correlation_id: 'dev-console',
};

function getUserIdFromToken(token: string | null): string | null {
  if (!token) return null;
  try {
    const parts = token.split('.');
    if (parts.length !== 3) return null;
    const payload = JSON.parse(atob(parts[1]));
    return payload.sub || payload.user_id || null;
  } catch {
    return null;
  }
}

function getMatchIdFromUrl(): string | null {
  const url = window.location.pathname;
  // Match patterns like /lobby/match/:matchId or /match/:matchId
  const patterns = [
    /\/lobby\/match\/([a-f0-9-]{36})/i,
    /\/match\/([a-f0-9-]{36})/i,
    /\/matches\/([a-f0-9-]{36})/i,
  ];

  for (const pattern of patterns) {
    const match = url.match(pattern);
    if (match) {
      return match[1];
    }
  }

  // Also check URL search params
  const params = new URLSearchParams(window.location.search);
  const matchId = params.get('matchId') || params.get('match_id');
  if (matchId && /^[a-f0-9-]{36}$/i.test(matchId)) {
    return matchId;
  }

  return null;
}

function getApiBaseUrl(): string {
  const env = import.meta.env as Record<string, string | boolean | undefined>;
  const base = (env.VITE_API_BASE_URL as string | undefined) || '';
  return base.endsWith('/') ? base.slice(0, -1) : base;
}

function getAccessTokenFromStorage(): string | null {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (!raw) {
      return null;
    }

    const parsed = JSON.parse(raw);
    if (typeof parsed === 'string') {
      return parsed;
    }

    if (parsed?.access_token && typeof parsed.access_token === 'string') {
      return parsed.access_token;
    }

    if (parsed?.currentSession?.access_token && typeof parsed.currentSession.access_token === 'string') {
      return parsed.currentSession.access_token;
    }

    if (parsed?.session?.access_token && typeof parsed.session.access_token === 'string') {
      return parsed.session.access_token;
    }

    return null;
  } catch (error) {
    console.warn('[WinspireDev] Failed to read token from storage:', error);
    return null;
  }
}

function createCommandHandlers(
  send: DevConsole['sendEvent'],
): Record<string, DevCommandHandler> {
  return EVENT_DEFINITIONS.reduce<Record<string, DevCommandHandler>>((handlers, definition) => {
    handlers[definition.name] = async (
      payload?: Record<string, unknown>,
      metadata?: Record<string, string>,
    ) => {
      const mergedPayload = {
        ...definition.samplePayload,
        ...(payload || {}),
      };

      return send(definition.eventType, mergedPayload, metadata);
    };
    return handlers;
  }, {});
}

function createDevConsole(): DevConsole {
  let cachedToken: string | null = null;

  const sendEvent = async (
    eventType: string,
    payload: Record<string, unknown> = {},
    metadata: Record<string, string> = defaultMetadata,
  ) => {
    const token = cachedToken ?? getAccessTokenFromStorage();
    cachedToken = token;

    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
    };

    if (token) {
      headers.Authorization = `Bearer ${token}`;
    } else {
      console.warn('[WinspireDev] access_token not found, sending unauthenticated request.');
    }

    const response = await fetch(eventEndpoint, {
      method: 'POST',
      headers,
      body: JSON.stringify({
        event_type: eventType,
        payload,
        metadata,
      }),
    });

    const contentType = response.headers.get('content-type') || '';
    const parser = contentType.includes('application/json')
      ? () => response.json()
      : () => response.text();

    const responseBody = await parser().catch(() => null);

    if (!response.ok) {
      throw new Error(`[WinspireDev] Request failed (${response.status}): ${JSON.stringify(responseBody)}`);
    }

    return responseBody;
  };

  const getToken = () => cachedToken ?? getAccessTokenFromStorage();

  const winMatch = async (options: WinMatchOptions = {}): Promise<unknown> => {
    const token = getToken();
    if (!token) {
      throw new Error('[WinspireDev] No auth token found. Please log in first.');
    }

    const userId = getUserIdFromToken(token);
    if (!userId) {
      throw new Error('[WinspireDev] Cannot extract user ID from token.');
    }

    const matchId = options.matchId || getMatchIdFromUrl();
    if (!matchId) {
      throw new Error('[WinspireDev] No matchId provided and could not extract from URL. Provide matchId option or navigate to a match page.');
    }

    const apiBase = getApiBaseUrl();
    const endpoint = `${apiBase}/v1/matchmaking/matches/${matchId}/complete`;

    console.info(`[WinspireDev] Completing match ${matchId} with winner ${userId}`);

    const response = await fetch(endpoint, {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        winner_id: userId,
        loser_id: options.loserId || '00000000-0000-0000-0000-000000000000', // Placeholder if not provided
        score_winner: options.scoreWinner ?? 3,
        score_loser: options.scoreLoser ?? 0,
        result_source: 'manual_host',
      }),
    });

    const contentType = response.headers.get('content-type') || '';
    const parser = contentType.includes('application/json')
      ? () => response.json()
      : () => response.text();

    const responseBody = await parser().catch(() => null);

    if (!response.ok) {
      throw new Error(`[WinspireDev] Failed to complete match (${response.status}): ${JSON.stringify(responseBody)}`);
    }

    console.info('[WinspireDev] Match completed successfully!', responseBody);
    return responseBody;
  };

  return {
    endpoint: eventEndpoint,
    commands: createCommandHandlers(sendEvent),
    sendEvent,
    listCommands: () => EVENT_DEFINITIONS,
    getAccessToken: getToken,
    refreshToken: () => {
      cachedToken = getAccessTokenFromStorage();
      return cachedToken;
    },
    getUserId: () => getUserIdFromToken(getToken()),
    winMatch,
    getMatchIdFromUrl,
  };
}

export function setupDevConsole(): void {
  if (typeof window === 'undefined') {
    return;
  }

  if (window.WinspireDev) {
    return;
  }

  window.WinspireDev = createDevConsole();

  console.info('[WinspireDev] Dev console ready.');
  console.info('[WinspireDev] Quick actions:');
  console.info('  WinspireDev.winMatch()         - Complete current match as winner (extracts matchId from URL)');
  console.info('  WinspireDev.getUserId()        - Get your user ID from token');
  console.info('  WinspireDev.getMatchIdFromUrl() - Get match ID from current URL');
  console.info('');
  console.info('[WinspireDev] Event commands:');
  console.table(
    EVENT_DEFINITIONS.map(({ name, eventType, description }) => ({
      command: `WinspireDev.commands.${name}(payload?)`,
      eventType,
      description,
    })),
  );
  console.info('[WinspireDev] Use WinspireDev.listCommands() to inspect sample payloads.');
}
