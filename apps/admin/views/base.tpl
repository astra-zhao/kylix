<!doctype html>
<html>
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{ page_title }} · KylixAdmin</title>
<link rel="stylesheet" href="/static/admin.css">
</head>
<body>
<div class="layout">
  <aside class="sidebar">
    <div class="brand">KylixAdmin</div>
    <nav>
      {{#if nav_dashboard}}<a href="/dashboard">Dashboard</a>{{/if}}
      {{#if nav_users}}<a href="/users">Users</a>{{/if}}
      {{#if nav_roles}}<a href="/roles">Roles</a>{{/if}}
      {{#if nav_logs}}<a href="/logs">Logs</a>{{/if}}
    </nav>
    {{#if user}}
    <form method="post" action="/logout" class="logoutform">
      <input type="hidden" name="_csrf" value="{{ csrf }}">
      <div class="who">{{ user }}</div>
      <button type="submit">Sign out</button>
    </form>
    {{/if}}
  </aside>
  <main class="main">
{{{ content }}}
  </main>
</div>
</body>
</html>
