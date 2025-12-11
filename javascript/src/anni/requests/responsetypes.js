import { z } from 'zod';
import { PersonSchema, NaughtyOrNiceClassificationEnum } from '../types/datatypes.js';

// CheckElfLogicRequestPerson - same as Person
export const CheckElfLogicRequestPersonSchema = PersonSchema;

// ScoreElfLogicRequestPerson
export const ScoreElfLogicRequestPersonSchema = z.object({
  id: z.number().int(),
  naughtyOrNice: z.enum([
    NaughtyOrNiceClassificationEnum.NONE,
    NaughtyOrNiceClassificationEnum.NAUGHTY,
    NaughtyOrNiceClassificationEnum.NICE
  ])
});

// ScoreElfLogicRequest
export const ScoreElfLogicRequestSchema = z.object({
  naughtyNiceCandidateRequests: z.array(ScoreElfLogicRequestPersonSchema)
});

// ScoreElfLogicResponse
export const ScoreElfLogicResponseSchema = z.object({
  score: z.number().int()
});


