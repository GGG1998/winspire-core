import fs from 'fs';
import path from 'path';
import { APIRequestContext } from '@playwright/test';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Use environment variables for flexibility
const GAME_MGMT_API_URL = process.env.GAME_MANAGEMENT_API_URL ?? 'http://localhost:8087';
const INTERNAL_KEY = process.env.GAME_MANAGEMENT_INTERNAL_API_KEY ?? 'IlikeCookies';

/**
 * Helper to recursively get all files in a directory
 */
function getAllFiles(dirPath: string, arrayOfFiles: string[] = []) {
    const files = fs.readdirSync(dirPath);

    files.forEach(function (file) {
        const fullPath = path.join(dirPath, file);
        if (fs.statSync(fullPath).isDirectory()) {
            arrayOfFiles = getAllFiles(fullPath, arrayOfFiles);
        } else {
            arrayOfFiles.push(fullPath);
        }
    });

    return arrayOfFiles;
}

/**
 * Seeds games from the e2e/fixtures/games folder into the game-management service.
 * Follows the mini-admin implementation: upload all files first, then create the game record.
 */
export async function seedGames(request: APIRequestContext) {
    const gamesDir = path.join(__dirname, 'games');

    if (!fs.existsSync(gamesDir)) {
        console.warn(`[seedGames] Games directory not found at ${gamesDir}`);
        return;
    }

    const gameFolders = fs.readdirSync(gamesDir).filter(f =>
        fs.statSync(path.join(gamesDir, f)).isDirectory()
    );

    console.log(`[seedGames] Found ${gameFolders.length} game(s) to seed.`);

    for (const folder of gameFolders) {
        const gamePath = path.join(gamesDir, folder);
        const slug = folder.toLowerCase();
        const version = '1.0.0'; // Default version for seeding

        console.log(`[seedGames] Seeding game: ${folder} (slug: ${slug})`);

        // 1. Upload files first (mini-admin pattern)
        const allFiles = getAllFiles(gamePath);
        console.log(`[seedGames] Uploading ${allFiles.length} files for ${folder}...`);

        for (const filePath of allFiles) {
            const relativePath = path.relative(gamePath, filePath);

            // Note: Playwright's request.post with multipart handles fs.ReadStream
            const uploadRes = await request.post(`${GAME_MGMT_API_URL}/v1/admin/games/uploads`, {
                headers: { 'X-Internal-Service-Key': INTERNAL_KEY },
                multipart: {
                    slug: String(slug),
                    version: String(version),
                    filePath: String(relativePath),
                    files: fs.createReadStream(filePath)
                }
            });

            if (!uploadRes.ok()) {
                const text = await uploadRes.text();
                console.error(`[seedGames] Failed to upload ${relativePath}: ${uploadRes.status()} ${text}`);
            } else {
                console.log(`[seedGames] Uploaded: ${relativePath}`);
            }
        }

        // 2. Create game entry after all files are uploaded
        const createRes = await request.post(`${GAME_MGMT_API_URL}/v1/admin/games`, {
            headers: { 'X-Internal-Service-Key': INTERNAL_KEY },
            data: {
                slug,
                version,
                name: folder,
                description: `Seeded game: ${folder}`,
                gameIntegrationId: null,
                storagePath: `${slug}/${version}`,
                logoURL: `https://dummyimage.com/200x200/000/fff&text=${folder}`,
            }
        });

        if (!createRes.ok()) {
            const text = await createRes.text();
            // Handle existing games
            if (createRes.status() === 409 || text.includes('already exists') || text.includes('duplicate key')) {
                console.log(`[seedGames] Game ${folder} record already exists, files were refreshed.`);
            } else {
                console.error(`[seedGames] Failed to create game record ${folder}: ${createRes.status()} ${text}`);
            }
        } else {
            console.log(`[seedGames] Successfully created game record and seeded ${folder}`);
        }
    }
}
