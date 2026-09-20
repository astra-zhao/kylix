{{< base}}
<div class="pagehead">
  <h1>Dashboard</h1>
  <p class="muted">Signed in as <strong>{{ user }}</strong> · <a href="/profile">Profile</a></p>
</div>
<div class="stats" data-stats="1">{{{ cards }}}</div>
<div class="card">
  <div class="cardhead"><h2>Logins, last 7 days</h2></div>
  {{{ chart }}}
</div>
<div class="card">
  <h2>Your permissions</h2>
  <p class="muted"><code>{{ perms }}</code></p>
</div>
