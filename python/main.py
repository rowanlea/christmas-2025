import argparse
import os
import sys
# Ensure `src` is on the module search path when running this file directly
sys.path.append(os.path.join(os.path.dirname(__file__), "src"))
from typing import List
from pydantic import TypeAdapter
from anni.types.datatypes import OccupationEnum, Person
from anni.analyser import AutomatedNaughtyOrNiceAnalyser
from testharness.testharness import TestHarness

def _load_people_from_file(path: str) -> List[Person]:
    adapter = TypeAdapter(List[Person])
    with open(path, "r", encoding="utf-8") as fh:
        return adapter.validate_json(fh.read())

def main():
    parser = argparse.ArgumentParser(description="Automated Naughty or Nice Analyzer")
    parser.add_argument("--test", action="store_true", help="Run in test harness mode")
    parser.add_argument("--score", action="store_true", help="Run in scoring mode. WILL SUBMIT DATA TO SERVER FOR SCORING.")
    args = parser.parse_args()

    if args.test:
        harness = TestHarness()
        person = Person(
            id=789433,
            name="poo",
            age=21,
            location="lahndon",
            occupation=OccupationEnum.PROGRAMMER,
            events=[],
        )
        result = harness.check_elf_logic(person)
        print("Test Harness Result:", result)
    elif args.score:
        analyser = AutomatedNaughtyOrNiceAnalyser()
        data_file = os.path.join(os.path.dirname(__file__), "src", "data", "validation-data.json")
        people = _load_people_from_file(data_file)
        analysed_people = analyser.analyse(people)
        result = analyser.submit(analysed_people)
        print("Scoring Result:", result)
    else:
        parser.print_help()

if __name__ == "__main__":
    main()