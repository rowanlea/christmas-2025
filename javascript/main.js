import { readFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';
import { z } from 'zod';
import { OccupationEnum, PersonSchema } from './src/anni/types/datatypes.js';
import { AutomatedNaughtyOrNiceAnalyser } from './src/anni/analyser.js';
import { TestHarness } from './src/testharness/testharness.js';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

function loadPeopleFromFile(path) {
  const fileContent = readFileSync(path, 'utf-8');
  const data = JSON.parse(fileContent);
  const peopleArraySchema = z.array(PersonSchema);
  return peopleArraySchema.parse(data);
}

async function main() {
  const args = process.argv.slice(2);
  const isTest = args.includes('--test');
  const isScore = args.includes('--score');

  if (isTest) {
    const harness = new TestHarness();
    const person = {
      id: 789433,
      name: "poo",
      age: 21,
      location: "lahndon",
      occupation: OccupationEnum.PROGRAMMER,
      favourite_katas: [],
      events: []
    };
    
    try {
      const result = await harness.checkElfLogic(person);
      console.log("Test Harness Result:", result);
    } catch (error) {
      console.error("Test failed:", error.message);
    }
  } else if (isScore) {
    const analyser = new AutomatedNaughtyOrNiceAnalyser();
    const dataFile = join(__dirname, 'src', 'data', 'validation-data.json');
    
    try {
      const people = loadPeopleFromFile(dataFile);
      const analysedPeople = analyser.analyse(people);
      const result = await analyser.submit(analysedPeople);
      console.log("Scoring Result:", result);
    } catch (error) {
      console.error("Scoring failed:", error.message);
      if (error.response) {
        console.error("Response data:", error.response.data);
      }
    }
  } else {
    console.log(`
Automated Naughty or Nice Analyzer

Usage:
  node main.js --test    Run in test harness mode
  node main.js --score   Run in scoring mode (WILL SUBMIT DATA TO SERVER FOR SCORING)
  
Options:
  --test     Test the elf logic with a sample person
  --score    Analyze validation data and submit for scoring
    `);
  }
}

main().catch(error => {
  console.error("Fatal error:", error);
  process.exit(1);
});


