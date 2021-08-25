import {
  Controller, Nav, Toast, Provider, List, Offcanvas
} from '@djthorpe/js-framework';

import Instance from '../model/mdns/instance';
import Node from '../view/node';

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
        const tags = instance.txt ? Array.from(instance.txt,k => {
          return Node.badge('bg-primary',`${k[0]}: ${k[1]}`);
        }) : [];
        row
        .replace('._name',Node.div('',Node.strong('',instance.name)),Node.div('',Node.small('',Node.badge('bg-secondary',instance.service))))
        .replace('._host',Node.small('',instance.host && instance.port ? `${instance.host}:${instance.port}` : ''))
        .replace('._txt',...tags)
      }
    });
    this.instances.addEventListener('provider:deleted', (sender, instance) => {
      console.log(`deleted: ${instance}`);
      if (this.list) {
        this.list.deleteForKey(instance.key);
      }
    });
    this.instances.addEventListener('provider:completed', (sender) => {
      if (this.list) {
        this.list.sortForKeys(this.instances.keys.sort((a, b) => {
          const namea = this.instances.objectForKey(a).name.toLowerCase() ;
          const nameb = this.instances.objectForKey(b).name.toLowerCase() ;
          return namea.localeCompare(nameb);
        }));
      }
    });

    // Define view of instances
    const listNode = document.querySelector('#instances tbody');
    if (listNode) {
      super.define('list', new List(listNode, '_template'));
      this.list.addEventListener('list:click', (sender, target,key) => {  
        const instance = this.instances.objectForKey(key); 
        if (instance && this.detail) {
          this.detail.show();
        }
      });      
     }

     // Define the detail view
     const detailNode = document.querySelector('#offcanvas');
     if (detailNode) {
       super.define('detail', new Offcanvas(detailNode));
     } 
  }

  main() {
    // Request the connection data
    this.instances.request(null, null, API_FETCH_DELTA);
  }
}
