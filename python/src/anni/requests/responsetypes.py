from pydantic import BaseModel

from anni.types.datatypes import NaughtyOrNiceClassificationEnum, Person

class CheckElfLogicRequestPerson(Person):
    pass

class ScoreElfLogicRequestPerson(BaseModel):
    id: int
    naughtyOrNice: NaughtyOrNiceClassificationEnum = NaughtyOrNiceClassificationEnum.NONE

class ScoreElfLogicRequest(BaseModel):
    naughtyNiceCandidateRequests: list[ScoreElfLogicRequestPerson]

class ScoreElfLogicResponse(BaseModel):
    score: int