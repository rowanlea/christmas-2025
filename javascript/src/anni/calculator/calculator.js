import { NaughtyOrNiceClassificationEnum } from '../types/datatypes.js';

export class AutomatedNaughtyOrNiceCalculator {
  constructor() {
    // Initialization if needed
  }

  classifyPeople(people) {
    return people.map(person => this.classify(person));
  }

  classify(person) {
    // Placeholder logic for classification; replace with real rules.
    const decision = this._decide(person.events);
    return {
      ...person,
      naughty_or_nice: decision
    };
  }

  _decide(events) {
    // Very naive example: any events -> NICE else NAUGHTY (adjust later)
    if (events && events.length > 0) {
      return NaughtyOrNiceClassificationEnum.NICE;
    }
    return NaughtyOrNiceClassificationEnum.NAUGHTY;
  }
}


