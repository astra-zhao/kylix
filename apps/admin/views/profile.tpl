{{< base}}
<div class="card" data-profile="1">
  <div class="cardhead"><h1>Profile</h1></div>
  {{#if err}}<p class="flash bad" data-err="1">{{ err }}</p>{{/if}}
  {{#if ok}}<p class="flash good" data-ok="1">{{ ok }}</p>{{/if}}

  <div class="profilerow">
    {{{ avatar_img }}}
    <div>
      <div class="who">{{ profile_user }}</div>
      <div class="muted">{{ profile_display }}</div>
    </div>
  </div>

  <form method="post" action="/profile/display" class="stack" data-form="display">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    <div class="field">
      <label for="f_display_name">Display name</label>
      <input id="f_display_name" type="text" name="display_name" value="{{ profile_display }}">
    </div>
    <button type="submit">Save name</button>
  </form>

  <form method="post" action="/profile/avatar" class="stack" data-form="avatar" enctype="application/x-www-form-urlencoded">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    <input type="hidden" name="avatar_data" id="avatar_data" value="">
    <div class="field">
      <label for="avatar_file">Avatar (PNG/JPEG, read in the browser and sent as base64)</label>
      <input id="avatar_file" type="file" accept="image/*" data-avatar-input="1">
    </div>
    <button type="submit">Save avatar</button>
    {{#if has_avatar}}<button type="submit" name="avatar_data" value="" class="btn ghost">Remove avatar</button>{{/if}}
  </form>

  <form method="post" action="/profile/password" class="stack" data-form="password">
    <input type="hidden" name="_csrf" value="{{ csrf }}">
    <div class="field">
      <label for="f_current_password">Current password</label>
      <input id="f_current_password" type="password" name="current_password" autocomplete="current-password">
    </div>
    <div class="field">
      <label for="f_new_password">New password <span class="req">*</span></label>
      <input id="f_new_password" type="password" name="new_password" autocomplete="new-password">
    </div>
    <div class="field">
      <label for="f_confirm_password">Confirm new password <span class="req">*</span></label>
      <input id="f_confirm_password" type="password" name="confirm_password" autocomplete="new-password">
    </div>
    <button type="submit">Change password</button>
  </form>
</div>
