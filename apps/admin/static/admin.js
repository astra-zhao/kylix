// admin.js — progressive enhancement for KylixAdmin (v0.11.0 P4).
// No build step, no framework: everything here is optional sugar over the
// server-rendered pages, so the app works with JavaScript disabled.
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
})();
