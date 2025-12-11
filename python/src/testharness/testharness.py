from anni.requests.requests import RequestHandler
from anni.types.datatypes import Person

class TestHarness:
    def __init__(self, handler: RequestHandler | None = None):
        self.handler = handler or RequestHandler()

    def check_elf_logic(self, people: list[Person]):
        return self.handler.check_elf_logic(people)