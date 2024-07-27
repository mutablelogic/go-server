import { assertNilOrInstanceOf } from '../core/assert.js';
import { Token } from './Token.js';

export class TokenWithValue extends Token {
  static get localName() {
    return Token.localName + '-with-value';
  }
  static get properties() {
    return {
      value: { type: String, jsonName: 'token' },
    };
  }
  constructor(data) {
    super(data);
    assertNilOrInstanceOf(data, Object);
  }  
}
