import {
  Offcanvas as BSOffcanvas, List
} from '@djthorpe/js-framework';
import Node from './node';

export default class Offcanvas extends BSOffcanvas {
  constructor(node) {
    super(node);
    this.$instance = undefined;

    // Add TXT
    const nodeTxt = this.query('#_txt');
    if (nodeTxt) {
      this.$txt = new List(nodeTxt, '_template');
    }

    // Add Addrs
    const nodeAddrs = this.query('#_addrs');
    if (nodeAddrs) {
      this.$addrs = new List(nodeAddrs, '_template');
    }

    // Add event listeners
    this.addEventListener(['offcanvas:hide'], () => {
      this.instance = undefined;
    });
  }

  get instance() {
    return this.$instance;
  }

  set instance(v) {
    this.$instance = v;
  }

  show(instance) {
    super.show();
    this.instance = instance;

    // Populate instance table
    this
      .replace('._name', Node.div('', Node.strong('', instance.name)), Node.div('', Node.small('', Node.badge('bg-secondary', instance.service.description || instance.service.service))))
      .replace('._note', instance.service.note || '')
      .replace('._service', Node.small('', instance.service.service))
      .replace('._zone', Node.small('', instance.zone ? instance.zone : ''))
      .replace('._host', Node.small('', instance.host ? instance.host : ''))
      .replace('._port', Node.small('', instance.port ? instance.port : ''));

    // Hide note if empty
    this.query('._note').style.display = instance.service.note ? 'block' : 'none';

    // Populate addrs and txt
    if (this.$addrs) {
      this.showAddrs(instance);
    }
    if (this.$txt) {
      this.showTxt(instance);
    }
  }

  showAddrs(instance) {
    if (instance.addrs && instance.addrs.length) {
      this.$addrs.clear();
      instance.addrs.forEach((value, index) => {
        this.$addrs.set(`a-${index}`)
          .replace('._value', Node.small('', value));
      });
    }
  }

  showTxt(instance) {
    let show = false;
    if (instance.txt && instance.txt.size) {
      this.$txt.clear();
      instance.txt.forEach((value, key) => {
        if (key) {
          show = true;
          this.$txt.set(key)
            .replace('._key', Node.badge('bg-primary', key))
            .replace('._value', Node.small('', value));
        }
      });
    }
    if (show) {
      this.$txt.parent.show();
    } else {
      this.$txt.parent.hide();
    }
  }
}
