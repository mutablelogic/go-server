import { assertInstanceOf, assertTypeOf } from "./assert";

/**
 * @class
 * @name ModelProperty 
 * @description Represents a model property
 */
export class ModelProperty extends Object {
  #name; // The name of the property
  #type; // The type of the property
  #elem; // The element type of the property where the type is an array or map
  #json; // The name of the property in JSON for reading and writing
  #readonly; // Whether the property is read-only

  static get localName() {
    return 'c-model-property';
  }

  constructor(name, attrs) {
    super();
    assertTypeOf(name, 'string');
    assertInstanceOf(attrs, Object);

    // Set defaults
    this.#name = name;
    this.#type = attrs.type;
    this.#elem = attrs.elem || null;
    this.#json = attrs.json || name;
    this.#readonly = attrs.readonly || false;
  }

  defineProperty(proto, getter, setter) {
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
      set: this.#readonly ? undefined : setter,
    });
  }

  get key() {
    return this.#name;
  }

  toString() {
    return `{ModelProperty.localName} name: {this.#name}`;
  }
}
