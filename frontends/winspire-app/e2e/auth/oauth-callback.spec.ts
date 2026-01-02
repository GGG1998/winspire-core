import { test, expect } from '@playwright/test';
import { getUser, getStreamer, type SeededAccount } from '../fixtures/auth.fixture';
import type { Page } from '@playwright/test';

/**
 * Set up authenticated session for a page by injecting session into localStorage.
 * Simulates what happens after successful OAuth code exchange.
 */
async function setupAuth(page: Page, account: SeededAccount) {
  await page.goto('/');
  await page.evaluate((session) => {
    window.localStorage.setItem('winspire-auth', JSON.stringify(session));
  }, account.session);
  await page.reload({ waitUntil: 'networkidle' });
  await page.waitForTimeout(500);
}

/**
 * Clear any existing session from localStorage.
 */
async function clearAuth(page: Page) {
  await page.goto('/');
  await page.evaluate(() => {
    window.localStorage.removeItem('winspire-auth');
  });
}

test.describe('OAuth Callback Page', () => {
  test.describe('Loading State', () => {
    test('shows loading state when navigating to callback', async ({ page }) => {
      // Navigate to callback without any session or code
      await page.goto('/auth/callback');

      // Should show the loading spinner and message
      await expect(page.getByRole('heading', { name: /completing sign in/i })).toBeVisible();
      await expect(page.getByText(/please wait while we complete your authentication/i)).toBeVisible();
    });
  });

  test.describe('User Profile (simulated Google OAuth)', () => {
    test('redirects to home when profile is complete', async ({ page }) => {
      // Get a user account with complete profile
      const user = await getUser(1);

      // Set up the authenticated session
      await setupAuth(page, user);

      // Navigate to callback - should detect existing session and redirect
      await page.goto('/auth/callback');

      // Should redirect to home page
      await page.waitForURL('/', { timeout: 10000 });
      expect(page.url()).toMatch(/\/$/);
    });

    test('redirects to /auth/user/login on timeout without session', async ({ page }) => {
      // Clear any existing session
      await clearAuth(page);

      // Navigate to callback without session or code
      await page.goto('/auth/callback');

      // Should show loading initially
      await expect(page.getByRole('heading', { name: /completing sign in/i })).toBeVisible();

      // Wait for timeout (15s) + redirect delay (3s) with some buffer
      await page.waitForURL(/\/auth\/user\/login/, { timeout: 25000 });

      // Should be on user login page (default for unknown/Google provider)
      expect(page.url()).toContain('/auth/user/login');
    });
  });

  test.describe('Streamer Profile (simulated Twitch/Discord OAuth)', () => {
    test('redirects to home when profile is complete', async ({ page }) => {
      // Get a streamer account with complete profile
      const streamer = await getStreamer(1);

      // Set up the authenticated session
      await setupAuth(page, streamer);

      // Navigate to callback - should detect existing session and redirect
      await page.goto('/auth/callback');

      // Should redirect to home page
      await page.waitForURL('/', { timeout: 10000 });
      expect(page.url()).toMatch(/\/$/);
    });
  });

  test.describe('Error Handling', () => {
    test('shows error message before redirect', async ({ page }) => {
      // Navigate to callback with an invalid code parameter
      // This will cause exchangeCodeForSession to fail
      await page.goto('/auth/callback?code=invalid-code-that-will-fail');

      // Should show error message
      await expect(page.getByRole('heading', { name: /authentication error/i })).toBeVisible({ timeout: 10000 });
      await expect(page.getByText(/redirecting to login page/i)).toBeVisible();

      // Should redirect to login after delay
      await page.waitForURL(/\/auth\/.*\/login/, { timeout: 10000 });
    });
  });

  test.describe('Profile Completion Flow', () => {
    // Note: Testing the incomplete profile flow requires creating a user without a nickname,
    // which is complex because the database trigger auto-creates profiles with the signup data.
    // For complete E2E coverage, this would require either:
    // 1. A special test fixture that creates users without nicknames
    // 2. Direct database manipulation to clear the nickname
    // For now, we test that the redirect logic works with complete profiles.

    test.skip('redirects to /auth/complete-profile when nickname is missing', async ({ page }) => {
      // This test is skipped because it requires special setup to create a user
      // without a nickname. The profile completion flow is tested manually or
      // in integration tests with database access.
    });
  });
});
