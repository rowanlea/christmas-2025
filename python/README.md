# Christmas Kata - Python Version

This is the Python implementation of the Automated Naughty or Nice Analyzer.

## Setup

1. Install dependencies using your preferred Python package manager:

   Using pip:
   ```bash
   pip install -r requirements.txt
   ```

   Or using uv (if available):
   ```bash
   uv sync
   ```

   Required packages:
   - faker >= 37.11.0
   - pydantic >= 2.12.3
   - pydantic-settings >= 2.11.0
   - requests >= 2.32.5

2. Configure your team settings in `pyproject.toml`:
   ```toml
   [tool.anni-settings]
   team_name = "your-team-name"
   base_url = "https://lscc.lifeatcodurance.com/"
   team_secret = "your-team-secret"
   ```

## Usage

### Test Mode
Run the test harness with a sample person:
```bash
python main.py --test
```

### Score Mode
Analyze validation data and submit for scoring (this will submit data to the server):
```bash
python main.py --score
```

## Project Structure

```
python/
├── main.py                              # Entry point
├── pyproject.toml                       # Configuration and dependencies
├── src/
│   ├── anni/
│   │   ├── analyser.py                  # Main analyser class
│   │   ├── calculator/
│   │   │   └── calculator.py            # Classification logic
│   │   ├── requests/
│   │   │   ├── requests.py              # HTTP request handler
│   │   │   └── responsetypes.py         # Request/response types
│   │   └── types/
│   │       ├── datatypes.py             # Core data models and enums
│   │       └── settings.py              # Settings management
│   ├── data/
│   │   ├── data-schema.json             # JSON schema for validation
│   │   ├── sample-data.json             # Sample data
│   │   └── validation-data.json         # Validation data for scoring
│   └── testharness/
│       └── testharness.py               # Test harness
└── README.md                            # This file
```

## Requirements

- Python >= 3.11

## Implementation Notes

This Python version uses:

- **Pydantic** for data validation and settings management
- **requests** for HTTP client functionality
- Type hints throughout for better code clarity
- Settings loaded from `pyproject.toml` using `pydantic-settings`

## Customizing the Classification Logic

The classification logic is in `src/anni/calculator/calculator.py`. The current implementation is a placeholder - modify the `_decide()` method to implement your own naughty/nice classification rules based on the person's events.

## Event Types

The system recognizes the following event types (defined in `src/anni/types/datatypes.py`):

**Nice Events (1-10):**
- HELPED_SOMEONE
- DONATED_TO_CHARITY
- WAVED_AT_NEIGHBOR
- HELPED_OLD_LADY_CROSS_STREET
- ATE_GREEN_VEGETABLES
- SANG_IN_SHOWER
- DANCED_LIKE_NO_ONE_WAS_WATCHING
- TOLD_A_FUNNY_JOKE
- MADE_A_FUNNY_FACE
- RESCUED_A_KITTEN

**Naughty Events (11-20):**
- MADE_MESS
- BROKE_SOMETHING
- LIED
- STOLE_SOMETHING
- CHEATED
- WAS_MEAN
- LEFT_TOILET_SEAT_UP
- DID_A_MASSIVE_FART_AND_BLAMED_DOG
- FLIPPED_OFF_A_COP
- LAUGHED_AT_OWN_JOKE

## Data Models

### Person
- id: int
- name: str
- age: int
- location: str
- occupation: OccupationEnum
- favourite_katas: List[str] (optional)
- events: List[Event | NegativeEvent]

### Event
- type: EventTypeEnum
- date: datetime

### NegativeEvent (extends Event)
- has_apologised: bool

### ScoredPerson (extends Person)
- naughty_or_nice: NaughtyOrNiceClassificationEnum (NONE, NAUGHTY, or NICE)


