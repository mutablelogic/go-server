
// Global styles
import '../css/vars.css';
import '../css/document.css';

// Core
import Controller from './core/controller';
import Provider from './core/provider';
import Model from './core/model';
import Event from './core/event';

// Icons
import './icon/wc-icon';

// Navigation Web Components
import './nav/wc-nav';
import './nav/wc-navbar';
import './nav/wc-nav-item';

// Import favicon
import icon from '../assets/favicon/mu-756x756.png';

// Set favicon
const link = document.querySelector("link[rel~='icon']");
if (link) {
  link.href = icon;
}

// Exports
export {
    Model, Controller, Provider, Event,
};

