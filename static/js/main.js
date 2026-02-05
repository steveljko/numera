import htmx from 'htmx.org';
import Alpine from 'alpinejs';
import Notify from './notify';
const notify = new Notify;

window.Alpine = Alpine;
Alpine.start();

htmx.on('toast', (e) => {
  const { type, text, desc } = e.detail;
  notify[type](text, desc);
});
