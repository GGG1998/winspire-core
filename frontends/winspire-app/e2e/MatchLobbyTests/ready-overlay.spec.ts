import { test, expect } from '@playwright/test';
import { getStreamer, getUser, type SeededAccount } from '../fixtures/auth.fixture';
import { seedMatchLobby1v1 } from './fixture';
import { seedGames } from '../fixtures/games.fixture';

import type { Page } from '@playwright/test';

// Game slug for the PacMan game seeded in fixtures
const PACMAN_GAME_SLUG = 'pacmanbuildfixed';

/**
 * Set up authenticated session for a page.
 * Must be called AFTER creating the page but BEFORE navigating to protected routes.
 */
async function setupAuth(page: Page, account: SeededAccount) {
  // Navigate to home page first
  await page.goto('/');

  // Inject session into localStorage
  await page.evaluate((session) => {
    window.localStorage.setItem('winspire-auth', JSON.stringify(session));
  }, account.session);

  // Reload to pick up session
  await page.reload({ waitUntil: 'networkidle' });
  await page.waitForTimeout(500);
}

test.describe('Ready Button Overlay Flow', () => {
  // Use streamer 5 and user 11-15 to avoid conflicts with other tests
  // Streamers: 1-5, Users: 1-20

  // Seed games before running tests (only needs to run once)
  test.beforeAll(async ({ request }) => {
    await seedGames(request);
  });

  test('GameFrame appears when both players join in pending status', async ({ browser }) => {
    const streamer = await getStreamer(5);
    const user = await getUser(11);
    const seed = await seedMatchLobby1v1({
      participant1: streamer,
      participant2: user,
      attachSecond: true,
      gameSlug: PACMAN_GAME_SLUG,
    });

    // Setup player 1
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Setup player 2
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Both should see VS indicator (both players present)
    await expect(page1.getByText('VS')).toBeVisible({ timeout: 10000 });
    await expect(page2.getByText('VS')).toBeVisible({ timeout: 10000 });

    // GameFrame (iframe) should be visible - this is the new behavior
    // Look for the game iframe container with aspect-ratio 16:9
    const gameFrame1 = page1.locator('iframe[title="Game"]');
    const gameFrame2 = page2.locator('iframe[title="Game"]');

    await expect(gameFrame1).toBeVisible({ timeout: 15000 });
    await expect(gameFrame2).toBeVisible({ timeout: 15000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  test('Ready button overlay appears after game loads', async ({ browser }) => {
    const streamer = await getStreamer(5);
    const user = await getUser(12);
    const seed = await seedMatchLobby1v1({
      participant1: streamer,
      participant2: user,
      attachSecond: true,
      gameSlug: PACMAN_GAME_SLUG,
    });

    // Setup player 1
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Setup player 2
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Wait for game to load (may take a few seconds)
    // The Ready button should appear as overlay AFTER game loads
    await expect(
      page1.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 20000 });
    await expect(
      page2.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 20000 });

    // The overlay text should be visible
    await expect(
      page1.getByText('Gra załadowana! Kliknij gdy będziesz gotowy.')
    ).toBeVisible({ timeout: 5000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  test('clicking Ready shows waiting for opponent overlay', async ({ browser }) => {
    const streamer = await getStreamer(5);
    const user = await getUser(13);
    const seed = await seedMatchLobby1v1({
      participant1: streamer,
      participant2: user,
      attachSecond: true,
      gameSlug: PACMAN_GAME_SLUG,
    });

    // Setup player 1
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Setup player 2
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Wait for Ready button overlay to appear (after game loads)
    await expect(
      page1.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 20000 });

    // Player 1 clicks Ready
    await page1.getByRole('button', { name: /Jestem gotowy/i }).click();

    // Player 1 should see "Waiting for opponent" overlay
    await expect(page1.getByText('Jesteś gotowy!')).toBeVisible({ timeout: 5000 });
    await expect(page1.getByText('Oczekiwanie na przeciwnika...')).toBeVisible({
      timeout: 5000,
    });

    // Player 2 should still see the Ready button
    await expect(
      page2.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 5000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  test('both players click Ready - overlays disappear and loading state begins', async ({
    browser,
  }) => {
    const streamer = await getStreamer(5);
    const user = await getUser(14);
    const seed = await seedMatchLobby1v1({
      participant1: streamer,
      participant2: user,
      attachSecond: true,
      gameSlug: PACMAN_GAME_SLUG,
    });

    // Setup player 1
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Setup player 2
    const ctx2 = await browser.newContext();
    const page2 = await ctx2.newPage();
    await setupAuth(page2, user);
    await page2.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Wait for Ready button overlay to appear (after game loads)
    await expect(
      page1.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 20000 });
    await expect(
      page2.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 20000 });

    // Both players click Ready
    await page1.getByRole('button', { name: /Jestem gotowy/i }).click();
    await page2.getByRole('button', { name: /Jestem gotowy/i }).click();

    // After both ready, status should change to "loading" or game loading state
    // The ready overlay should disappear (showReadyOverlay becomes false when status !== 'pending')
    await expect(page1.getByText('Jesteś gotowy!')).not.toBeVisible({ timeout: 10000 });
    await expect(page1.getByText('Oczekiwanie na przeciwnika...')).not.toBeVisible({
      timeout: 5000,
    });

    // Either loading status text or game running state
    const loadingOrStarted = page1
      .getByText(/Ładowanie gry|Status: W trakcie/i)
      .or(page1.getByRole('heading', { name: /Status: Ładowanie gry/i }));

    await expect(loadingOrStarted).toBeVisible({ timeout: 10000 });

    await Promise.all([ctx1.close(), ctx2.close()]);
  });

  test('game iframe is visible behind Ready overlay', async ({ browser }) => {
    const streamer = await getStreamer(5);
    const user = await getUser(15);
    const seed = await seedMatchLobby1v1({
      participant1: streamer,
      participant2: user,
      attachSecond: true,
      gameSlug: PACMAN_GAME_SLUG,
    });

    // Setup player 1
    const ctx1 = await browser.newContext();
    const page1 = await ctx1.newPage();
    await setupAuth(page1, streamer);
    await page1.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

    // Wait for Ready button to appear (means game has loaded)
    await expect(
      page1.getByRole('button', { name: /Jestem gotowy/i })
    ).toBeVisible({ timeout: 20000 });

    // The iframe should still be visible (behind the overlay)
    const gameFrame = page1.locator('iframe[title="Game"]');
    await expect(gameFrame).toBeVisible();

    // The overlay has semi-transparent background (bg-black/60)
    // So the game is visible behind it
    const overlay = page1.locator('.bg-black\\/60');
    await expect(overlay).toBeVisible();

    await ctx1.close();
  });
});
