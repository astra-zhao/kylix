{{< base}}
<div class="card">
  <div class="cardhead">
    <h1>Users</h1>
    {{#if can_write}}<a class="btn" href="/users/new">New user</a>{{/if}}
  </div>
  {{#if err}}<p class="flash bad">{{ err }}</p>{{/if}}
  <table>
    <thead>
      <tr><th>ID</th><th>Username</th><th>Display name</th><th>Status</th>{{#if can_write}}<th>Actions</th>{{/if}}</tr>
    </thead>
    <tbody>
{{{ user_rows }}}
    </tbody>
  </table>
{{{ pager }}}
</div>
