/**
 * Rewards API
 *
 * Supabase REST API integration for tournament rewards.
 * Handles CRUD operations with encrypted hidden values.
 */

import { supabase } from '../../../shared/api/supabase'
import type { TournamentReward, RewardFormData, RewardPlacement } from '../types'

// Secret key for encryption - should be stored in environment variable
const REWARD_SECRET_KEY = import.meta.env.VITE_REWARD_SECRET_KEY || 'winspire-rewards-secret-key-2025'

/**
 * Database row type from Supabase
 */
interface RewardRow {
  id: string
  tournament_id: string
  placement: number
  title: string
  description: string | null
  hidden_value_encrypted: string // Base64 encoded
  is_claimed: boolean
  claimed_by_user_id: string | null
  claimed_at: string | null
  created_at: string
  updated_at: string
}

/**
 * Transform database row to domain model
 * Note: hidden_value is NOT decrypted here - use getDecryptedReward for that
 */
function rowToReward(row: RewardRow): TournamentReward {
  return {
    id: row.id,
    tournamentId: row.tournament_id,
    placement: row.placement as RewardPlacement,
    title: row.title,
    description: row.description || '',
    hiddenValue: '', // Never expose encrypted value directly
    isClaimed: row.is_claimed,
    claimedById: row.claimed_by_user_id || undefined,
    claimedAt: row.claimed_at ? new Date(row.claimed_at) : undefined
  }
}

/**
 * Get all rewards for a tournament (public metadata only)
 */
export async function getRewardsForTournament(tournamentId: string): Promise<TournamentReward[]> {
  const { data, error } = await supabase
    .from('tournament_rewards')
    .select('id, tournament_id, placement, title, description, is_claimed, claimed_by_user_id, claimed_at, created_at, updated_at')
    .eq('tournament_id', tournamentId)
    .order('placement', { ascending: true })

  if (error) {
    console.error('Error fetching rewards:', error)
    throw new Error('Nie udało się pobrać nagród')
  }

  return (data || []).map(row => ({
    id: row.id,
    tournamentId: row.tournament_id,
    placement: row.placement as RewardPlacement,
    title: row.title,
    description: row.description || '',
    hiddenValue: '', // Public query - never expose
    isClaimed: row.is_claimed,
    claimedById: row.claimed_by_user_id || undefined,
    claimedAt: row.claimed_at ? new Date(row.claimed_at) : undefined
  }))
}

/**
 * Get decrypted reward value for a winner
 * Only call this for verified winners!
 */
export async function getDecryptedRewardValue(
  rewardId: string,
  userId: string
): Promise<string | null> {
  // Call the decrypt function via RPC
  const { data, error } = await supabase.rpc('decrypt_reward_value', {
    encrypted_value: null, // Will be fetched by function
    secret_key: REWARD_SECRET_KEY
  })

  if (error) {
    console.error('Error decrypting reward:', error)
    return null
  }

  // Alternative: Fetch and decrypt in application
  // For MVP, we'll use a simpler approach - fetch the encrypted value
  // and decrypt client-side (less secure but simpler)
  const { data: reward } = await supabase
    .from('tournament_rewards')
    .select('hidden_value_encrypted')
    .eq('id', rewardId)
    .single()

  if (!reward) return null

  // For MVP: The encryption is handled server-side
  // In production, this would call a secure edge function
  return data as string | null
}

/**
 * Save rewards for a tournament (create or update)
 */
