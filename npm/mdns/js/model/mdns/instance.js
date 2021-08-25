import { Model } from '@djthorpe/js-framework';

export default class Instance extends Model {
  static define() {
    super.define(Instance, {
      name: 'string',
      service: 'string',
      zone: 'string',
      host: 'string',
      port: 'number',
      addrs: '[]string',
      txt: '{}string',
    }, 'Instance');
  }

  get key() {
    // Replace non-alphanumeric characters with underscores
    const name = `${this.name}.${this.service}${this.zone}`;
    return `i-${name.replace(/[^a-zA-Z0-9]/g, '_')}`;
  }
}

Instance.define();
