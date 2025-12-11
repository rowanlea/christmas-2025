from pydantic_settings import BaseSettings, SettingsConfigDict

class Settings(BaseSettings):
    team_name: str = "We didn't read the instructions" # put your team name here
    base_url: str = "http://localhost:8000"
    team_secret: str = "LMAO" # Can be anything, just needs to be unique to your team to prevent other people using your team name. Will be generated on first run if not set.

    model_config = SettingsConfigDict(
        pyproject_toml_table_header=('tool', 'anni-settings')
    )