import { AclItem } from './AclItem';
import { Model, property } from './Model';

export class DatabaseMeta extends Model {
    @property() name: string;
    @property() owner: string;
    @property() acl: AclItem[];

    constructor(data: object) {
        super(data);
    }
}
