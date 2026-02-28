import type { z } from 'zod';
import type { UserSchema, AuthResponseSchema } from './schemas';

export type User = z.infer<typeof UserSchema>;
export type AuthResponse = z.infer<typeof AuthResponseSchema>;
