import { assertInstanceOf, assertNilOrInstanceOf } from "./assert.js";
import { ModelProperty } from "./ModelProperty.js";

/**
 * @class
 * @name Model 
 * @description Represents a data model
 */
export class Model extends EventTarget {
  // Define the values of the object
  #values = new Map();

  static get localName() {
    return 'c-model';
  }

  constructor(data) {
    super();
    assertNilOrInstanceOf(data, Object);

    // Create the getter and setter for each property
    let proto = this.constructor;
    while (proto && proto !== Object) {
      const name = proto.localName || proto.name;
      if (name && proto.properties && proto.properties instanceof Object) {
        // Create properties
        const properties = Object.entries(proto.properties).map(([key, value]) => {
          return new ModelProperty(key, value);
        });

        // Define the properties on the prototype
        properties.forEach((property) => {
          property.defineProperty(proto.prototype, this.#get.bind(this, property), this.#set.bind(this, property));
        });
      }

      // Descend the prototype chain
      proto = Object.getPrototypeOf(proto);
    }

    // If no data is provided, return
    if (!data) {
      return;
    }

    // Set the property values
    Object.entries(data).forEach(([key, value]) => {      
      this.#values.set(key, value);
    });
  }

  toString() {
    return `{Model.localName} name: {this.#name}`;
  }

  valueOf() {
    // TODO: 
    // Should return the primary index key of the model, to compare one object to another
  }

  equals(other) {
    // TODO:
    // Should return true if the other object is equal to this object, comparing
    // all properties of the object
  }

  // Getter for each property
  #get(property) {
    assertInstanceOf(property, ModelProperty);
    return this.#values.get(property.key);
  }

  // Setter for each property
  #set(property, value) {
    assertInstanceOf(property, ModelProperty);
    console.log(`Setting ${property.key} to ${value}`);
    return this.#values.set(property.key, value);
  }
}
