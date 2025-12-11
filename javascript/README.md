# Christmas Kata - JavaScript Version

This is the JavaScript implementation of the Automated Naughty or Nice Analyzer.


## Setup

1. Install dependencies:
   ```bash
   npm install
   ```

2. Register your team at https://lscc.lifeatcodurance.com/your-team-name/

3. Configure your team settings in `config.json`:
   ```json
   {
     "team_name": "your-team-name",
     "base_url": "https://lscc.lifeatcodurance.com/",
     "team_secret": "your-team-secret"
   }
   ```

## Usage

### Test Mode
Run the test harness with a sample person:
```bash
npm run test
```

or

```bash
node main.js --test
```

### Score Mode
Analyze validation data and submit for scoring (this will submit data to the server):
```bash
npm run score
```

or

```bash
node main.js --score
```

## Project Structure

```
javascript/
├── main.js                              # Entry point
├── config.json                          # Configuration file
├── package.json                         # Node.js dependencies
├── src/
│   ├── anni/
│   │   ├── analyser.js                  # Main analyser class
│   │   ├── calculator/
│   │   │   └── calculator.js            # Classification logic
│   │   ├── requests/
│   │   │   ├── requests.js              # HTTP request handler
│   │   │   └── responsetypes.js         # Request/response schemas
│   │   └── types/
│   │       ├── datatypes.js             # Core data models and enums
│   │       └── settings.js              # Settings management
│   ├── data/
│   │   ├── data-schema.json             # JSON schema for validation
│   │   ├── sample-data.json             # Sample data
│   │   └── validation-data.json         # Validation data for scoring
│   └── testharness/
│       └── testharness.js               # Test harness
└── README.md                            # This file
```

## Dependencies

- **axios** (^1.6.2): HTTP client for making API requests
- **zod** (^3.22.4): TypeScript-first schema validation

## Requirements

- Node.js >= 18.0.0

## Implementation Notes

This JavaScript version maintains the same structure and functionality as the Python version:

- Uses **Zod** for schema validation (similar to Pydantic in Python)
- Uses **axios** for HTTP requests (similar to requests in Python)
- Uses ES6 modules (`import`/`export`)
- Configuration is loaded from `config.json` instead of `pyproject.toml`
- All class structures and method names are preserved

## Customizing the Classification Logic

The classification logic is in `src/anni/calculator/calculator.js`. The current implementation is a placeholder - modify the `_decide()` method to implement your own naughty/nice classification rules based on the person's events.


