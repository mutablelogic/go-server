import { Model } from '@djthorpe/js-framework';

export default class Event extends Model {
  static define() {
    super.define(Event, {
      name: 'string',
      id: 'number',
      ts: 'date',
    }, 'Group');
  }

  get key() {
    // Replace non-alphanumeric characters with underscores
    return `e-${this.id}`;
  }
}

Event.define();
