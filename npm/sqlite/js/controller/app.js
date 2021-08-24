import {
  Controller, Nav, Toast, Provider,
} from '@djthorpe/js-framework';
import Connection from '../model/sqlite/connection';

const API_PREFIX = '/api/sqlite';
const API_FETCH_DELTA = 60 * 1000;

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

    // Define providers
    super.define('connection', new Provider(Connection, API_PREFIX));
    if (this.connection) {
      this.connection.addEventListener('provider:error', (sender, error) => {
        this.toast.show(error);
      });
      this.connection.addEventListener(['provider:added', 'provider:changed'], (sender, connection) => {
        console.log(`connection added or changed: ${connection}`);
      });
      this.connection.addEventListener('provider:deleted', (sender, connection) => {
        console.log(`connection deleted: ${connection}`);
      });
    }
  }

  // eslint-disable-next-line class-methods-use-this
  main() {
    this.connection.request(null, null, API_FETCH_DELTA);
  }
}
