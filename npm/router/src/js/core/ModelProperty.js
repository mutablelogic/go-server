import { assertInstanceOf, assertTypeOf } from "./assert";

/**
 * @class
 * @name ModelProperty 
 * @description Represents a model property
 */
export class ModelProperty extends Object {
  #name;
  #type;
  #jsonName;

  static get localName() {
    return 'c-model-property';
  }

  constructor(name, data) {
    super();
    assertTypeOf(name, 'string');
    assertInstanceOf(data, Object);

    // Set defaults
    this.#name = name;
    this.#type = data.type;
    this.#jsonName = data.jsonName || name;
  }

  defineProperty(proto,getter,setter) {
    assertInstanceOf(proto, Object);

    // Check if the property is already defined in this class
    if (proto.hasOwnProperty(this.#name)) {
      return;
    }

    // Define the property
    Object.defineProperty(proto, this.#name, {
      enumerable: true,
      configurable: true,
      get: getter,
      set: setter,
    });
  }

  get key() {
    return this.#name;
  }

  toString() {
    return `{ModelProperty.localName} name: {this.#name}`;
  }
}
