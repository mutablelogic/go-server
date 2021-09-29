import { Model } from '@djthorpe/js-framework';

export default class Service extends Model {
  static define() {
    super.define(Service, {
      service: 'string',
      description: 'string',
      note: 'string',
    }, 'Service');
  }

  get key() {
    // Replace non-alphanumeric characters with underscores
    return `s-${this.service.replace(/[^a-zA-Z0-9]/g, '_')}`;
  }
}

Service.define();
