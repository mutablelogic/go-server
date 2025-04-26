import { AclItem } from './AclItem';

export class Database  {
    oid: number;
    name: string;
    owner : string;
    acl: AclItem[];
    bytes: number;
}
