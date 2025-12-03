# Winspire Frontends

Frontend aplikacji Winspire Platform - SPA zbudowana w React z Vite.

## Struktura

```
frontends/
├── README.md              # Ten plik
└── winspire-app/          # Główna aplikacja React
    ├── src/
    │   ├── features/      # Feature modules (self-contained)
    │   │   ├── auth/      # Auth feature
    │   │   ├── host/      # Host Platform feature (future)
    │   │   └── lobby/     # Tournament Lobby feature
    │   ├── shared/        # Shared code across features
    │   ├── App.tsx        # Main app component
    │   └── main.tsx       # Entry point
    └── package.json
```

## Tech Stack

- **Framework**: React 18+ z Vite
- **UI Library**: Catalyst UI (@i4o/catalystui)
- **Styling**: Tailwind CSS
- **TypeScript**: Type safety
- **Routing**: React Router v6
- **State Management**: React Context API
- **Form Handling**: React Hook Form
- **SSE**: EventSource API dla real-time updates

## Quick Start

```bash
cd winspire-app
npm install
npm run dev
```

## Features

### Auth Feature
- Logowanie i rejestracja
- Zarządzanie profilem użytkownika
- Protected routes

### Tournament Lobby Feature
- Widok lobby turniejowego
- Real-time updates przez SSE
- Matchmaking queue
- Czat w lobby

### Host Feature (Future)
- Zarządzanie turniejami
- Dashboard hosta
- Ustawienia turnieju

## Architektura

Projekt używa **feature-based architecture** - każdy feature jest self-contained i zawiera:
- `components/` - Komponenty specyficzne dla feature
- `pages/` - Strony feature
- `layouts/` - Layouty specyficzne dla feature (jeśli potrzebne)
- `hooks/` - Custom hooks
- `api/` - API client
- `types.ts` - TypeScript types
- `index.ts` - Public exports

Wspólny kod znajduje się w `shared/`.

## Development

Zobacz dokumentację w `docs-site/docs/` dla szczegółowych specyfikacji.






