from pydantic_settings import BaseSettings, PydanticBaseSettingsSource, PyprojectTomlConfigSettingsSource, SettingsConfigDict

class Settings(BaseSettings):
    team_name: str = "We didn't read the instructions" # put your team name here
    base_url: str = "http://localhost:8000"
    team_secret: str = "LMAO" # Given to you by /register

    model_config = SettingsConfigDict(
        pyproject_toml_table_header=('tool', 'anni-settings')
    )

    @classmethod
    def settings_customise_sources(
        cls,
        settings_cls: type[BaseSettings],
        init_settings: PydanticBaseSettingsSource,
        env_settings: PydanticBaseSettingsSource,
        dotenv_settings: PydanticBaseSettingsSource,
        file_secret_settings: PydanticBaseSettingsSource,
    ) -> tuple[PydanticBaseSettingsSource, ...]:
        return (PyprojectTomlConfigSettingsSource(settings_cls),)
