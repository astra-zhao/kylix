// admin.js — progressive enhancement for KylixAdmin (v0.11.0 P5).
// No build step, no framework: every page works with JavaScript disabled, this
// file only adds convenience on top of the server-rendered HTML.
(function () {
  'use strict';

  // Delete confirmation: forms carrying data-confirm ask before submitting.
  document.addEventListener('submit', function (ev) {
    var form = ev.target;
    if (form && form.dataset && form.dataset.confirm) {
      if (!window.confirm(form.dataset.confirm)) {
        ev.preventDefault();
      }
    }
  });

  // Sidebar collapse, remembered per browser.
  var toggle = document.querySelector('[data-sidebar-toggle]');
  try {
    if (window.localStorage && localStorage.getItem('kyadmin.sidebar') === 'collapsed') {
      document.body.classList.add('sidebar-collapsed');
    }
  } catch (e) { /* private mode: ignore */ }
  if (toggle) {
    toggle.addEventListener('click', function () {
      document.body.classList.toggle('sidebar-collapsed');
      try {
        if (window.localStorage) {
          localStorage.setItem('kyadmin.sidebar',
            document.body.classList.contains('sidebar-collapsed') ? 'collapsed' : 'open');
        }
      } catch (e) { /* ignore */ }
    });
  }

  // Theme: the server already rendered the right palette from the cookie; this
  // only extends the session cookie's life so the choice survives a restart.
  var themeLink = document.querySelector('[data-theme-toggle]');
  if (themeLink) {
    themeLink.addEventListener('click', function () {
      var want = themeLink.getAttribute('data-theme-toggle');
      try {
        document.cookie = 'theme=' + want + '; Path=/; Max-Age=31536000; SameSite=Lax';
      } catch (e) { /* ignore */ }
    });
  }

  // Avatar picker: read the chosen file in the browser and drop a data URL
  // into the hidden field, so the upload is a plain urlencoded POST (the LLVM
  // backend rejects multipart at its CSRF gate).
  var file = document.querySelector('[data-avatar-input]');
  var hidden = document.getElementById('avatar_data');
  if (file && hidden && window.FileReader) {
    file.addEventListener('change', function () {
      var f = file.files && file.files[0];
      if (!f) { return; }
      var reader = new FileReader();
      reader.onload = function () { hidden.value = String(reader.result || ''); };
      reader.readAsDataURL(f);
    });
  }

  // Toasts: turn server-side flash messages into transient notifications, and
  // auto-dismiss anything pushed later through window.kyadminToast.
  var host = document.querySelector('[data-toasts]');
  function toast(text, kind) {
    if (!host || !text) { return; }
    var el = document.createElement('div');
    el.className = 'toast ' + (kind || '');
    el.textContent = text;
    host.appendChild(el);
    setTimeout(function () { el.remove(); }, 4000);
  }
  window.kyadminToast = toast;
  var flash = document.querySelector('.flash');
  if (flash) {
    toast(flash.textContent.trim(), flash.classList.contains('bad') ? 'bad' : 'good');
  }
})();
