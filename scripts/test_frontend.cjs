const fs = require('node:fs');
const vm = require('node:vm');
const path = require('node:path');
const assert = require('node:assert/strict');
const sourcePath = process.argv[2] || path.join(__dirname, '../frontend/app.js');
const source = fs.readFileSync(sourcePath, 'utf8').split('\nif (new URLSearchParams')[0];

function fixture(hasKey = false, language = 'en') {
  const elements = new Map();
  const get = (selector) => {
    if (!elements.has(selector)) elements.set(selector, {value:'', placeholder:'', hidden:false, textContent:'', classList:{toggle(){}}, addEventListener(){}});
    return elements.get(selector);
  };
  const events = new Map();
  const backend = {settings:{endpoint:'https://api.deepseek.com',model:'deepseek-chat',temperature:.3,language:'en',shortcut:'Alt+S',englishPrompt:'English prompt',chinesePrompt:'Chinese prompt'},hasKey,revision:0};
  backend.settings.language = language;
  const calls = [];
  const gate = {holdSave:false, holdPreferences:false, save:null, commitSave:null, replySave:null, failSave:null, preferences:null};
  const preferences = () => structuredClone({...backend,screenRecording:true});
  const context = vm.createContext({document:{querySelector:get,querySelectorAll:()=>[]},window:{mygo:{
    on(name, handler) { events.set(name, handler); },
    call(method, ...args) {
      calls.push([method,...args]);
      if (method === 'Desktop.Preferences') {
        const snapshot = preferences();
        if (gate.holdPreferences) return new Promise(resolve => { gate.preferences = () => resolve(snapshot); });
        return Promise.resolve(snapshot);
      }
      if (method === 'Desktop.Save') {
        const commit = () => { Object.assign(backend.settings,args[0]); if (args[1]) backend.hasKey = true; backend.revision++; return preferences(); };
        if (gate.holdSave) return new Promise((resolve,reject) => {
          let committed;
          gate.commitSave = () => { committed = commit(); };
          gate.replySave = () => resolve(committed);
          gate.save = () => { gate.commitSave(); gate.replySave(); };
          gate.failSave = () => reject(new Error('Could not save settings.'));
        });
        return Promise.resolve(commit());
      }
      if (method === 'Desktop.DefaultPrompt') return Promise.resolve(args[0] === 'zh' ? 'Chinese default' : 'English default');
      return Promise.resolve();
    }
  }},console,Promise});
  vm.runInContext(fs.readFileSync(path.join(__dirname,'../frontend/i18n.js'),'utf8'),context);
  vm.runInContext(source,context,{filename:sourcePath});
  return {get,backend,events,calls,context,gate};
}
async function start(value) { await vm.runInContext('initSettings()',value.context); }
async function submit(value) { await value.get('#settings-form').onsubmit({preventDefault(){},submitter:{disabled:false}}); }
const tick = () => new Promise(setImmediate);

