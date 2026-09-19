{{< base}}
<div class="loginbox">
  <div class="loginhead">KylixAdmin</div>
  {{#if error}}<div class="alert">{{ error }}</div>{{/if}}
  <form method="post" action="/login">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    <label>Username
      <input name="username" autofocus autocomplete="username">
    </label>
    <label>Password
      <input name="password" type="password" autocomplete="current-password">
    </label>
    <label class="check">
      <input type="checkbox" name="remember" value="1"> Remember me for 30 days
    </label>
    <button type="submit">Sign in</button>
  </form>
</div>
