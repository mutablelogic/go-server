import { AclItem } from './AclItem';
import { Model, property } from './Model';

export class Database extends Model {
    @property() oid: number;
    @property() name: string;
    @property() owner: string;
    @property() acl: AclItem[];
    @property() bytes: number;

    constructor(data: object) {
        super(data);
    }
}
