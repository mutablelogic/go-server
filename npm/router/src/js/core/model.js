import { assertInstanceOf } from "./assert.js";
import { ModelProperty } from "./ModelProperty.js";

/**
 * @class
 * @name Model 
 * @description Represents a model in the MVC pattern
 */
export class Model extends EventTarget {
  static get localName() {
    return 'c-model';
  }

  constructor(data) {
    super();

    // Populate the classes map with the properties of the model
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

    if (data) {
      console.log('TODO: Model#constructor', data);
    }
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

  // Define a getter for each property
  #get(property) {
    return this[`_${property.key}`];
  }

  // Define a setter for each property
  #set(property, value) {
    this[`_${property.key}`] = value;    
  }
}
