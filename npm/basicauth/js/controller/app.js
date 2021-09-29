import {
  Controller, Nav, Toast, Provider, List,
} from '@djthorpe/js-framework';
import Service from '../model/basicauth/service';
import Node from '../view/node';

const API_PREFIX = '/api/basicauth';
const API_FETCH_DELTA = 10 * 1000;

export default class App extends Controller {
  constructor() {
    super();


    // Define views, providers for page elements
    const navNode = document.querySelector('#nav');
    if (navNode) {
      super.define('nav', new Nav(navNode));
    }
    const toastNode = document.querySelector('#toast');
    if (toastNode) {
      super.define('toast', new Toast(toastNode));
    }

    // Get users and groups
    this.define('service', new Provider(Service, API_PREFIX));
    this.service.addEventListener('provider:error', (sender, error) => {
      this.toast.show(error);
    });
    this.service.addEventListener(['provider:added', 'provider:changed'], (sender, service) => {
      console.log(`added or changed: ${service}`);
      if (service.users && this.users) {
        this.users.clear();
        service.users.forEach((user) => {
          this.users.set(user).replace('._name', Node.badge('bg-primary', user));
        });
      }

      if (service.groups && this.groups) {
        this.groups.clear();
        service.groups.forEach((group) => {
          this.groups.set(group.key).replace('._name', group.name);
          this.groups.set(group.key).replace('._members', ...group.members);
        });
      }
    });

    // Define view of users
    const usersNode = document.querySelector('#users tbody');
    if (usersNode) {
      super.define('users', new List(usersNode, '_template'));
    }

    // Define view of groups
    const groupsNode = document.querySelector('#groups tbody');
    if (groupsNode) {
      super.define('groups', new List(groupsNode, '_template'));
    }
  }

  main() {
    // Request the connection data
    this.service.request('/', null, API_FETCH_DELTA);
  }
}
