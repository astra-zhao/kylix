{{< base}}
<div class="card">
  <h1>{{#if mode_edit}}Edit user{{/if}}{{#if mode_create}}New user{{/if}}</h1>
  {{#if err}}<p class="flash bad">{{ err }}</p>{{/if}}
  <form method="post" action="{{ form_action }}" class="stack">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    <input type="hidden" name="id" value="{{ f_id }}">
    <label>Username <input type="text" name="username" value="{{ f_username }}"></label>
    <label>Display name <input type="text" name="display_name" value="{{ f_display }}"></label>
    {{#if mode_edit}}<label>Password <input type="password" name="password" placeholder="leave blank to keep current"></label>{{/if}}
    {{#if mode_create}}<label>Password <input type="password" name="password"></label>{{/if}}
    <label class="check"><input type="checkbox" name="is_active" value="1" {{#if f_active}}checked{{/if}}> Active</label>
    <div><button type="submit">Save</button> <a class="btn ghost" href="/users">Cancel</a></div>
  </form>
</div>
