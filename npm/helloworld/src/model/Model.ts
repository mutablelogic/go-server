
export class Model {
    constructor(data: object) {
        console.log('Model constructor', data);
    }
}

export function property() {
    return function (target: any, propertyKey: string) {
        console.log("property(): called", target, propertyKey);
    };
}
