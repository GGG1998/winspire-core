import { StrictMode } from 'react';
import { createRoot } from 'react-dom/client';
import './index.css';
import App from './App.tsx';

const envVars = import.meta.env as Record<string, string | boolean | undefined>;
const debugValue = envVars.VITE_DEBUG ?? envVars.DEBUG;
const isDebug =
  (typeof debugValue === 'string'
    ? debugValue.toLowerCase() === 'true'
    : Boolean(debugValue)) || false;

if (import.meta.env.DEV || isDebug) {
  import('./shared/dev-websocket')
    .then(({ setupDevConsole }) => setupDevConsole())
    .catch((error) => console.error('[WinspireDev] Failed to init dev console', error));
}

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <App />
  </StrictMode>,
);
