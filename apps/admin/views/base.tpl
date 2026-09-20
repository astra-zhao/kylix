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
{{{ nav }}}
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
<script src="/static/admin.js"></script>
</body>
</html>
