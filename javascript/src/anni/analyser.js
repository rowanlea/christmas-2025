import { AutomatedNaughtyOrNiceCalculator } from './calculator/calculator.js';
import { RequestHandler } from './requests/requests.js';

export class AutomatedNaughtyOrNiceAnalyser {
  constructor(calculator = null, handler = null) {
    this.calculator = calculator || new AutomatedNaughtyOrNiceCalculator();
    this.handler = handler || new RequestHandler();
  }

  analyse(people) {
    return this.calculator.classifyPeople(people);
  }

  async submit(scoredPeople) {
    return await this.handler.submitList(scoredPeople);
  }
}


