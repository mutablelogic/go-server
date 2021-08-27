import { Model } from '@djthorpe/js-framework';

export default class Instance extends Model {
  static define() {
    super.define(Instance, {
      name: 'string',
      instance: 'string',
      service: 'Service',
      zone: 'string',
      host: 'string',
      port: 'number',
      addrs: '[]string',
      txt: '{}string',
    }, 'Instance');
  }

  get key() {
    // Replace non-alphanumeric characters with underscores
    return `i-${this.instance.replace(/[^a-zA-Z0-9]/g, '_')}`;
  }
}

Instance.define();
