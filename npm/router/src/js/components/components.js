// CSS
import './components.css';

// Classes
import { IconElement } from './IconElement.js';
import { DivElement } from './DivElement.js';

// Web Components
customElements.define(IconElement.localName, IconElement); // <c-icon>
customElements.define(DivElement.localName, DivElement); // <c-div>
