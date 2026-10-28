const $ = (selector) => document.querySelector(selector);
const call = (method, ...args) => window.mygo.call(`Desktop.${method}`, ...args);
const ui = window.InlineExplainI18n;
const t = (key) => ui.t(key);
const icons = {
  copy: '<rect x="8" y="8" width="12" height="13" rx="2"/><path d="M16 8V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v10a2 2 0 0 0 2 2h3"/>',
  speaker: '<path d="m11 4-6 5H2v6h3l6 5V4ZM15 8a6 6 0 0 1 0 8M18 5a10 10 0 0 1 0 14"/>',
};
const icon = (name) => `<svg viewBox="0 0 24 24" aria-hidden="true">${icons[name]}</svg>`;
document.querySelectorAll('[data-icon]').forEach((button) => { button.innerHTML = icon(button.dataset.icon); });
ui.setLanguage('zh');

let state = null;
let resizePending = false;
function resize() {
  if (resizePending || $('#card').hidden) return;
  resizePending = true;
  requestAnimationFrame(() => {
    resizePending = false;
    const height = Math.ceil($('#content-body').scrollHeight);
    call('Resize', height).catch(() => {});
  });
}
function failure(error) {
  $('#error-section').hidden = false;
  $('#error').textContent = ui.error(error);
  resize();
}
function element(tag, className, text) {
  const node = document.createElement(tag);
  if (className) node.className = className;
  if (text !== undefined) node.textContent = text;
  return node;
}
function render(value) {
  if (state && value.id < state.id) return;
  state = value;
  const answer = value.answer || {};
  const chinese = (value.language || 'zh') === 'zh';
  ui.setLanguage(chinese ? 'zh' : 'en');
  $('#term').textContent = answer.term || value.passage?.selection || 'InlineExplain';
  const pronunciations = [];
  if (answer.ukIpa) pronunciations.push(`UK ${answer.ukIpa}`);
  if (answer.usIpa && answer.usIpa !== answer.ukIpa) pronunciations.push(`US ${answer.usIpa}`);
  if (answer.partOfSpeech) pronunciations.push(answer.partOfSpeech);
  $('#pronunciation').textContent = pronunciations.join(' · ');
  $('#meaning').textContent = answer.meaning || '';
  $('#loading').hidden = !value.busy;
  $('#loading').textContent = t('loading');
  $('#context-section').hidden = !answer.context;
  $('#context-label').textContent = t('context');
  $('#context').textContent = answer.context || '';
  $('#source-label').textContent = t('source');
  $('#source').replaceChildren();
  const passage = value.passage || {};
  $('#source').append(document.createTextNode(`${passage.before || ''} `), element('mark', '', passage.selection || ''), document.createTextNode(` ${passage.after || ''}`));
  $('#source').lang = 'en';
  $('#examples-label').textContent = t('examples');
  $('#examples-section').hidden = !answer.examples?.length;
  $('#examples').replaceChildren(...(answer.examples || []).map((example) => {
    const row = element('div', 'example');
    const content = element('div');
    const text = element('p', '', example.text); text.lang = 'en'; content.append(text);
    if (example.translation) { const translation = element('p', 'translation', example.translation); translation.lang = 'zh-Hans'; content.append(translation); }
    const speak = element('button', 'icon-button');
    speak.innerHTML = icon('speaker'); speak.title = t('readExample'); speak.setAttribute('aria-label', t('readExample'));
    speak.onclick = () => call('Speak', example.text).catch(failure);
    row.append(content, speak);
    return row;
  }));
  const relations = [];
  for (const [key, label] of [['synonyms', t('similar')], ['antonyms', t('opposite')]]) {
    if (!answer[key]) continue;
    const row = element('div'); row.append(element('span', 'relation-label', label), document.createTextNode(answer[key])); relations.push(row);
  }
  $('#relations').replaceChildren(...relations); $('#relations').hidden = relations.length === 0;
  $('#error-section').hidden = !value.error; $('#error').textContent = value.error ? ui.error(value.error) : '';
  $('#retry').hidden = !passage.selection; $('#retry').textContent = t('retry');
  $('#error-settings').textContent = t('settings');
  $('#copy').disabled = !answer.meaning;
  resize();
}

async function initCard() {
  $('#error-settings').onclick = () => call('OpenSettings');
  $('#copy').onclick = () => call('Copy').then(() => { $('#copy').title = t('copied'); }).catch(failure);
  $('#retry').onclick = () => call('Retry').catch(failure);
  $('#source-details').addEventListener('toggle', resize);
  document.addEventListener('keydown', (event) => { if (event.key === 'Escape') call('Hide'); });
  new ResizeObserver(resize).observe($('#content-body'));
  window.mygo.on('lookup:changed', render);
  render(await call('Read'));
}

