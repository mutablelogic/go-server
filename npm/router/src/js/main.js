
// Favicon
import favicon from '../assets/favicon.png';
import { setFavIcon } from './favicon.js';
import { Provider } from './core/provider.js';
import { Event } from './core/event.js';

window.addEventListener('load', () => {
  // Initialize the application here
  setFavIcon(favicon);

  /* Set the token from local storage */
  const token = localStorage.getItem("token");
  if (token) {
    document.getElementById('token').value = token;
  }

  /* Add the token to the request headers */
  const provider = new Provider('http://127.0.0.1/');
  provider.addEventListener(Event.ERROR, (e) => {
    console.error("Error:", e.message);
  });

  /* Add the token to the request headers */
  provider.addEventListener(Event.START, (e) => {
    let request = e.detail;
    const token = localStorage.getItem("token");
    if (e.detail && token) {
      request.headers.set('Authorization', `Bearer ${token}`);
    }
  });

  // Set the token from the input field
  document.getElementById('token').addEventListener('change', (e) => {
    localStorage.setItem("token", e.target.value);
  });

  // Fetch the data every 20 seconds
  try {
    provider.Fetch('/api/router/', null, 20 * 1000);
  } catch (e) {
    console.error('Error: ', e.message);
  }
});
