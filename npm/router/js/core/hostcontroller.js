/*
    js-components
    Web Components Design System
    github.com/mutablelogic/js-components
*/

import Event from './event';
import Model from './model';

/**
 * HostController binds a data source to a host web component.
 * @class
*/
export default class HostController {
  constructor(host, provider, renderer) {
    // Store a reference to the host and provider
    this.$host = host;
    this.$provider = provider;
    this.$renderer = renderer;

    // Register for lifecycle updates
    host.addController(this);
  }

  hostConnected() {
    this.$provider.addEventListener(Event.ADD, (e) => {
      const view = this.$host.add(e.detail);
      if (this.$renderer) {
        this.$renderer(e.detail, view);
      }
    });
    this.$provider.addEventListener(Event.CHANGE, (e) => {
      const view = this.$host.change(e.detail);
      if (this.$renderer) {
        this.$renderer(e.detail, view);
      }
    });
    this.$provider.addEventListener(Event.DELETE, (e) => {
      const view = this.$host.delete(e.detail);
      console.log(`deleted ${e.detail} => ${view}`);
    });
  }

  // eslint-disable-next-line class-methods-use-this
  hostDisconnected() {
    // TODO: Remove event listeners
  }
}
