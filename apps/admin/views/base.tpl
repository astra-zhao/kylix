<!doctype html>
<html lang="en" data-theme="{{ theme }}">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>{{ page_title }} · KylixAdmin</title>
<link rel="stylesheet" href="/static/admin.css">
</head>
<body>
<div class="layout">
  <aside class="sidebar">
    <div class="brand">Kylix<span>Admin</span></div>
    <nav data-nav="1">
{{{ nav }}}
    </nav>
    <div class="spacer"></div>
    {{#if user}}
    <form method="post" action="/logout" class="logoutform">
      <input type="hidden" name="_csrf" value="{{ csrf }}">
      <div class="who">{{ user }}</div>
      <button type="submit" class="btn ghost"><span>Sign out</span></button>
    </form>
    {{/if}}
  </aside>
  <div class="content">
    <header class="topbar">
      <button type="button" class="iconbtn" data-sidebar-toggle="1" aria-label="Toggle navigation">☰</button>
      <div class="crumbs"><a href="/dashboard">Home</a> / {{ page_title }}</div>
      <div class="grow"></div>
      {{#if user}}
      <a class="iconbtn" href="/profile" data-profile-link="1">👤 Profile</a>
      {{/if}}
      {{#if theme_dark}}
      <a class="iconbtn" href="/theme?t=light&next={{ theme_next }}" data-theme-toggle="light" title="Switch to light">☀️</a>
      {{else}}
      <a class="iconbtn" href="/theme?t=dark&next={{ theme_next }}" data-theme-toggle="dark" title="Switch to dark">🌙</a>
      {{/if}}
    </header>
    <main class="main">
{{{ content }}}
    </main>
  </div>
</div>
<div id="toasts" data-toasts="1"></div>
<script src="/static/admin.js"></script>
</body>
</html>
