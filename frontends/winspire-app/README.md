# Winspire App

A React + TypeScript + Vite application for the Winspire platform, featuring user and streamer authentication with social login support.

## Features

- **User Authentication**: Email/password and Google OAuth login for users
- **Streamer Authentication**: Email/password, Twitch, and Discord OAuth login for streamers
- **Profile Management**: Automatic profile creation and management
- **Lobby System**: Real-time tournament lobbies and matchmaking

## Prerequisites

- Node.js (v18 or higher)
- Supabase project configured with OAuth providers

## Environment Variables

Create a `.env` file in the root of this directory with the following variables:

```env
VITE_SUPABASE_URL=your_supabase_project_url
VITE_SUPABASE_ANON_KEY=your_supabase_anon_key
```

## OAuth Setup

This application uses Supabase Auth for social login. The following OAuth providers are configured:

### For Users
- **Google OAuth**: Users can sign in with their Google accounts

### For Streamers
- **Twitch OAuth**: Streamers can sign in with their Twitch accounts
- **Discord OAuth**: Streamers can sign in with their Discord accounts

### Setting Up OAuth Providers

OAuth providers are configured in the Supabase config (`platform/supabase/config.toml`). To set up OAuth providers:

1. **Google OAuth**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Create a new project or select an existing one
   - Enable the Google+ API
   - Go to "Credentials" and create OAuth 2.0 Client ID
   - Add authorized redirect URI: `http://127.0.0.1:54321/auth/v1/callback`
   - Copy the Client ID and Client Secret
   - Set environment variables:
     ```bash
     export SUPABASE_AUTH_EXTERNAL_GOOGLE_CLIENT_ID="your_client_id"
     export SUPABASE_AUTH_EXTERNAL_GOOGLE_SECRET="your_client_secret"
     ```

2. **Twitch OAuth**:
   - Go to [Twitch Developer Console](https://dev.twitch.tv/console/apps)
   - Register a new application
   - Add OAuth Redirect URL: `http://127.0.0.1:54321/auth/v1/callback`
   - Copy the Client ID and Client Secret
   - Set environment variables:
     ```bash
     export SUPABASE_AUTH_EXTERNAL_TWITCH_CLIENT_ID="your_client_id"
     export SUPABASE_AUTH_EXTERNAL_TWITCH_SECRET="your_client_secret"
     ```

3. **Discord OAuth**:
   - Go to [Discord Developer Portal](https://discord.com/developers/applications)
   - Create a new application
   - Go to "OAuth2" section
   - Add redirect URL: `http://127.0.0.1:54321/auth/v1/callback`
   - Copy the Client ID and Client Secret
   - Set environment variables:
     ```bash
     export SUPABASE_AUTH_EXTERNAL_DISCORD_CLIENT_ID="your_client_id"
     export SUPABASE_AUTH_EXTERNAL_DISCORD_SECRET="your_client_secret"
     ```

### How OAuth Works

1. When a user clicks on a social login button, they are redirected to the OAuth provider
2. After successful authentication, the provider redirects back to `/auth/callback`
3. The application automatically determines the profile type:
   - Google login → creates a User profile
   - Twitch/Discord login → creates a Streamer profile
4. Account linking is automatic when the same email is used across providers

## Development

Install dependencies:

```bash
npm install
```

Start the development server:

```bash
npm run dev
```

## Building for Production

```bash
npm run build
```

## React + TypeScript + Vite

This template provides a minimal setup to get React working in Vite with HMR and some ESLint rules.

Currently, two official plugins are available:

- [@vitejs/plugin-react](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react) uses [Babel](https://babeljs.io/) (or [oxc](https://oxc.rs) when used in [rolldown-vite](https://vite.dev/guide/rolldown)) for Fast Refresh
- [@vitejs/plugin-react-swc](https://github.com/vitejs/vite-plugin-react/blob/main/packages/plugin-react-swc) uses [SWC](https://swc.rs/) for Fast Refresh

## React Compiler

The React Compiler is not enabled on this template because of its impact on dev & build performances. To add it, see [this documentation](https://react.dev/learn/react-compiler/installation).

## Expanding the ESLint configuration

If you are developing a production application, we recommend updating the configuration to enable type-aware lint rules:

```js
export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...

      // Remove tseslint.configs.recommended and replace with this
      tseslint.configs.recommendedTypeChecked,
      // Alternatively, use this for stricter rules
      tseslint.configs.strictTypeChecked,
      // Optionally, add this for stylistic rules
      tseslint.configs.stylisticTypeChecked,

      // Other configs...
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```

You can also install [eslint-plugin-react-x](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-x) and [eslint-plugin-react-dom](https://github.com/Rel1cx/eslint-react/tree/main/packages/plugins/eslint-plugin-react-dom) for React-specific lint rules:

```js
// eslint.config.js
import reactX from 'eslint-plugin-react-x'
import reactDom from 'eslint-plugin-react-dom'

export default defineConfig([
  globalIgnores(['dist']),
  {
    files: ['**/*.{ts,tsx}'],
    extends: [
      // Other configs...
      // Enable lint rules for React
      reactX.configs['recommended-typescript'],
      // Enable lint rules for React DOM
      reactDom.configs.recommended,
    ],
    languageOptions: {
      parserOptions: {
        project: ['./tsconfig.node.json', './tsconfig.app.json'],
        tsconfigRootDir: import.meta.dirname,
      },
      // other options...
    },
  },
])
```
