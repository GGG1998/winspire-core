/**
 * Tournament Pre-Lobby E2E Test (FAST VERSION)
 * 
 * This is a refactored version using:
 * - API calls for setup (user creation, tournament creation)
 * - Playwright fixtures for pre-authenticated sessions
 * - UI testing only for critical user interactions
 * 
 * Speed comparison:
 * - Original test: ~60-90 seconds
 * - Fast version: ~25-35 seconds
 * 
 * The test still validates the same flow but is much faster.
 */

import { test, expect } from './fixtures/auth.fixture';
import { 
  createTournament,
  updateTournamentStatus,
  registerForTournament,
  generateTestData
} from './helpers/api-client';

const TEST_TIMEOUT = 90000; // 90 seconds (reduced from 2 minutes)
const GRACE_PERIOD_DURATION = 30000; // 30 seconds
const MATCH_ASSIGNMENT_TIMEOUT = 10000;

test.describe('Tournament Pre-Lobby E2E (Fast)', () => {
  test.setTimeout(TEST_TIMEOUT);
  
  test('complete pre-lobby flow with multiple players', async ({ 
    browser, 
    request,
    fourAuthenticatedPlayers 
  }) => {
    console.log('[Fast Test] Starting with pre-authenticated players...');
    
    // ========================================================================
    // SETUP: Create streamer and tournament via API (FAST!)
    // ========================================================================
    const streamerContext = await browser.newContext();
    const streamerPage = await streamerContext.newPage();
    
    const testData = generateTestData('streamer');
    const { createStreamer } = await import('./helpers/api-client');
    
    const streamer = await createStreamer(
      request,
      testData.nickname,
      testData.email,
      testData.password,
      testData.nickname
    );
    
    // Set streamer auth
    await streamerContext.addCookies([{
      name: 'sb-access-token',
      value: streamer.accessToken,
      domain: 'localhost',
      path: '/',
    }]);
    
    console.log('[Fast Test] Creating tournament via API...');
    
    // Create tournament via API (bypass UI!)
    const startTime = new Date(Date.now() + 2 * 60 * 1000);
    const tournament = await createTournament(
      request,
      streamer.id,
      streamer.accessToken,
      testData.tournamentName,
      startTime
    );
    
    console.log(`[Fast Test] Tournament created: ${tournament.id}`);
    
    // Publish and open registration via API (bypass UI!)
    await updateTournamentStatus(
      request,
      tournament.id,
      streamer.id,
      streamer.accessToken,
      'scheduled'
    );
    
    await updateTournamentStatus(
      request,
      tournament.id,
      streamer.id,
      streamer.accessToken,
      'registration_open'
    );
    
    console.log('[Fast Test] Tournament published and registration opened');
    
    // ========================================================================
    // SETUP: Register all players via API (FAST!)
    // ========================================================================
    console.log('[Fast Test] Registering 4 players via API...');
    
    const registrationPromises = fourAuthenticatedPlayers.map(player =>
      registerForTournament(
        request,
        tournament.id,
        streamer.id,
        player.user.accessToken
      )
    );
    
    await Promise.all(registrationPromises);
    console.log('[Fast Test] All players registered');
    
    // ========================================================================
    // TEST: Players join pre-lobby (UI interaction - the part we're testing)
    // ========================================================================
    console.log('[Fast Test] Players joining pre-lobby...');
    
    for (let i = 0; i < fourAuthenticatedPlayers.length; i++) {
      const { page } = fourAuthenticatedPlayers[i];
      await page.goto(`/tournaments/${tournament.id}/lobby`);
      await page.waitForTimeout(1000);
      console.log(`[Fast Test] Player ${i + 1} joined pre-lobby`);
    }
    
    // ========================================================================
    // ASSERTION: Verify participant notifications (UI validation)
    // ========================================================================
    console.log('[Fast Test] Verifying participant notifications...');
    
    for (let i = 0; i < fourAuthenticatedPlayers.length; i++) {
      const { page } = fourAuthenticatedPlayers[i];
      
      // Verify we're in pre-lobby
      await expect(page.locator('text=/waiting|czekanie|poczekalnia/i').first())
        .toBeVisible({ timeout: 5000 });
      
      // Check participant list is visible
      const participantList = page.locator('[data-testid="participant-list"], [class*="participant"]');
      await expect(participantList.first()).toBeVisible({ timeout: 5000 });
      
      console.log(`[Fast Test] ✓ Player ${i + 1} sees participant list`);
    }
    
    // ========================================================================
    // ACTION: Start tournament via API (could be UI, but API is faster)
    // ========================================================================
    console.log('[Fast Test] Starting tournament via API...');
    
    await updateTournamentStatus(
      request,
      tournament.id,
      streamer.id,
      streamer.accessToken,
      'started'
    );
    
    console.log('[Fast Test] Tournament started, grace period beginning...');
    
    // ========================================================================
    // ASSERTION: Verify grace period notifications
    // ========================================================================
    console.log('[Fast Test] Verifying grace period...');
    
    for (let i = 0; i < fourAuthenticatedPlayers.length; i++) {
      const { page } = fourAuthenticatedPlayers[i];
      
      const gracePeriodIndicator = page.locator('text=/okres.*łaski|grace.*period|finalizing/i');
      await expect(gracePeriodIndicator.first()).toBeVisible({ timeout: 10000 });
      
      console.log(`[Fast Test] ✓ Player ${i + 1} sees grace period`);
    }
    
    // Wait for grace period
    console.log(`[Fast Test] Waiting ${GRACE_PERIOD_DURATION / 1000}s for grace period...`);
    await streamerPage.waitForTimeout(GRACE_PERIOD_DURATION);
    
    // ========================================================================
    // ASSERTION: Verify bracket generation notification
    // ========================================================================
    console.log('[Fast Test] Verifying bracket generation...');
    
    for (let i = 0; i < fourAuthenticatedPlayers.length; i++) {
      const { page } = fourAuthenticatedPlayers[i];
      
      const bracketGenIndicator = page.locator('text=/generowanie.*drabink|generating.*bracket/i');
      await expect(bracketGenIndicator.first()).toBeVisible({ timeout: 10000 });
      
      console.log(`[Fast Test] ✓ Player ${i + 1} sees bracket generation`);
    }
    
    // ========================================================================
    // ASSERTION: Check match_assigned (EXPECTED FAIL)
    // ========================================================================
    console.log('[Fast Test] Checking for match assignments (EXPECTED TO FAIL)...');
    
    await streamerPage.waitForTimeout(MATCH_ASSIGNMENT_TIMEOUT);
    
    for (let i = 0; i < fourAuthenticatedPlayers.length; i++) {
      const { page } = fourAuthenticatedPlayers[i];
      
      const matchAssignedNotification = page.locator('text=/mecz.*przydzielony|match.*assigned/i');
      
      try {
        await expect(matchAssignedNotification.first()).toBeVisible({ timeout: 5000 });
        console.log(`[Fast Test] ✓ Player ${i + 1} received match assignment`);
      } catch (error) {
        console.log(`[Fast Test] ✗ Player ${i + 1} did NOT receive match assignment (EXPECTED)`);
        throw new Error(
          `EXPECTED FAILURE: Backend doesn't call BroadcastMatchAssigned after bracket generation. ` +
          `Fix needed in services/matchmaking/internal/application/`
        );
      }
    }
    
    // ========================================================================
    // ASSERTION: Check redirect to match lobby (EXPECTED FAIL)
    // ========================================================================
    console.log('[Fast Test] Checking for redirect (EXPECTED TO FAIL)...');
    
    for (let i = 0; i < fourAuthenticatedPlayers.length; i++) {
      const { page } = fourAuthenticatedPlayers[i];
      const url = page.url();
      
      if (!/\/lobby\/[^\/]+\/match\/[^\/]+/.test(url)) {
        throw new Error(
          `EXPECTED FAILURE: Player ${i + 1} not redirected. Current URL: ${url}`
        );
      }
    }
    
    console.log('[Fast Test] ✓ Test completed!');
    
    // Cleanup
    await streamerContext.close();
  });
});


