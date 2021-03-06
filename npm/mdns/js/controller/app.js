import {
  Controller, Nav, Toast, Provider, List, Button, Form,
} from '@djthorpe/js-framework';

import Instance from '../model/mdns/instance';
import Service from '../model/mdns/service';
import Node from '../view/node';
import Offcanvas from '../view/offcanvas';

const API_PREFIX = '/api/mdns';
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

    // Instance provider returns instances
    super.define('instances', new Provider(Instance, API_PREFIX));
    this.instances.addEventListener('provider:error', (sender, error) => {
      this.toast.show(error);
    });
    this.instances.addEventListener(['provider:added', 'provider:changed'], (sender, instance) => {
      console.log(`added or changed: ${instance}`);
      if (this.list) {
        const row = this.list.set(instance.key);
        const tags = instance.txt ? Array.from(instance.txt, (k) => {
          if (k[0] && k[1]) {
            return Node.badge('bg-primary', `${k[0]}: ${k[1]}`);
          }
          if (k[0]) {
            return Node.badge('bg-primary', `${k[0]}`);
          }
          return '';
        }) : [];
        row
          .replace('._name', Node.div('', Node.strong('', instance.name)), Node.div('', Node.small('', Node.badge('bg-secondary', instance.service.description || instance.service.service))))
          .replace('._host', Node.small('', instance.host && instance.port ? `${instance.host}:${instance.port}` : ''))
          .replace('._txt', ...tags);
      }
      // Update detail if view shows changed instance
      if (this.detail && this.detail.instance) {
        if (instance.key === this.detail.instance.key) {
          this.detail.show(instance);
        }
      }
    });
    this.instances.addEventListener('provider:deleted', (sender, instance) => {
      console.log(`deleted: ${instance}`);
      if (this.list) {
        this.list.deleteForKey(instance.key);
      }
      // Hide detail if view shows deleted instance
      if (this.detail && this.detail.instance) {
        if (instance.key === this.detail.instance.key) {
          this.detail.hide();
        }
      }
    });
    this.instances.addEventListener('provider:completed', () => {
      if (this.list) {
        this.list.sortForKeys(this.instances.keys.sort((a, b) => {
          const namea = this.instances.objectForKey(a).name.toLowerCase();
          const nameb = this.instances.objectForKey(b).name.toLowerCase();
          return namea.localeCompare(nameb);
        }));
      }
    });

    // Define view of instances
    const listNode = document.querySelector('#instances tbody');
    if (listNode) {
      super.define('list', new List(listNode, '_template'));
      this.list.addEventListener('list:click', (sender, target, key) => {
        const instance = this.instances.objectForKey(key);
        if (instance && this.detail) {
          this.detail.show(instance);
        }
      });
    }

    // Define the detail view
    const detailNode = document.querySelector('#offcanvas');
    if (detailNode) {
      super.define('detail', new Offcanvas(detailNode));
    }

    // Actions
    const actionServicesNode = document.querySelector('#action-services');
    if (actionServicesNode) {
      super.define('actionservices', new Button(actionServicesNode));
      this.actionservices.addEventListener('button:click', () => {
        if (this.modalservices) {
          this.modalservices.show();
        }
      });
    }
    const actionInstancesNode = document.querySelector('#action-instances');
    if (actionInstancesNode) {
      super.define('actioninstances', new Button(actionInstancesNode));
      this.actioninstances.addEventListener('button:click', () => {
        console.log('click-instances');
      });
    }

    // Modals
    const modalServicesNode = document.querySelector('#modal-services');
    if (modalServicesNode) {
      super.define('modalservices', new Form(modalServicesNode));
      this.modalservices.addEventListener('form:show', () => {
        this.instances.do('/', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/x-www-form-urlencoded',
          },
          body: new URLSearchParams({
            timeout: '2s',
          }), // body data type must match "Content-Type" header
        });
      });
    }
  }

  main() {
    // Request the connection data
    this.instances.request('/', null, API_FETCH_DELTA);
  }
}
