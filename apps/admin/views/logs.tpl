{{< base}}
<div class="card">
  <h1>Logins</h1>
  <table>
    <thead>
      <tr><th>Time</th><th>Username</th><th>Result</th><th>Reason</th><th>IP</th></tr>
    </thead>
    <tbody>
{{{ login_rows }}}
    </tbody>
  </table>
{{{ login_pager }}}
</div>
<div class="card">
  <h1>Operations</h1>
  <table>
    <thead>
      <tr><th>Time</th><th>Username</th><th>Method</th><th>Path</th><th>Status</th><th>IP</th></tr>
    </thead>
    <tbody>
{{{ op_rows }}}
    </tbody>
  </table>
{{{ op_pager }}}
</div>
