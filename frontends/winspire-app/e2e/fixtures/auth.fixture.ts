import {
  getAccessToken,
  getAccountByEmail,
  getStreamer,
  getStreamers,
  getUser,
  getUsers,
  type SeededAccount,
} from '../helpers/auth-seed';

/**
 * Auth helpers for E2E tests.
 * 
 * Usage:
 *   const streamer = await getStreamer(1);  // streamer1@test.local
 *   const user = await getUser(1);          // user01@test.local
 *   const token = await getAccessToken('streamer1@test.local');
 */

export { getAccessToken, getAccountByEmail, getStreamer, getStreamers, getUser, getUsers };
export type { SeededAccount };

