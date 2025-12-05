/**
 * Authentication Fixtures for E2E Tests
 * 
 * Provides pre-authenticated browser contexts with cached sessions.
 * This avoids slow UI registration/login flows in every test.
 */

import { test as base, Page, BrowserContext } from '@playwright/test';
import { 
  createUser, 
  createStreamer, 
  generateTestData,
  TestUser 
} from '../helpers/api-client';

export interface AuthFixtures {
  // Single authenticated player
  authenticatedPlayer: { page: Page; user: TestUser };
  
  // Single authenticated streamer
  authenticatedStreamer: { page: Page; user: TestUser };
  
  // Multiple authenticated players (for multi-user tests)
  fourAuthenticatedPlayers: Array<{ page: Page; user: TestUser; context: BrowserContext }>;
}

export const test = base.extend<AuthFixtures>({
  /**
   * Fixture: Single authenticated player
   * Use for tests that need one logged-in player
   */
  authenticatedPlayer: async ({ browser, request }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    
    // Create user via API (fast!)
    const testData = generateTestData('player');
    const user = await createUser(
      request,
      testData.nickname,
      testData.email,
      testData.password
    );
    
    // Set auth state (bypass login UI)
    await context.addCookies([
      {
        name: 'sb-access-token',
        value: user.accessToken,
        domain: 'localhost',
        path: '/',
      }
    ]);
    
    // Navigate to verify auth works
    await page.goto('/');
    await page.waitForTimeout(500);
    
    await use({ page, user });
    
    await context.close();
  },

  /**
   * Fixture: Single authenticated streamer
   * Use for tests that need one logged-in streamer
   */
  authenticatedStreamer: async ({ browser, request }, use) => {
    const context = await browser.newContext();
    const page = await context.newPage();
    
    const testData = generateTestData('streamer');
    const user = await createStreamer(
      request,
      testData.nickname,
      testData.email,
      testData.password,
      `Channel_${testData.nickname}`
    );
    
    await context.addCookies([
      {
        name: 'sb-access-token',
        value: user.accessToken,
        domain: 'localhost',
        path: '/',
      }
    ]);
    
    await page.goto('/');
    await page.waitForTimeout(500);
    
    await use({ page, user });
    
    await context.close();
  },

  /**
   * Fixture: Four authenticated players
   * Use for multi-user tests (like pre-lobby)
   */
  fourAuthenticatedPlayers: async ({ browser, request }, use) => {
    type PlayerData = { page: Page; user: TestUser; context: BrowserContext };
    const players: PlayerData[] = [];
    
    // Create 4 players in parallel (much faster!)
    const playerPromises: Promise<PlayerData>[] = [];
    for (let i = 1; i <= 4; i++) {
      playerPromises.push((async (): Promise<PlayerData> => {
        const context = await browser.newContext();
        const page = await context.newPage();
        
        const testData = generateTestData(`player${i}`);
        const user = await createUser(
          request,
          testData.nickname,
          testData.email,
          testData.password
        );
        
        await context.addCookies([
          {
            name: 'sb-access-token',
            value: user.accessToken,
            domain: 'localhost',
            path: '/',
          }
        ]);
        
        await page.goto('/');
        await page.waitForTimeout(500);
        
        return { page, user, context };
      })());
    }
    
    const createdPlayers = await Promise.all(playerPromises);
    players.push(...createdPlayers);
    
    await use(players);
    
    // Cleanup
    for (const player of players) {
      await player.context.close();
    }
  },
});

export { expect } from '@playwright/test';

