{{< base}}
<div class="card">
  <div class="cardhead">
    <h1>{{ entity_label }}</h1>
    {{#if can_write}}<a class="btn" href="{{ base }}/new" data-new="1">New</a>{{/if}}
  </div>
  {{#if err}}<p class="flash bad" data-err="1">{{ err }}</p>{{/if}}
  {{#if readonly}}<p class="muted">Read-only table</p>{{/if}}
  <form method="get" action="{{ base }}" class="searchbar" data-search="1">
    <input type="text" name="q" value="{{ q }}" placeholder="Search">
    <button type="submit">Search</button>
    {{#if searching}}<a class="btn ghost" href="{{ base }}">Clear</a>{{/if}}
  </form>
  <table>
    <thead>{{{ table_head }}}</thead>
    <tbody>{{{ rows }}}</tbody>
  </table>
  {{{ pager }}}
</div>
