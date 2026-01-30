/**
 * Zod validation schemas for batch tasks
 */

import { z } from 'zod';

/** Schema for batch task priority */
export const batchTaskPrioritySchema = z.enum(['low', 'normal', 'high']);

/** Schema for batch task status */
export const batchTaskStatusSchema = z.enum(['pending', 'running', 'completed', 'failed']);

/** Schema for a single batch task */
export const batchTaskSchema = z.object({
  id: z.string(),
  prompt: z.string().min(1, 'Prompt is required'),
  context: z.string().optional(),
  priority: batchTaskPrioritySchema,
  status: batchTaskStatusSchema,
  createdAt: z.date(),
  result: z.string().optional(),
  error: z.string().optional(),
});

/** Schema for batch task form data */
export const batchTaskFormSchema = z.object({
  prompt: z.string().min(1, 'Prompt is required').max(10000, 'Prompt is too long'),
  context: z.string().max(5000, 'Context is too long').optional(),
  priority: batchTaskPrioritySchema.default('normal'),
});

/** Schema for validating a list of batch tasks */
export const batchTaskListSchema = z.array(batchTaskSchema);