async function initSettings() {
  $('#card').hidden = true; $('#settings-page').hidden = false;
  let saved = null;
  const draft = {};
  let editingLanguage = 'en';
  let refreshID = 0;
  let pendingSave = null;
  let latestRevision = -1;
  let latestPreferences = null;
  const promptKey = () => editingLanguage === 'zh' ? 'chinesePrompt' : 'englishPrompt';
  const status = (message, error = false) => { $('#save-status').textContent = error ? ui.error(message) : message; $('#save-status').classList.toggle('error', error); };
  const fields = {endpoint:'#endpoint-input', model:'#model-input', temperature:'#temperature-input', language:'#language-input', shortcut:'#shortcut-input'};
  function edit(key, value) {
    if (!pendingSave && saved && value === saved[key]) delete draft[key];
    else draft[key] = value;
  }
  function display(preferences) {
    const revision = preferences.revision ?? 0;
    if (revision < latestRevision) return;
    latestRevision = revision;
    latestPreferences = preferences;
    $('#app-version').textContent = preferences.version || 'dev';
    saved = preferences.settings;
    // A refresh can still contain the pre-save snapshot. Keep in-flight
    // edits until Save acknowledges them, including Undo to the old value.
    if (!pendingSave) for (const key of Object.keys(draft)) if (draft[key] === saved[key]) delete draft[key];
    const current = {...saved, ...draft};
    ui.setLanguage(current.language);
    for (const [key, selector] of Object.entries(fields)) $(selector).value = current[key];
    editingLanguage = current.language;
    $('#prompt-input').value = current[promptKey()];
    const hasKey = preferences.hasKey && current.endpoint === saved.endpoint;
    $('#key-input').placeholder = t(hasKey ? 'keySaved' : 'keyAdd');
    $('#key-note').textContent = t(hasKey ? 'keySavedNote' : 'keyMissingNote');
    $('#permission-note').textContent = t(preferences.screenRecording ? 'screenAllowed' : 'screenMissing');
    $('#permission-button').textContent = t(preferences.screenRecording ? 'screenOpen' : 'screenAllow');
  }
  async function refresh() {
    const id = ++refreshID;
    const preferences = await call('Preferences');
    if (id === refreshID) display(preferences);
  }
  for (const [key, selector] of Object.entries(fields)) {
    $(selector).oninput = () => edit(key, key === 'temperature' ? Number($(selector).value) : $(selector).value.trim());
  }
  $('#prompt-input').oninput = () => edit(promptKey(), $('#prompt-input').value);
  $('#language-input').onchange = () => {
    edit(promptKey(), $('#prompt-input').value);
    edit('language', $('#language-input').value);
    editingLanguage = $('#language-input').value;
    $('#prompt-input').value = ({...saved, ...draft})[promptKey()];
    if (latestPreferences) display(latestPreferences);
  };
  $('#permission-button').onclick = async () => {
    try { await call('RequestScreenRecording'); await call('OpenScreenSettings'); $('#permission-note').textContent = t('screenReopen'); }
    catch (error) { status(error.message || String(error), true); }
  };
  $('#reset-prompt').onclick = async () => {
    const language = editingLanguage;
    const value = await call('DefaultPrompt', language);
    edit(language === 'zh' ? 'chinesePrompt' : 'englishPrompt', value);
    if (editingLanguage === language) $('#prompt-input').value = value;
    status(t('resetDone'));
  };
  $('#settings-form').onsubmit = async (event) => {
    event.preventDefault();
    edit(promptKey(), $('#prompt-input').value);
    for (const [key, selector] of Object.entries(fields)) edit(key, key === 'temperature' ? Number($(selector).value) : $(selector).value.trim());
    const submitted = {...draft};
    const apiKey = $('#key-input').value;
    const button = event.submitter; button.disabled = true; status(t('saving'));
    pendingSave = submitted;
    ++refreshID;
    try {
      const preferences = await call('Save', submitted, apiKey);
      ++refreshID;
      for (const key of Object.keys(submitted)) if (draft[key] === submitted[key]) delete draft[key];
      if ($('#key-input').value === apiKey) $('#key-input').value = '';
      display(preferences);
      status(t(preferences.hasKey ? 'saved' : 'savedNoKey'));
    } catch (error) { status(error.message || String(error), true); }
    finally {
      ++refreshID;
      pendingSave = null;
      if (latestPreferences) display(latestPreferences);
      button.disabled = false;
    }
  };
  window.mygo.on('preferences:refresh', () => refresh().catch((error) => status(error.message || String(error), true)));
  await refresh();
}

if (new URLSearchParams(location.search).get('view') === 'settings') initSettings().catch((error) => { $('#save-status').textContent = ui.error(error); });
else initCard().catch(failure);
