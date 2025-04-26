
// Stores a list of decorated property keys for each Model subclass constructor
const decoratedPropertiesRegistry = new Map<Function, string[]>();

/**
 * Base class for models that automatically assigns data from a constructor object
 * to properties decorated with @property.
 */
export class Model {
    constructor(data: object) {
        const properties = this.decoratedPropertyKeys;
        for (const key in data) {
            if (properties.includes(key) && data.hasOwnProperty(key)) {
                // Type assertion needed because 'this' is implicitly 'any' here
                (this as any)[key] = data[key];
            } else {
                console.warn(`Property '${key}' from data is not a decorated property or not present on model ${this.constructor.name}`);
            }
        }
    }

    /**
     * Gets the list of property keys decorated with @property for this instance's class.
     * Collects keys from the entire prototype chain.
     */
    get decoratedPropertyKeys(): string[] {
        const keys: string[] = [];
        let currentProto = Object.getPrototypeOf(this);
        while (currentProto && currentProto !== Object.prototype) {
            const constructor = currentProto.constructor;
            const registeredKeys = decoratedPropertiesRegistry.get(constructor);
            if (registeredKeys) {
                // Add keys, avoiding duplicates if overridden in subclasses
                registeredKeys.forEach(key => {
                    if (!keys.includes(key)) {
                        keys.push(key);
                    }
                });
            }
            currentProto = Object.getPrototypeOf(currentProto);
        }
        return keys;
    }

    /**
     * Returns a JSON string representation of the model,
     * including only properties decorated with @property.
     */
    toString(): string {
        const json: { [key: string]: any } = {};
        const properties = this.decoratedPropertyKeys;

        properties.forEach((propertyKey: string) => {
            // Access the property via its public getter, not the backing field
            json[propertyKey] = (this as any)[propertyKey];
        });

        return JSON.stringify(json);
    }
}

export function property() {
    return function (target: any, propertyKey: string) {
        // Register the property key for this specific class constructor
        const constructor = target.constructor;
        if (!decoratedPropertiesRegistry.has(constructor)) {
            decoratedPropertiesRegistry.set(constructor, []);
        }

        // Avoid duplicate registration if decorator is somehow applied multiple times
        if (!decoratedPropertiesRegistry.get(constructor)!.includes(propertyKey)) {
            decoratedPropertiesRegistry.get(constructor)!.push(propertyKey);
        }

        // Define the name for the hidden backing field on the instance
        const backingFieldName = `_${propertyKey}`;

        const getter = function (this: Model): any { // Add type for 'this'
            return (this as any)[backingFieldName];
        };

        const setter = function (this: Model, newVal: any): void { // Add type for 'this'
            const currentVal = (this as any)[backingFieldName];

            // Add logic here to prevent unnecessary updates 
            if (newVal === currentVal) return;

            // Basic type handling/coercion (can be expanded)
            let processedValue = newVal; // Default to assigning as is
            switch (typeof newVal) {
                case 'string':
                    // Keep as string
                    break;
                case 'number':
                    // Ensure it's stored as a number
                    processedValue = Number(newVal);
                    break;
                case 'boolean':
                    // Ensure it's stored as a boolean
                    processedValue = Boolean(newVal);
                    break;
                case 'object':
                    // Handle null explicitly, otherwise assign object reference
                    if (newVal === null) {
                        processedValue = null;
                    } else {
                        // TODO: Decide if deep/shallow copy is needed for objects
                        console.log(`TODO: Assigning object to property ${propertyKey}:`, newVal);
                    }
                    break;
                case 'undefined':
                    processedValue = null;
                    break;
                default:
                    // Potentially throw error or handle unsupported types
                    console.warn(`Assigning value of unhandled type '${typeof newVal}' to property ${propertyKey}`);
            }
            (this as any)[backingFieldName] = processedValue;
        };

        // Define the property on the target
        Object.defineProperty(target, propertyKey, {
            get: getter,
            set: setter,
            enumerable: true, // Might want this to be true
            configurable: true // Might want this to be true            
        });
    };
}
