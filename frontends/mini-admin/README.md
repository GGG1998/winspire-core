# Mini Admin Frontend

React + Vite frontend for managing game files uploaded to S3.

## Features

- **Upload Game Tab:**
  - Choose between new game or existing game
  - Form with name, slug, version, description, logo
  - Optional S3 versioning toggle (with cost warning)
  - Drag & drop file upload zone
  - Multiple file selection

- **Games List Tab:**
  - Table view of all games
  - Edit button to update game details
  - Delete button (soft delete)
  - Copy S3 URL button
  - Status badges (active/inactive, versioning enabled/disabled)

## Quick Start

### Install Dependencies

```bash
npm install
```

### Configure Environment

Create a `.env` file:

```bash
VITE_API_URL=http://localhost:8088/v1
```

### Development

```bash
npm run dev
```

Runs on `http://localhost:3001`

### Build

```bash
npm run build
```

Output is in `dist/` directory.

### Preview Production Build

```bash
npm run preview
```

## Deployment to AWS Amplify

1. **Connect Repository:**
   - Go to AWS Amplify Console
   - Connect your Git repository
   - Select the `frontends/mini-admin` directory as the root

2. **Build Settings:**
   - Amplify will automatically detect `amplify.yml`
   - Set environment variable: `VITE_API_URL`

3. **Deploy:**
   - Push to your repository
   - Amplify will auto-build and deploy

## Project Structure

```
frontends/mini-admin/
├── public/              # Static assets
├── src/
│   ├── api/            # API client and endpoints
│   │   ├── client.ts   # Axios configuration
│   │   └── games.ts    # Games API methods
│   ├── features/games/ # Game-related components
│   │   ├── UploadGame.tsx
│   │   ├── UploadGame.css
│   │   ├── GamesList.tsx
│   │   └── GamesList.css
│   ├── App.tsx         # Main app with tabs
│   ├── App.css
│   ├── main.tsx        # Entry point
│   └── index.css       # Global styles
├── amplify.yml         # AWS Amplify build spec
├── index.html          # HTML template (required as per spec)
├── package.json
├── tsconfig.json
└── vite.config.ts
```

## Tech Stack

- **React 18**: UI library
- **TypeScript**: Type safety
- **Vite**: Build tool
- **Axios**: HTTP client
- **CSS**: Styling (no framework, pure CSS)

## API Integration

The frontend communicates with the backend API using Axios. All API calls are defined in `src/api/games.ts`:

```typescript
// Example usage
import { gamesApi } from './api/games'

// Get all games
const games = await gamesApi.getAllGames()

// Create a game
const newGame = await gamesApi.createGame({
  slug: 'my-game',
  name: 'My Game',
  version: '1.0.0',
  versioningEnabled: false,
})

// Upload files
await gamesApi.uploadFiles(gameId, files)
```

## Environment Variables

- `VITE_API_URL`: Backend API base URL (default: `http://localhost:8088/v1`)

## Development Tips

1. **Hot Reload:** Vite provides instant hot module replacement
2. **TypeScript:** All API types are defined in `src/api/games.ts`
3. **Error Handling:** API errors are caught and displayed to users
4. **CORS:** Ensure backend has CORS configured for your frontend URL

## License

MIT










