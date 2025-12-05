/**
 * Test Cleanup Utilities
 * 
 * Helpers for cleaning up test data after E2E tests.
 * Use sparingly - prefer test isolation over cleanup.
 */

import { APIRequestContext } from '@playwright/test';

const SUPABASE_URL = process.env.SUPABASE_URL || 'http://localhost:54321';
const SUPABASE_SERVICE_ROLE_KEY = process.env.SUPABASE_SERVICE_ROLE_KEY || '';

/**
 * Delete test users by email pattern (requires service role key)
 * 
 * WARNING: This is destructive! Use with caution.
 * 
 * @example
 * // Delete all users with @test.local emails
 * await cleanupTestUsers(request, '@test.local');
 */
export async function cleanupTestUsers(
  request: APIRequestContext,
  emailPattern: string
): Promise<number> {
  if (!SUPABASE_SERVICE_ROLE_KEY) {
    console.warn('[Cleanup] SUPABASE_SERVICE_ROLE_KEY not set, skipping cleanup');
    return 0;
  }

  console.log(`[Cleanup] Looking for users matching "${emailPattern}"...`);

  try {
    // Note: This requires a custom admin endpoint or direct DB access
    // Supabase doesn't provide a bulk user deletion API
    
    // For now, log a warning
    console.warn(
      '[Cleanup] Automatic user cleanup not implemented. ' +
      'Clean test users manually via Supabase Dashboard or direct SQL.'
    );
    
    // SQL you can run manually:
    console.log(`
      -- SQL to clean test users:
      DELETE FROM auth.users 
      WHERE email LIKE '%${emailPattern}%';
      
      -- Or via Supabase CLI:
      supabase db reset --linked
    `);

    return 0;
  } catch (error) {
    console.error('[Cleanup] Failed to cleanup users:', error);
    return 0;
  }
}

/**
 * Best practice: Use unique data per test instead of cleanup
 * 
 * Instead of:
 * ```
 * test.afterAll(async () => {
 *   await cleanupTestUsers(request, '@test.local');
 * });
 * ```
 * 
 * Do this:
 * ```
 * const testData = generateTestData('player'); // Already unique!
 * ```
 */

