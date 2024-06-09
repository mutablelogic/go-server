import { assertInstanceOf, assertTypeOf } from "./assert";

/**
 * @class
 * @name ModelClass
 * @description Represents a model class and its properties
 */
export class ModelClass extends Object {
  #name = ""; // The name of the property
  #properties = new Map(); // The properties

  static get localName() {
    return 'c-model-class';
  }

  constructor(name, properties) {
    super();
    assertTypeOf(name, 'string');
    assertInstanceOf(attrs, Array);
  }
}