(async () => {
  const failures = [];
  const check = async (name, run) => {
    try { await run(); console.log(`PASS: ${name}`); }
    catch (error) { failures.push(name); console.error(`FAIL: ${name}: ${error.message}`); }
  };
  await check('Settings refresh preserves a card language change', async () => {
    const settings = fixture(); await start(settings);
    settings.backend.settings.language = 'zh'; settings.backend.revision++;
    if (settings.events.has('preferences:refresh')) { settings.events.get('preferences:refresh')(); await tick(); }
    assert.equal(settings.get('#language-input').value,'zh','Reopened Settings must show the current language');
    settings.get('#temperature-input').value = '0.4'; settings.get('#temperature-input').oninput?.();
    await submit(settings);
    assert.equal(settings.backend.settings.language,'zh','Saving a different field must preserve the card language');
  });
  await check('Missing API key is not reported as saved', async () => {
    const noKey = fixture(); await start(noKey); await submit(noKey);
    assert.equal(noKey.backend.hasKey,false);
    assert.notEqual(noKey.get('#key-input').placeholder,'Saved in Keychain','The form must not claim a missing key is saved');
  });
  await check('An existing saved API key remains visible', async () => {
    const withKey = fixture(true); await start(withKey); await submit(withKey);
    assert.equal(withKey.get('#key-input').placeholder,'Saved in Keychain');
  });
  await check('Refreshing keeps unsaved prompts and submits only edited fields', async () => {
    const draft = fixture(); await start(draft);
    draft.get('#prompt-input').value = 'My unsaved English instructions'; draft.get('#prompt-input').oninput?.();
    draft.backend.settings.language = 'zh'; draft.backend.revision++;
    if (draft.events.has('preferences:refresh')) { draft.events.get('preferences:refresh')(); await tick(); }
    assert.equal(draft.get('#language-input').value,'zh');
    await submit(draft);
    assert.equal(draft.backend.settings.englishPrompt,'My unsaved English instructions');
    const sent = draft.calls.filter(([method])=>method==='Desktop.Save').at(-1)[1];
    assert.ok(!Object.hasOwn(sent,'language'),'Unedited fields must not be submitted');
  });
  await check('Undo during a pending Save keeps the newer edit', async () => {
    const value = fixture(); await start(value);
    const original = value.get('#prompt-input').value;
    value.get('#prompt-input').value = 'Submitted prompt'; value.get('#prompt-input').oninput();
    value.gate.holdSave = true;
    const saving = submit(value);
    assert.equal(typeof value.gate.save,'function');
    value.get('#prompt-input').value = original; value.get('#prompt-input').oninput();
    value.gate.save(); await saving;
    assert.equal(value.get('#prompt-input').value,original,'A pending Save discarded a newer Undo edit');
  });
  await check('A late pre-save refresh cannot overwrite the committed form', async () => {
    const value = fixture(); await start(value);
    value.get('#model-input').value = 'new-model'; value.get('#model-input').oninput();
    value.gate.holdSave = true;
    const saving = submit(value);
    value.gate.holdPreferences = true;
    value.events.get('preferences:refresh')();
    assert.equal(typeof value.gate.preferences,'function');
    value.gate.save(); await saving;
    value.gate.preferences(); await tick();
    assert.equal(value.get('#model-input').value,'new-model','A stale refresh replaced a successful Save');
  });
  await check('A late Save reply cannot overwrite a newer card language', async () => {
    const value = fixture(); await start(value);
    value.get('#prompt-input').value = 'New prompt'; value.get('#prompt-input').oninput();
    value.gate.holdSave = true;
    const saving = submit(value);
    value.gate.commitSave();
    value.backend.settings.language = 'zh'; value.backend.revision++;
    value.events.get('preferences:refresh')(); await tick();
    value.gate.replySave(); await saving;
    assert.equal(value.get('#language-input').value,'zh','A late Save reply replaced a newer language');
  });
  await check('Undo stays dirty when Settings refreshes during Save', async () => {
    const value = fixture(); await start(value);
    const original = value.get('#prompt-input').value;
    value.get('#prompt-input').value = 'Submitted prompt'; value.get('#prompt-input').oninput();
    value.gate.holdSave = true;
    const saving = submit(value);
    value.get('#prompt-input').value = original; value.get('#prompt-input').oninput();
    value.events.get('preferences:refresh')(); await tick();
    value.gate.save(); await saving;
    assert.equal(value.get('#prompt-input').value,original);
  });
  await check('Refreshing does not roll back the value currently being saved', async () => {
    const value = fixture(); await start(value);
    value.get('#prompt-input').value = 'Submitted prompt'; value.get('#prompt-input').oninput();
    value.gate.holdSave = true;
    const saving = submit(value);
    value.events.get('preferences:refresh')(); await tick();
    assert.equal(value.get('#prompt-input').value,'Submitted prompt');
    value.gate.save(); await saving;
  });
  await check('A rejected Save keeps the latest unsaved Undo/Redo value', async () => {
    const value = fixture(); await start(value);
    const original = value.get('#prompt-input').value;
    value.get('#prompt-input').value = 'Submitted prompt'; value.get('#prompt-input').oninput();
    value.gate.holdSave = true;
    const saving = submit(value);
    value.get('#prompt-input').value = original; value.get('#prompt-input').oninput();
    value.get('#prompt-input').value = 'Submitted prompt'; value.get('#prompt-input').oninput();
    value.gate.failSave(); await saving;
    assert.equal(value.get('#prompt-input').value,'Submitted prompt','A failed Save discarded an unsaved edit');
    assert.match(value.get('#save-status').textContent,/Could not save settings/);
  });
  await check('Chinese interface selects Chinese prompt and switches together to English', async () => {
    const value = fixture(false,'zh'); await start(value);
    assert.equal(vm.runInContext('window.InlineExplainI18n.t("save")',value.context),'保存设置');
    assert.equal(value.get('#prompt-input').value,'Chinese prompt');
    assert.equal(value.get('#key-input').placeholder,'填写 API Key');
    value.get('#language-input').value = 'en'; value.get('#language-input').onchange();
    assert.equal(vm.runInContext('window.InlineExplainI18n.t("save")',value.context),'Save settings');
    assert.equal(value.get('#prompt-input').value,'English prompt');
    await submit(value);
    assert.equal(value.backend.settings.language,'en');
  });
  if (failures.length) process.exitCode = 1;
})().catch(error=>{ console.error(error.message); process.exitCode=1; });
