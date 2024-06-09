import { assertNilOrInstanceOf } from '../core/assert.js';
import { Model } from '../core/Model.js';

export class Token extends Model {
  static get localName() {
    return Model.localName + '-token';
  }
  static get properties() {
    return {
      name: { type: String },
      accessTime: { type: Date, jsonName: 'access_time' },
      scopes: { type: Array, elem: String },
      valid: { type: Boolean },
    };
  }
  constructor(data) {
    super(data);
    assertNilOrInstanceOf(data, Object);
  }
}
