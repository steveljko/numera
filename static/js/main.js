import htmx from 'htmx.org';
import Alpine from 'alpinejs';
import Notify from './notify';
const notify = new Notify;

window.Alpine = Alpine;
Alpine.start();

const modal = document.getElementById("modal")

function closeModal() {
  modal.classList.add("hidden")
  document.getElementById("dialog").innerHTML = ""
}

function openModal() {
  modal.classList.remove("hidden")
}

window.closeModal = closeModal

htmx.on("htmx:afterSwap", (e) => {
  if (e.detail.target.id == "dialog") {
    openModal()
  }
})

htmx.on("htmx:beforeSwap", (e) => {
  if (e.detail.target.id == "dialog" && !e.detail.xhr.response) {
    closeModal()
    e.detail.shouldSwap = false
  }
})

modal?.addEventListener("click", (e) => {
  if (e.target === modal) {
    closeModal()
  }
})

htmx.on('toast', (e) => {
  const { type, text, desc } = e.detail;
  notify[type](text, desc);
});
