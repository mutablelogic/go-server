import { Model, property } from './Model';

export class AclItem extends Model {
  @property() role: string;
  @property() priv: string[];

  constructor(data: object) {
    super(data);
  }
}
