{{< base}}
<div class="card">
  <h1>{{#if mode_edit}}Edit role{{/if}}{{#if mode_create}}New role{{/if}}</h1>
  {{#if err}}<p class="flash bad">{{ err }}</p>{{/if}}
  <form method="post" action="{{ form_action }}" class="stack">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    <input type="hidden" name="id" value="{{ f_id }}">
    <label>Name <input type="text" name="name" value="{{ f_name }}"></label>
    <label>Description <input type="text" name="description" value="{{ f_description }}"></label>
    <fieldset>
      <legend>Permissions</legend>
{{{ perm_rows }}}
    </fieldset>
    <div><button type="submit">Save</button> <a class="btn ghost" href="/roles">Cancel</a></div>
  </form>
</div>
