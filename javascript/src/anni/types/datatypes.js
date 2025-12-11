import { z } from 'zod';

// Occupation Enum
export const OccupationEnum = {
  TEACHER: "Teacher",
  ENGINEER: "Engineer",
  ARTIST: "Artist",
  DOCTOR: "Doctor",
  LAWYER: "Lawyer",
  CHILD: "Child",
  PROGRAMMER: "Programmer"
};

// Event Type Enum
export const EventTypeEnum = {
  HELPED_SOMEONE: 1,
  DONATED_TO_CHARITY: 2,
  WAVED_AT_NEIGHBOR: 3,
  HELPED_OLD_LADY_CROSS_STREET: 4,
  ATE_GREEN_VEGETABLES: 5,
  SANG_IN_SHOWER: 6,
  DANCED_LIKE_NO_ONE_WAS_WATCHING: 7,
  TOLD_A_FUNNY_JOKE: 8,
  MADE_A_FUNNY_FACE: 9,
  RESCUED_A_KITTEN: 10,
  MADE_MESS: 11,
  BROKE_SOMETHING: 12,
  LIED: 13,
  STOLE_SOMETHING: 14,
  CHEATED: 15,
  WAS_MEAN: 16,
  LEFT_TOILET_SEAT_UP: 17,
  DID_A_MASSIVE_FART_AND_BLAMED_DOG: 18,
  FLIPPED_OFF_A_COP: 19,
  LAUGHED_AT_OWN_JOKE: 20
};

// Kata Enum
export const KataEnum = {
  FIZZBUZZ: "fizzbuzz",
  PALINDROME: "palindrome",
  ANAGRAM: "anagram",
  PRIME_NUMBERS: "prime_numbers",
  SORTING_ALGORITHM: "sorting_algorithm",
  MARSROVER: "marsrover"
};

// Naughty or Nice Classification Enum
export const NaughtyOrNiceClassificationEnum = {
  NONE: "NONE",
  NAUGHTY: "NAUGHTY",
  NICE: "NICE"
};

// Zod Schemas for validation
export const EventSchema = z.object({
  type: z.number().int().min(1).max(20),
  date: z.string().datetime()
});

export const NegativeEventSchema = EventSchema.extend({
  has_apologised: z.boolean()
});

export const ScoredEventSchema = z.object({
  event: z.union([EventSchema, NegativeEventSchema]),
  score: z.number().int()
});

export const PersonSchema = z.object({
  id: z.number().int(),
  name: z.string(),
  age: z.number().int(),
  location: z.string(),
  occupation: z.enum([
    OccupationEnum.TEACHER,
    OccupationEnum.ENGINEER,
    OccupationEnum.ARTIST,
    OccupationEnum.DOCTOR,
    OccupationEnum.LAWYER,
    OccupationEnum.CHILD,
    OccupationEnum.PROGRAMMER
  ]),
  favourite_katas: z.array(z.string()).default([]),
  events: z.array(z.union([EventSchema, NegativeEventSchema]))
});

export const ScoredPersonSchema = PersonSchema.extend({
  naughty_or_nice: z.enum([
    NaughtyOrNiceClassificationEnum.NONE,
    NaughtyOrNiceClassificationEnum.NAUGHTY,
    NaughtyOrNiceClassificationEnum.NICE
  ]).default(NaughtyOrNiceClassificationEnum.NONE)
});

// Helper function to create a Person
export function createPerson(data) {
  return PersonSchema.parse(data);
}

// Helper function to create a ScoredPerson
export function createScoredPerson(data) {
  return ScoredPersonSchema.parse(data);
}


