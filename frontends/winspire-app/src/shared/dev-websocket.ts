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

const eventEndpoint = (() => {
  const env = import.meta.env as Record<string, string | boolean | undefined>;
  const explicit = env.VITE_MATCHMAKING_DEV_EVENTS_ENDPOINT;
  if (explicit) {
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

  return {
    endpoint: eventEndpoint,
    commands: createCommandHandlers(sendEvent),
    sendEvent,
    listCommands: () => EVENT_DEFINITIONS,
    getAccessToken: () => cachedToken ?? getAccessTokenFromStorage(),
    refreshToken: () => {
      cachedToken = getAccessTokenFromStorage();
      return cachedToken;
    },
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
  console.table(
    EVENT_DEFINITIONS.map(({ name, eventType, description }) => ({
      command: `WinspireDev.commands.${name}(payload?)`,
      eventType,
      description,
    })),
  );
  console.info('[WinspireDev] Use WinspireDev.listCommands() to inspect sample payloads.');
}
