import {
  Controller, Nav, Toast, Provider, List, Button, Form,
} from '@djthorpe/js-framework';

const API_PREFIX = '/api/basicauth';
const API_FETCH_DELTA = 10 * 1000;

export default class App extends Controller {
  constructor() {
    super();
    this.define('basicauth',new Provider('basicauth',API_PREFIX));
  }

  main() {
    // Request the connection data
    this.basicauth.request('/', null, API_FETCH_DELTA);
  }
}
