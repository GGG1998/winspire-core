import { z } from 'zod';

export const loginCredentialsSchema = z.object({
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address'),
  password: z
    .string()
    .min(1, 'Password is required'),
});

const baseRegisterSchemaObject = {
  email: z
    .string()
    .min(1, 'Email is required')
    .email('Invalid email address'),
  password: z
    .string()
    .min(6, 'Password must be at least 6 characters'),
  confirmPassword: z
    .string()
    .min(1, 'Please confirm your password'),
  nickname: z
    .string()
    .min(1, 'Nickname is required'),
  first_name: z
    .string()
    .min(1, 'First name is required'),
  last_name: z
    .string()
    .min(1, 'Last name is required'),
  country_id: z
    .union([
      z.string().uuid('Invalid country selection'),
      z.literal(''),
    ])
    .optional(),
  city: z
    .string()
    .optional(),
};

export const baseRegisterSchema = z.object(baseRegisterSchemaObject).refine(
  (data) => data.password === data.confirmPassword,
  {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  }
);

export const userRegisterSchema = z.object({
  ...baseRegisterSchemaObject,
  profileType: z.literal('user'),
}).refine(
  (data) => data.password === data.confirmPassword,
  {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  }
);

export const streamerRegisterSchema = z.object({
  ...baseRegisterSchemaObject,
  profileType: z.literal('streamer'),
}).refine(
  (data) => data.password === data.confirmPassword,
  {
    message: 'Passwords do not match',
    path: ['confirmPassword'],
  }
);

export type LoginCredentialsInput = z.infer<typeof loginCredentialsSchema>;
export type UserRegisterInput = z.infer<typeof userRegisterSchema>;
export type StreamerRegisterInput = z.infer<typeof streamerRegisterSchema>;

