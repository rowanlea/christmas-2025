from datetime import datetime
from enum import Enum, IntEnum
from typing import List
from pydantic import BaseModel

class OccupationEnum(str, Enum):
    TEACHER = "Teacher"
    ENGINEER = "Engineer"
    ARTIST = "Artist"
    DOCTOR = "Doctor"
    LAWYER = "Lawyer"
    CHILD = "Child"
    PROGRAMMER = "Programmer"

class EventTypeEnum(IntEnum):
    HELPED_SOMEONE = 1
    DONATED_TO_CHARITY = 2
    WAVED_AT_NEIGHBOR = 3
    HELPED_OLD_LADY_CROSS_STREET = 4
    ATE_GREEN_VEGETABLES = 5
    SANG_IN_SHOWER = 6
    DANCED_LIKE_NO_ONE_WAS_WATCHING = 7
    TOLD_A_FUNNY_JOKE = 8
    MADE_A_FUNNY_FACE = 9
    RESCUED_A_KITTEN = 10
    MADE_MESS = 11
    BROKE_SOMETHING = 12
    LIED = 13
    STOLE_SOMETHING = 14
    CHEATED = 15
    WAS_MEAN = 16
    LEFT_TOILET_SEAT_UP = 17
    DID_A_MASSIVE_FART_AND_BLAMED_DOG = 18
    FLIPPED_OFF_A_COP = 19
    LAUGHED_AT_OWN_JOKE = 20

class KataEnum(str, Enum):
    FIZZBUZZ = "fizzbuzz"
    PALINDROME = "palindrome"
    ANAGRAM = "anagram"
    PRIME_NUMBERS = "prime_numbers"
    SORTING_ALGORITHM = "sorting_algorithm"
    MARSROVER = "marsrover"

class Event(BaseModel):
    type: EventTypeEnum
    date: datetime

class NegativeEvent(Event):
    has_apologised: bool

class ScoredEvent(Event):
    event: Event | NegativeEvent
    score: int

class Person(BaseModel):
    id: int
    name: str
    age: int
    location: str
    occupation: OccupationEnum
    favourite_katas: List[str] = []
    events: List[Event | NegativeEvent]

class NaughtyOrNiceClassificationEnum(str, Enum):
    NAUGHTY = "Naughty"
    NICE = "Nice"

class ScoredPerson(Person):
    naughty_or_nice: NaughtyOrNiceClassificationEnum
