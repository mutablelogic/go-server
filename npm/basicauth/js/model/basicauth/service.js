import { Model } from '@djthorpe/js-framework';
// eslint-disable-next-line no-unused-vars
import Group from './group';

export default class Service extends Model {
  static define() {
    super.define(Service, {
      users: '[]string',
      groups: '[]Group',
    }, 'Service');
  }
}

Service.define();
