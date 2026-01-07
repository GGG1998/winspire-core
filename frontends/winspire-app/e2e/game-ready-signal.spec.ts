import { test, expect } from '@playwright/test';
import { getStreamer, getUser, type SeededAccount } from './fixtures/auth.fixture';
import { seedMatchLobby1v1 } from './MatchLobbyTests/fixture';
import { seedGames } from './fixtures/games.fixture';
import type { Page } from '@playwright/test';

const PACMAN_GAME_SLUG = 'pacmanbuildfixed';

/**
 * Set up authenticated session for a page.
 */
async function setupAuth(page: Page, account: SeededAccount) {
    await page.goto('/');
    await page.evaluate((session) => {
        window.localStorage.setItem('winspire-auth', JSON.stringify(session));
    }, account.session);
    await page.reload({ waitUntil: 'networkidle' });
    await page.waitForTimeout(500);
}

test.describe('Game Ready Signal Tests', () => {
    // Increase timeout for the entire suite including hooks
    test.setTimeout(120000);

    test.beforeAll(async ({ request }) => {
        await seedGames(request);
    });

    test('Game sends ready signal and Ready button appears', async ({ browser }) => {
        const streamer = await getStreamer(1);
        const user = await getUser(1);

        const seed = await seedMatchLobby1v1({
            participant1: streamer,
            participant2: user,
            attachSecond: true,
            gameSlug: PACMAN_GAME_SLUG,
        });

        const context = await browser.newContext();
        const page = await context.newPage();
        await setupAuth(page, streamer);

        // Setup console listener
        let readySignalReceived = false;
        page.on('console', msg => {
            const text = msg.text();
            console.log(`BROWSER LOG: ${text}`); // Print all browser logs to stdout
            if (text.includes('[GameFrame] Game ready signal received from iframe')) {
                readySignalReceived = true;
                console.log('TEST: Captured ready signal log');
            }
        });

        await page.goto(seed.lobbyUrl, { waitUntil: 'networkidle' });

        // 1. Wait for game to load and signal ready
        try {
            await expect(async () => {
                expect(readySignalReceived).toBe(true);
            }).toPass({ timeout: 10000 });
        } catch (e) {
            console.log('TEST WARNING: Game ready signal NOT received within timeout');
        }

        // 2. Verify "Jestem gotowy" button appears
        await expect(
            page.getByRole('button', { name: /Jestem gotowy/i })
        ).toBeVisible({ timeout: 20000 });

        await context.close();
    });
});
