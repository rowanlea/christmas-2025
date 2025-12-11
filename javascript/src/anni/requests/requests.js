import axios from 'axios';
import { Settings } from '../types/settings.js';
import { 
  ScoreElfLogicRequestSchema, 
  ScoreElfLogicResponseSchema 
} from './responsetypes.js';
import { NaughtyOrNiceClassificationEnum } from '../types/datatypes.js';
import { z } from 'zod';

export class RequestHandler {
  constructor(settings = null, session = null) {
    this.settings = settings || new Settings();
    this.session = session || axios.create();
  }

  _getTeamName() {
    if (!this.settings.team_name) {
      throw new Error("Team name is not set in settings.");
    }
    return this.settings.team_name;
  }

  _getTeamSecret() {
    if (!this.settings.team_secret) {
      throw new Error("Unique key is not set in settings.");
    }
    return this.settings.team_secret;
  }

  _getAuthHeaders() {
    return {
      "X-Team-Name": this._getTeamName(),
      "X-Unique-Key": this._getTeamSecret(),
      "Content-Type": "application/json",
      "Accept": "application/json"
    };
  }

  _url(endpoint) {
    const baseUrl = this.settings.base_url.replace(/\/$/, '');
    const cleanEndpoint = endpoint.replace(/^\//, '');
    return `${baseUrl}/${cleanEndpoint}`;
  }

  async _post(endpoint, jsonData) {
    try {
      const response = await this.session.post(
        this._url(endpoint),
        jsonData,
        {
          headers: this._getAuthHeaders(),
          timeout: 10000
        }
      );
      return response.data;
    } catch (error) {
      if (error.response) {
        throw new Error(`HTTP ${error.response.status}: ${error.response.statusText}`);
      }
      throw error;
    }
  }

  async _getSingle(endpoint, jsonData) {
    try {
      const response = await this.session.request({
        method: 'GET',
        url: this._url(endpoint),
        data: jsonData,
        headers: this._getAuthHeaders(),
        timeout: 10000
      });
      return response.data;
    } catch (error) {
      if (error.response) {
        throw new Error(`HTTP ${error.response.status}: ${error.response.statusText}`);
      }
      throw error;
    }
  }

  // Public API
  async checkElfLogic(person) {
    const endpoint = "naughtyornice";
    const requestData = { ...person };
    const parsedResponse = await this._getSingle(endpoint, requestData);
    
    // Validate response is a valid NaughtyOrNiceClassificationEnum value
    const enumSchema = z.enum([
      NaughtyOrNiceClassificationEnum.NONE,
      NaughtyOrNiceClassificationEnum.NAUGHTY,
      NaughtyOrNiceClassificationEnum.NICE
    ]);
    
    return enumSchema.parse(parsedResponse);
  }

  async submitList(people) {
    const endpoint = "score";
    const requestData = {
      naughtyNiceCandidateRequests: people.map(p => ({
        id: p.id,
        naughtyOrNice: p.naughty_or_nice
      }))
    };
    
    // Validate request data
    ScoreElfLogicRequestSchema.parse(requestData);
    
    const response = await this._post(endpoint, requestData);
    const parsedResponse = ScoreElfLogicResponseSchema.parse(response);
    
    return parsedResponse
  }
}


