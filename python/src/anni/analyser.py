from typing import List
from anni.types.datatypes import Person, ScoredPerson
from anni.calculator.calculator import AutomatedNaughtyOrNiceCalculator
from anni.requests.requests import RequestHandler

class AutomatedNaughtyOrNiceAnalyser:
    def __init__(self, calculator: AutomatedNaughtyOrNiceCalculator | None = None, handler: RequestHandler | None = None):
        self.calculator = calculator or AutomatedNaughtyOrNiceCalculator()
        self.handler = handler or RequestHandler()

    def analyse(self, people: List[Person]) -> List[ScoredPerson]:
        return self.calculator.classify_people(people)

    def submit(self, scored: List[ScoredPerson]):
        return self.handler.submit_list(scored)