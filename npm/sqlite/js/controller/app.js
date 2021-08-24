import {
  Controller, Nav, Toast, Provider,
} from '@djthorpe/js-framework';

import Connection from '../model/sqlite/connection';
import Object from '../model/sqlite/object';

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

    // Connection provider returns schemas in database connection
    super.define('connection', new Provider(Connection, API_PREFIX));
    if (this.connection) {
      this.connection.addEventListener('provider:error', (sender, error) => {
        this.toast.show(error);
      });
      this.connection.addEventListener(['provider:added', 'provider:changed'], (sender, connection) => {
        // Set the current schema as the first one
        this.setActiveSchema(connection.schemas[1]);
        console.log(`connection added or changed: ${connection}`);
      });
      this.connection.addEventListener('provider:deleted', (sender, connection) => {
        console.log(`connection deleted: ${connection}`);
      });
    }

    // Database provider returns tables, indexes and views in database connection
    super.define('database', new Provider(Object, `${API_PREFIX}/`));
    if (this.database) {
      this.database.addEventListener('provider:error', (sender, error) => {
        this.toast.show(error);
      });
      this.database.addEventListener(['provider:added', 'provider:changed'], (sender, object) => {
        console.log(`object added or changed: ${object}`);
      });
      this.database.addEventListener('provider:deleted', (sender, database) => {
        console.log(`object deleted: ${database}`);
      });
    }
  }

  main() {
    // Request the connection data
    this.connection.request(null, null, API_FETCH_DELTA);
  }

  setActiveSchema(schema) {
    // Read in the table information
    if (this.database) {
      this.database.request(schema, null, API_FETCH_DELTA);
    }
  }
}
