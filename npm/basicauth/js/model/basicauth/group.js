import { Model } from '@djthorpe/js-framework';

export default class Group extends Model {
  static define() {
    super.define(Group, {
      name: 'string',
      members: '[]string',
    }, 'Group');
  }

  get key() {
    // Replace non-alphanumeric characters with underscores
    return `g-${this.name.replace(/[^a-zA-Z0-9]/g, '_')}`;
  }
}

Group.define();