export async function saveRewards(
  tournamentId: string,
  rewards: RewardFormData
): Promise<void> {
  const placements: Array<{ key: keyof RewardFormData; placement: RewardPlacement }> = [
    { key: 'first', placement: 1 },
    { key: 'second', placement: 2 },
    { key: 'third', placement: 3 }
  ]

  for (const { key, placement } of placements) {
    const rewardData = rewards[key]

    // Skip empty rewards
    if (!rewardData.title && !rewardData.hiddenValue) {
      // Delete existing reward if it exists
      await supabase
        .from('tournament_rewards')
        .delete()
        .eq('tournament_id', tournamentId)
        .eq('placement', placement)
      continue
    }

    // Check if reward exists
    const { data: existing } = await supabase
      .from('tournament_rewards')
      .select('id')
      .eq('tournament_id', tournamentId)
      .eq('placement', placement)
      .single()

    if (existing) {
      // Update existing reward
      const { error } = await supabase
        .from('tournament_rewards')
        .update({
          title: rewardData.title,
          description: rewardData.description,
          // For MVP: Store hidden value using RPC encryption function
          hidden_value_encrypted: await encryptValue(rewardData.hiddenValue)
        })
        .eq('id', existing.id)

      if (error) {
        console.error('Error updating reward:', error)
        throw new Error(`Nie udało się zaktualizować nagrody za ${placement}. miejsce`)
      }
    } else {
      // Insert new reward
      const { error } = await supabase
        .from('tournament_rewards')
        .insert({
          tournament_id: tournamentId,
          placement,
          title: rewardData.title,
          description: rewardData.description,
          hidden_value_encrypted: await encryptValue(rewardData.hiddenValue)
        })

      if (error) {
        console.error('Error inserting reward:', error)
        throw new Error(`Nie udało się zapisać nagrody za ${placement}. miejsce`)
      }
    }
  }
}

/**
 * Encrypt value using Supabase RPC function
 * Falls back to base64 encoding for MVP if RPC not available
 */
async function encryptValue(plainText: string): Promise<string> {
  try {
    const { data, error } = await supabase.rpc('encrypt_reward_value', {
      plain_text: plainText,
      secret_key: REWARD_SECRET_KEY
    })

    if (error) throw error
    return data as string
  } catch {
    // MVP fallback: Use base64 encoding (NOT secure for production!)
    console.warn('Encryption RPC not available, using base64 fallback')
    return btoa(plainText)
  }
}

/**
 * Claim a reward (mark as claimed by winner)
 */
export async function claimReward(
  rewardId: string,
  userId: string
): Promise<{ success: boolean; hiddenValue?: string }> {
  const { error } = await supabase
    .from('tournament_rewards')
    .update({
      is_claimed: true,
      claimed_by_user_id: userId,
      claimed_at: new Date().toISOString()
    })
    .eq('id', rewardId)

  if (error) {
    console.error('Error claiming reward:', error)
    return { success: false }
  }

  // Get decrypted value for winner
  const hiddenValue = await getDecryptedRewardValue(rewardId, userId)

  return {
    success: true,
    hiddenValue: hiddenValue || undefined
  }
}

/**
 * Delete all rewards for a tournament
 */
export async function deleteRewardsForTournament(tournamentId: string): Promise<void> {
  const { error } = await supabase
    .from('tournament_rewards')
    .delete()
    .eq('tournament_id', tournamentId)

  if (error) {
    console.error('Error deleting rewards:', error)
    throw new Error('Nie udało się usunąć nagród')
  }
}

/**
 * Get rewards for editing (includes encrypted values for owner)
 * Only for tournament owner!
 */
export async function getRewardsForEditing(tournamentId: string): Promise<RewardFormData> {
  const { data, error } = await supabase
    .from('tournament_rewards')
    .select('*')
    .eq('tournament_id', tournamentId)
    .order('placement', { ascending: true })

  if (error) {
    console.error('Error fetching rewards for editing:', error)
    throw new Error('Nie udało się pobrać nagród do edycji')
  }

  const emptyReward = { title: '', description: '', hiddenValue: '' }
  const result: RewardFormData = {
    first: { ...emptyReward },
    second: { ...emptyReward },
    third: { ...emptyReward }
  }

  for (const row of data || []) {
    const key = row.placement === 1 ? 'first' : row.placement === 2 ? 'second' : 'third'
    result[key] = {
      title: row.title,
      description: row.description || '',
      // Decrypt for editing - MVP uses base64 fallback
      hiddenValue: await decryptValue(row.hidden_value_encrypted)
    }
  }

  return result
}

/**
 * Decrypt value - MVP fallback to base64
 */
async function decryptValue(encryptedValue: string): Promise<string> {
  try {
    const { data, error } = await supabase.rpc('decrypt_reward_value', {
      encrypted_value: encryptedValue,
      secret_key: REWARD_SECRET_KEY
    })

    if (error) throw error
    return data as string
  } catch {
    // MVP fallback: Decode base64
    try {
      return atob(encryptedValue)
    } catch {
      return ''
    }
  }
}

export const rewardsApi = {
  getRewardsForTournament,
  getDecryptedRewardValue,
  saveRewards,
  claimReward,
  deleteRewardsForTournament,
  getRewardsForEditing
}
