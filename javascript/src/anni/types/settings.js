import { readFileSync } from 'fs';
import { join, dirname } from 'path';
import { fileURLToPath } from 'url';

const __filename = fileURLToPath(import.meta.url);
const __dirname = dirname(__filename);

export class Settings {
  constructor() {
    try {
      const configPath = join(__dirname, '../../../config.json');
      const configData = JSON.parse(readFileSync(configPath, 'utf-8'));
      
      this.team_name = configData.team_name || "We didn't read the instructions";
      this.base_url = configData.base_url || "http://localhost:8000";
      this.team_secret = configData.team_secret || "LMAO";
    } catch (error) {
      console.warn('Could not load config.json, using defaults');
      this.team_name = "We didn't read the instructions";
      this.base_url = "http://localhost:8000";
      this.team_secret = "LMAO";
    }
  }
}


