import {
  Controller, Nav, Toast, Provider,
} from '@djthorpe/js-framework';

import Event from '../model/events/event';

const API_PREFIX = '/api/events';
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

    // Define event provider
    this.define('service', new Provider(Event, API_PREFIX));
    this.service.addEventListener('provider:error', (sender, error) => {
      this.toast.show(error);
    });
    this.service.addEventListener(['provider:added', 'provider:changed'], (sender, event) => {
      console.log(`added or changed: ${event}`);
    });
    this.service.addEventListener(['provider:deleted'], (sender, event) => {
      console.log(`deleted: ${event}`);
    });
  }

  main() {
    // Request the events data
    this.service.request('/', null, API_FETCH_DELTA);
  }
}
