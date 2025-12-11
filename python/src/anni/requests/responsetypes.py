from pydantic import BaseModel

from anni.types.datatypes import NaughtyOrNiceClassificationEnum, Person

class CheckElfLogicRequestPerson(Person):
    pass

class CheckElfLogicResponse(BaseModel):
    id: int
    name: str
    status: NaughtyOrNiceClassificationEnum

class ScoreElfLogicRequestPerson(BaseModel):
    id: int
    name: str
    status: NaughtyOrNiceClassificationEnum

class ScoreElfLogicResponse(BaseModel):
    teamName: str
    correct: int
    incorrect: int 