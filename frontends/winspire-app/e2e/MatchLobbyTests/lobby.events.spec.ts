import { test, expect, BrowserContext } from '@playwright/test';
import { getStreamer, getUser, type SeededAccount } from '../fixtures/auth.fixture';
import { seedMatchLobby1v1 } from './fixture';

import type { Page } from '@playwright/test';

/** 
 * Set up authenticated session for a page.
 * Must be called AFTER creating the page but BEFORE navigating to protected routes.
 */
async function setupAuth(page: Page, account: SeededAccount) {
  // Navigate to home page first
  await page.goto('/');
  
  // Inject session into localStorage
  // NOTE: Supabase expects the session directly, not wrapped in an object
  await page.evaluate((session) => {
    window.localStorage.setItem('winspire-auth', JSON.stringify(session));
  }, account.session);
  
  // Reload to pick up session
  await page.reload({ waitUntil: 'networkidle' });
  await page.waitForTimeout(500);
}

test.describe('Match Lobby websocket readiness', () => {
  test('user1 cannot ready until opponent appears', async ({ browser }) => {
    const streamer = await getStreamer(1);
    const seed = await seedMatchLobby1v1({ participant1: streamer, attachSecond: false });

    const context1 = await browser.newContext();
    const page1 = await context1.newPage();
    
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Should show waiting placeholder for opponent and no ready button
    await expect(page1.getByRole('heading', { name: 'Czekanie na przeciwnika' })).toBeVisible();
    await expect(page1.getByRole('button', { name: 'Jestem gotowy' })).toHaveCount(0);

    await context1.close();
  });

  test('both players see lobby and ready button is available', async ({ browser }) => {
    const streamer = await getStreamer(1);
    const user = await getUser(1);
    const seed = await seedMatchLobby1v1({ participant1: streamer, participant2: user, attachSecond: true });

    // Context for participant1 (streamer)
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Context for participant2 (user)
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Both pages should show the lobby with VS indicator (both players present)
    await expect(page1.getByText('VS')).toBeVisible({ timeout: 10000 });
    await expect(page2.getByText('VS')).toBeVisible({ timeout: 10000 });

    // Both should see "Jestem gotowy!" button when both players are present
    await expect(page1.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });
    await expect(page2.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  test('clicking ready changes button to cancel ready', async ({ browser }) => {
    const streamer = await getStreamer(2);
    const user = await getUser(2);
    const seed = await seedMatchLobby1v1({ participant1: streamer, participant2: user, attachSecond: true });

    // Context for both players
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Wait for lobby to be ready
    await expect(page1.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });
    await expect(page2.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });

    // Player 1 clicks ready - button should change to "Anuluj gotowość"
    await page1.getByRole('button', { name: /Jestem gotowy/i }).click();
    await expect(page1.getByRole('button', { name: /Anuluj gotowość/i })).toBeVisible({ timeout: 5000 });

    // Player 2 clicks ready - this triggers both players ready
    // IMPORTANT: When both players are ready, the backend immediately sends match_ready_to_load
    // which changes status to 'loading' and hides the ReadyButton.
    // So we need to check for either the button OR the loading state.
    await page2.getByRole('button', { name: /Jestem gotowy/i }).click();
    
    // After player 2 clicks, expect either:
    // 1. Button changed to "Anuluj gotowość" (if API response arrives before match_ready_to_load)
    // 2. OR loading state appears (if match_ready_to_load arrives first - which is the common case)
    const cancelButton = page2.getByRole('button', { name: /Anuluj gotowość/i });
    const loadingText = page2.getByText(/Ładowanie gry|Obaj gracze gotowi/i);
    
    await expect(cancelButton.or(loadingText)).toBeVisible({ timeout: 5000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  // Test WebSocket event reception - requires matchmaking backend
  test('WebSocket receives ready_updated event from opponent (requires backend)', async ({ browser }) => {
    // Use different users than previous tests to avoid conflicts
    const streamer = await getStreamer(4);
    const user = await getUser(4);
    const seed = await seedMatchLobby1v1({ participant1: streamer, participant2: user, attachSecond: true });

    // Context for player 1 (streamer)
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();

    // Capture console logs for WebSocket debugging
    const wsMessages: { type: string; payload?: unknown }[] = [];
    page1.on('console', (msg) => {
      const text = msg.text();
      if (text.includes('[useMatchLobby] Received WebSocket message:')) {
        const match = text.match(/message: (\w+)/);
        if (match) {
          wsMessages.push({ type: match[1] });
        }
      }
    });

    // Monitor API responses
    const apiResponses: { url: string; status: number }[] = [];
    page1.on('response', (res) => {
      if (res.url().includes('/ready') || res.url().includes('/lobby')) {
        apiResponses.push({ url: res.url(), status: res.status() });
      }
    });

    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Context for player 2 (user) - will click ready
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    
    // Monitor ready API on page2
    let readyApiStatus: number | null = null;
    page2.on('response', (res) => {
      if (res.url().includes('/ready')) {
        readyApiStatus = res.status();
        console.log('[E2E] Ready API response:', res.status());
      }
    });

    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Wait for lobby to load
    await expect(page1.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });
    await expect(page2.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });

    // Player 2 clicks ready
    await page2.getByRole('button', { name: /Jestem gotowy/i }).click();
    
    // Wait for button change (local UI update)
    await expect(page2.getByRole('button', { name: /Anuluj gotowość/i })).toBeVisible({ timeout: 5000 });

    // Give WebSocket time to deliver the message
    await page1.waitForTimeout(3000);

    // Check what WebSocket messages were received
    console.log('[E2E] WebSocket messages captured:', wsMessages);
    console.log('[E2E] API responses captured:', apiResponses);
    console.log('[E2E] Ready API status:', readyApiStatus);

    // Find ready_updated message
    const readyUpdatedMsg = wsMessages.find((m) => m.type === 'ready_updated');
    
    // Assert we received the event
    expect(readyUpdatedMsg, 
      `Expected 'ready_updated' WebSocket event.\n` +
      `WS messages: ${wsMessages.map(m => m.type).join(', ') || 'none'}\n` +
      `Ready API status: ${readyApiStatus || 'N/A'}\n` +
      `Ensure matchmaking backend is running with WebSocket broadcast.`
    ).toBeDefined();

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  // Full flow test - requires working backend
  test.skip('both players ready triggers countdown (requires backend)', async ({ browser }) => {
    const streamer = await getStreamer(5);
    const user = await getUser(5);
    const seed = await seedMatchLobby1v1({ participant1: streamer, participant2: user, attachSecond: true });

    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    await expect(page1.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });
    await expect(page2.getByRole('button', { name: /Jestem gotowy/i })).toBeVisible({ timeout: 10000 });

    // Both click ready
    await page1.getByRole('button', { name: /Jestem gotowy/i }).click();
    await page2.getByRole('button', { name: /Jestem gotowy/i }).click();

    // With backend: both should see "Gotowy" badge on player cards
    await expect(page1.locator('.text-green-500', { hasText: 'Gotowy' })).toHaveCount(2, { timeout: 10000 });
    await expect(page2.locator('.text-green-500', { hasText: 'Gotowy' })).toHaveCount(2, { timeout: 10000 });

    // Countdown overlay should appear
    await expect(page1.getByRole('heading', { name: /Mecz rozpoczyna się/i })).toBeVisible({ timeout: 10000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });
});
