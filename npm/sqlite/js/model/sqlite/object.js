import { Model } from '@djthorpe/js-framework';

export default class Object extends Model {
  static define() {
    super.define(Object, {
      schema: 'string',
      name: 'string',
      table: 'string',
      type: 'string',
      sql: 'string',
    }, 'Object');
  }
}

Object.define();
