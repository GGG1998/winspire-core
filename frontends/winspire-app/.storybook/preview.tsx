import type { Preview } from "@storybook/react";
import type { ReactRenderer } from "@storybook/react";
import type { DecoratorFunction } from "@storybook/types";
import { AuthProvider } from "../src/features/auth/context/AuthContext";
import "../src/index.css";

// Mock Supabase before any imports
// This needs to happen before the actual supabase client is created
if (typeof window !== 'undefined') {
  // Store the original fetch
  const originalFetch = window.fetch;

  // Override fetch to mock API calls
  window.fetch = async (input: RequestInfo | URL, init?: RequestInit) => {
    const url = typeof input === 'string' ? input : input instanceof URL ? input.href : input.url;
    
    console.log('Storybook fetch intercept:', url);
    
    // Mock Supabase auth endpoints
    if (url.includes('supabase') && url.includes('auth')) {
      return new Response(
        JSON.stringify({
          access_token: 'storybook-mock-token',
          user: {
            id: 'storybook-mock-user-id',
            email: 'storybook@mock.com',
          },
        }),
        { 
          status: 200, 
          headers: { 'Content-Type': 'application/json' } 
        }
      );
    }
    
    // Mock tournament API GET requests
    if (url.includes('/tournaments') && (!init?.method || init?.method === 'GET')) {
      return new Response(
        JSON.stringify({ tournaments: [] }),
        { 
          status: 200, 
          headers: { 'Content-Type': 'application/json' } 
        }
      );
    }
    
    // Mock tournament API POST requests
    if (url.includes('/tournaments') && init?.method === 'POST') {
      const body = init.body ? JSON.parse(init.body as string) : {};
      return new Response(
        JSON.stringify({
          id: `storybook-mock-${Date.now()}`,
          ...body,
          status: 'DRAFT',
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
          hostId: 'storybook-mock-host-id',
        }),
        { 
          status: 201, 
          headers: { 'Content-Type': 'application/json' } 
        }
      );
    }
    
    // For all other requests, use original fetch
    return originalFetch(input, init);
  };
}

// Global decorator to wrap all stories with AuthProvider
const withAuthProvider: DecoratorFunction<ReactRenderer> = (Story) => (
  <AuthProvider>
    <Story />
  </AuthProvider>
);

const preview: Preview = {
  decorators: [withAuthProvider],
  parameters: {
    controls: {
      matchers: {
        color: /(background|color)$/i,
        date: /Date$/i,
      },
    },
  },
};

export default preview;
