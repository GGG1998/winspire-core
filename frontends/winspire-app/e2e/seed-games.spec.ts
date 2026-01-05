import { test } from '@playwright/test';
import { seedGames } from './fixtures/games.fixture';

test.describe('Database Seeding', () => {
    test('seed games from fixtures', async ({ request }) => {
        // Increase timeout for file uploads
        test.setTimeout(120000);

        console.log('Starting game seeding...');
        await seedGames(request);
        console.log('Game seeding complete.');
    });
});
