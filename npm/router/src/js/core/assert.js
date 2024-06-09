

export function assertTypeOf(value, type) {
    if (typeof value !== type) {
        throw new TypeError(`Expected ${type}, got ${typeof value}`);
    }
}

export function assertNilOrTypeOf(value, type) {
    if (value === null || value === undefined) {
        return;
    }
    assertTypeOf(value, type);
}

export function assertInstanceOf(value, instance) {
    if (!(value instanceof instance)) {
        throw new TypeError(`Expected instance of ${instance}, got ${value.constructor.name}`);
    }
}

export function assertNilOrInstanceOf(value, instance) {
    if (value === null || value === undefined) {
        return;
    }
    assertInstanceOf(value, instance);
}
