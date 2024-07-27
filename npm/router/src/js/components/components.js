// CSS
import './components.css';

// Classes
import { CanvasElement } from './CanvasElement.js';
import { CanvasContentElement } from './CanvasContentElement.js';
import { CanvasNavElement } from './CanvasNavElement.js';
import { IconElement } from './IconElement.js';
import { NavGroupElement } from './NavGroupElement.js';
import { NavItemElement } from './NavItemElement.js';
import { FormControlElement } from './FormControlElement.js';
import { FormSwitchElement } from './FormSwitchElement.js';

// Web Components
customElements.define(IconElement.localName, IconElement); // <c-icon>
customElements.define(CanvasElement.localName, CanvasElement); // <c-canvas>
customElements.define(CanvasContentElement.localName, CanvasContentElement); // <c-canvas-content>
customElements.define(CanvasNavElement.localName, CanvasNavElement); // <c-canvas-nav>
customElements.define(NavGroupElement.localName, NavGroupElement); // <c-nav-group>
customElements.define(NavItemElement.localName, NavItemElement); // <c-nav-item>
customElements.define(FormControlElement.localName, FormControlElement); // <c-form-control>
customElements.define(FormSwitchElement.localName, FormSwitchElement); // <c-form-switch>
