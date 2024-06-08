/* Code to reload in the esbuild serve development environment */
window.addEventListener('load', () => {
  // eslint-disable-next-line no-restricted-globals
  new EventSource('/esbuild').addEventListener('change', () => location.reload());
});
