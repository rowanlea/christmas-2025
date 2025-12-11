from typing import List
from pydantic import TypeAdapter
from anni.requests.responsetypes import (
    CheckElfLogicRequestPerson,
    ScoreElfLogicRequest,
    ScoreElfLogicRequestPerson,
    ScoreElfLogicResponse,
)
from anni.types.datatypes import NaughtyOrNiceClassificationEnum, Person, ScoredPerson
from anni.types.settings import Settings
import requests

class RequestHandler:
    def __init__(self, settings: Settings | None = None, session: requests.Session | None = None):
        self.settings = settings or Settings()
        self.session = session or requests.Session()
        self.check_elf_logic_adaptor = TypeAdapter(List[CheckElfLogicRequestPerson])
        self.elf_check_response_adaptor = TypeAdapter(NaughtyOrNiceClassificationEnum)
        self.elf_logic_request_adaptor = TypeAdapter(List[ScoreElfLogicRequestPerson])

    def _get_team_name(self):
        if not self.settings.team_name:
            raise ValueError("Team name is not set in settings.")
        return self.settings.team_name

    def _get_team_secret(self):
        if not self.settings.team_secret:
            raise ValueError("Unique key is not set in settings.")
        return self.settings.team_secret

    def _get_auth_headers(self):
        return {
            "X-Team-Name": self._get_team_name(),
            "X-Unique-Key": self._get_team_secret(),
            "Content-Type": "application/json",
            "Accept": "application/json",
        }

    def _url(self, endpoint: str) -> str:
        return f"{self.settings.base_url.rstrip('/')}/{endpoint.lstrip('/')}"

    def _post(self, endpoint: str, json_data: dict):
        response = self.session.post(
            self._url(endpoint),
            json=json_data,
            headers=self._get_auth_headers(),
            timeout=10,
        )
        print(response.request.body)
        response.raise_for_status()
        return response.json()
    
    def _getSingle(self, endpoint: str, json_data: dict):
        response = self.session.get(
            self._url(endpoint),
            json=json_data,
            headers=self._get_auth_headers(),
            timeout=10,
        )
        response.raise_for_status()
        return response.json()

    # Public API
    def check_elf_logic(self, person: Person):
        endpoint = "naughtyornice"
        request_data = CheckElfLogicRequestPerson(**person.model_dump()).model_dump()
        parsed_response = self._getSingle(endpoint, request_data)
        return self.elf_check_response_adaptor.validate_python(parsed_response)

    def submit_list(self, people: List[ScoredPerson]):
        endpoint = "score"
        request_data = ScoreElfLogicRequest(
            naughtyNiceCandidateRequests=[ScoreElfLogicRequestPerson(id=p.id, naughtyOrNice=p.naughty_or_nice) for p in people]
        ).model_dump()

        parsed_response = self._post(endpoint, request_data)
        return ScoreElfLogicResponse.model_validate(parsed_response)