import { property } from './Model';
import { DatabaseMeta } from './DatabaseMeta';

export class Database extends DatabaseMeta {
    @property() oid: number;
    @property() bytes: number;

    constructor(data: object) {
        super(data);
    }
}
