{{< base}}
<div class="card">
  <div class="cardhead">
    <h1>{{ form_title }}</h1>
  </div>
  {{#if err}}<p class="flash bad" data-err="1">{{ err }}</p>{{/if}}
  <form method="post" action="{{ form_action }}" class="stack" data-form="1">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    {{#if row_id}}<input type="hidden" name="id" value="{{ row_id }}">{{/if}}
    {{{ fields }}}
    <div class="formactions">
      <button type="submit">Save</button>
      <a class="btn ghost" href="{{ back_url }}">Cancel</a>
    </div>
  </form>
</div>
