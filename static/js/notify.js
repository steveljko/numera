export default class {
  constructor(options = {}) {
    this.container = null;
    this.notifications = [];
    this.defaults = {
      offset: 32,
      duration: 3000,
      gap: 10,
      ...options
    };
    this.init();
  }

  init() {
      this.container = document.createElement('div');
      this.container.className = 'notify-container';
      this.container.style.bottom = `${this.defaults.offset}px`;
      this.container.style.gap = `${this.defaults.gap}px`;
      document.body.appendChild(this.container);
  }

  show(options) {
    const config = { ...this.defaults, ...options };
    const notification = this.createNotification(config);

    this.container.appendChild(notification);
    this.notifications.push(notification);

    requestAnimationFrame(() => notification.classList.add('notify-show'));

    // auto dismiss
    if (config.duration > 0) {
      setTimeout(() => this.dismiss(notification), config.duration);
    }

    return notification;
  }

  createNotification(config) {
    const el = document.createElement('div');
    el.className = `notify notify-${config.type}`;

    const icons = {
      success: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="m4.5 12.75 6 6 9-13.5" /></svg>',
      warning: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m-9.303 3.376c-.866 1.5.217 3.374 1.948 3.374h14.71c1.73 0 2.813-1.874 1.948-3.374L13.949 3.378c-.866-1.5-3.032-1.5-3.898 0L2.697 16.126ZM12 15.75h.007v.008H12v-.008Z" /></svg>',
      error: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9 3.75h.008v.008H12v-.008Z" /></svg>',
      info: '<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6"><path stroke-linecap="round" stroke-linejoin="round" d="m11.25 11.25.041-.02a.75.75 0 0 1 1.063.852l-.708 2.836a.75.75 0 0 0 1.063.853l.041-.021M21 12a9 9 0 1 1-18 0 9 9 0 0 1 18 0Zm-9-3.75h.008v.008H12V8.25Z" /></svg>'
    };

    let html = `
      <div class="notify-content">
        <div class="notify-icon">${icons[config.type]}</div>
        <div class="notify-body">
          <div class="notify-title">${config.title}</div>
          ${config.description ? `<div class="notify-description">${config.description}</div>` : ''}
        </div>
        <button class="notify-close">
            <svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor" class="size-6">
              <path stroke-linecap="round" stroke-linejoin="round" d="M6 18 18 6M6 6l12 12" />
            </svg>
        </button>
      </div>
    `;

    el.innerHTML = html;

    // click to dismiss action
    el.addEventListener('click', () => this.dismiss(el));
    el.querySelector('.notify-close').addEventListener('click', (e) => {
      e.stopPropagation();
      this.dismiss(el);
    });

    return el;
  }

  dismiss(notification) {
    notification.classList.remove('notify-show');
    notification.style.opacity = '0';
    notification.style.transform = 'translateY(20px)';

    setTimeout(() => {
      if (notification.parentNode) {
        notification.parentNode.removeChild(notification);
      }
      const idx = this.notifications.indexOf(notification);
      if (idx > -1) {
        this.notifications.splice(idx, 1);
      }
    }, 300);
  }

  success(title, description = '') {
    return this.show({ type: 'success', title, description });
  }

  warning(title, description = '') {
    return this.show({ type: 'warning', title, description });
  }

  error(title, description = '') {
    return this.show({ type: 'error', title, description });
  }

  info(title, description = '') {
    return this.show({ type: 'info', title, description });
  }

  clearAll() {
    this.notifications.forEach(n => this.dismiss(n));
  }
}
