from typing import Iterable, List
from anni.types.datatypes import Event, NaughtyOrNiceClassificationEnum, Person, ScoredPerson


class AutomatedNaughtyOrNiceCalculator:
    def __init__(self):
        pass

    def classify_people(self, people: Iterable[Person]) -> List[ScoredPerson]:
        return [self.classify(person) for person in people]

    def classify(self, person: Person) -> ScoredPerson:
        # Placeholder logic for classification; replace with real rules.
        decision = self._decide(person.events)
        return ScoredPerson(**person.model_dump(), naughty_or_nice=decision)

    def _decide(self, events: List[Event]) -> NaughtyOrNiceClassificationEnum:
        # Very naive example: any events -> NICE else NAUGHTY (adjust later)
        if events:
            return NaughtyOrNiceClassificationEnum.NICE
        return NaughtyOrNiceClassificationEnum.NAUGHTY