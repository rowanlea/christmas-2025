# LSCC Christmas Kata - Automated Naughty or Nice Analyzer

This repository contains implementations of the Automated Naughty or Nice Analyzer in both Python and JavaScript.

## Project Structure

```
lscc-christmas-kata/
├── python/          # Python implementation
└── javascript/      # JavaScript implementation
```

## Available Implementations

### Python Version

Located in the `python/` directory.

**Requirements:**
- Python >= 3.11
- Dependencies: faker, pydantic, pydantic-settings, requests

**Setup:**
```bash
cd python
# Install dependencies using uv, pip, or your preferred package manager
```

**Configuration:**
Edit `pyproject.toml` to set your team name and secret:
```toml
[tool.anni-settings]
team_name = "your-team-name"
base_url = "https://lscc.lifeatcodurance.com/"
team_secret = "your-team-secret"
```

**Usage:**
```bash
# Test mode
python main.py --test

# Scoring mode (submits to server)
python main.py --score
```

See [python/README.md](python/README.md) for more details.

### JavaScript Version

Located in the `javascript/` directory.

**Requirements:**
- Node.js >= 18.0.0
- Dependencies: axios, zod

**Setup:**
```bash
cd javascript
npm install
```

**Configuration:**
Edit `config.json` to set your team name and secret:
```json
{
  "team_name": "your-team-name",
  "base_url": "https://lscc.lifeatcodurance.com/",
  "team_secret": "your-team-secret"
}
```

**Usage:**
```bash
# Test mode
npm run test

# Scoring mode (submits to server)
npm run score
```

See [javascript/README.md](javascript/README.md) for more details.

## Project Overview

This project analyzes people's events throughout the year to classify them as "Naughty" or "Nice" for Santa's list.

### Key Components

Both implementations share the same architecture:

1. **Data Types** - Define the core models (Person, Event, etc.)
2. **Calculator** - Contains the classification logic
3. **Request Handler** - Manages API communication
4. **Analyser** - Orchestrates the analysis process
5. **Test Harness** - Provides testing utilities

### Data Files

Located in `src/data/` in each implementation:
- `data-schema.json` - JSON schema for validation
- `sample-data.json` - Sample data for testing
- `validation-data.json` - Validation data for scoring

## Customizing Classification Logic

To implement your own naughty/nice classification rules:

**Python:** Edit `python/src/anni/calculator/calculator.py` - modify the `_decide()` method

**JavaScript:** Edit `javascript/src/anni/calculator/calculator.js` - modify the `_decide()` method

The current implementation is a placeholder that classifies people with events as "NICE" and people without events as "NAUGHTY".

## License

This is a coding kata project for LSCC.


