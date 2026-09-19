{{< base}}
<div class="card">
  <div class="cardhead">
    <h1>Roles</h1>
    {{#if can_write}}<a class="btn" href="/roles/new">New role</a>{{/if}}
  </div>
  {{#if err}}<p class="flash bad">{{ err }}</p>{{/if}}
  <table>
    <thead>
      <tr><th>ID</th><th>Name</th><th>Description</th><th>Permissions</th>{{#if can_write}}<th>Actions</th>{{/if}}</tr>
    </thead>
    <tbody>
{{{ role_rows }}}
    </tbody>
  </table>
{{{ pager }}}
</div>
