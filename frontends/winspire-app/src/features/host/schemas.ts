/**
 * Tournament Feature Schemas
 * Feature: 002-streamer-tournament-creation
 * 
 * Zod validation schemas for tournament forms
 */

import { z } from 'zod';
import { TOURNAMENT_VALIDATION, VALIDATION_MESSAGES } from './constants';

// ============================================================================
// Tournament Form Schema
// ============================================================================

/**
 * Schema for creating a new tournament
 */
export const createTournamentSchema = z.object({
  name: z
    .string()
    .min(1, VALIDATION_MESSAGES.name.required)
    .min(TOURNAMENT_VALIDATION.name.minLength, VALIDATION_MESSAGES.name.minLength)
    .max(TOURNAMENT_VALIDATION.name.maxLength, VALIDATION_MESSAGES.name.maxLength)
    .transform((val) => val.trim()),
  
  startTime: z
    .string()
    .min(1, VALIDATION_MESSAGES.startTime.required)
    .refine(
      (val) => {
        const date = new Date(val);
        return !isNaN(date.getTime());
      },
      { message: 'Nieprawidłowy format daty' }
    )
    .refine(
      (val) => {
        const date = new Date(val);
        const minStartTime = new Date(Date.now() + 60000); // +1 minute
        return date >= minStartTime;
      },
      { message: 'Czas rozpoczęcia musi być co najmniej 1 minutę w przyszłości' }
    ),
  // TODO clean, we use gameSnapshot instead of gameId
  game: z.string().min(1, 'Gra jest wymagana'),
  gameSnapshot: z.object({
    id: z.string().min(1, 'Game ID jest wymagane'),
    slug: z.string().min(1, 'Game slug jest wymagany'),
    name: z.string().min(1, 'Nazwa gry jest wymagana'),
    version: z.string().min(1, 'Wersja gry jest wymagana'),
    storagePath: z.string().min(1, 'Ścieżka do plików gry jest wymagana'),
    logoUrl: z.string().optional(),
    description: z.string().optional(),
  }),

  teamMode: z.string(),
  minPlayers: z.number().min(2),
  maxPlayers: z.number().max(250),
});

// ============================================================================
// Type Inference
// ============================================================================

export type CreateTournamentFormData = z.infer<typeof createTournamentSchema>;
