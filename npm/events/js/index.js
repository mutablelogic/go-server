// CSS
import '../css/index.css';

// Application Controllers
import App from './controller/app';

// Import favicon
import icon from '../assets/favicon/mu-756x756.png';

// Set favicon
const link = document.querySelector("link[rel~='icon']");
if (link) {
  link.href = icon;
}

// Import js-framework
const jsf = require('@djthorpe/js-framework');

// Run
window.addEventListener('DOMContentLoaded', () => {
  const app = jsf.Controller.New(App);
  console.log('Running application', app.constructor.name);
  app.main();
});
