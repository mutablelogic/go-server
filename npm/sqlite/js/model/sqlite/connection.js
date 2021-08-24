import { Model } from '@djthorpe/js-framework';

export default class Connection extends Model {
  static define() {
    super.define(Connection, {
      version: 'version string',
      modules: 'modules []string',
      schemas: 'schemas []string',
    }, 'Connection');
  }
}

Connection.define();
