// sw.js — KylixAdmin service worker (v0.13.0).
//
// Scope note: registered from /static/sw.js, so its scope is /static/ — it
// caches the console's static assets (CSS/JS/icons/manifest) and nothing else.
// Pages are network-only by design: a stale page would carry a stale CSRF
// token, and the login/admin pages must never be served from a cache.
const CACHE = 'kyadmin-v1';
const ASSETS = [
  '/static/admin.css',
  '/static/admin.js',
  '/static/manifest.json',
  '/static/icons/icon-192.png',
  '/static/icons/icon-512.png'
];

self.addEventListener('install', function (ev) {
  ev.waitUntil(caches.open(CACHE).then(function (c) { return c.addAll(ASSETS); }));
});

self.addEventListener('activate', function (ev) {
  ev.waitUntil(caches.keys().then(function (keys) {
    return Promise.all(keys.filter(function (k) { return k !== CACHE; })
      .map(function (k) { return caches.delete(k); }));
  }));
});

// Same-origin /static/ GET only: cache-first, fill on miss. Everything else
// bypasses the worker entirely (pages, POSTs, cross-origin).
self.addEventListener('fetch', function (ev) {
  var url = new URL(ev.request.url);
  if (ev.request.method !== 'GET' || url.origin !== self.location.origin) { return; }
  if (url.pathname.indexOf('/static/') !== 0) { return; }
  ev.respondWith(
    caches.match(ev.request).then(function (hit) {
      if (hit) { return hit; }
      return fetch(ev.request).then(function (resp) {
        if (resp && resp.ok) {
          var copy = resp.clone();
          caches.open(CACHE).then(function (c) { c.put(ev.request, copy); });
        }
        return resp;
      });
    })
  );
});
