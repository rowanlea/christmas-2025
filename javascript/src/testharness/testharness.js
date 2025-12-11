import { RequestHandler } from '../anni/requests/requests.js';

export class TestHarness {
  constructor(handler = null) {
    this.handler = handler || new RequestHandler();
  }

  async checkElfLogic(person) {
    return await this.handler.checkElfLogic(person);
  }
}


