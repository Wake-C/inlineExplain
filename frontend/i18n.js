(() => {
  const en = {
    card:'Word explanation',controls:'Explanation controls',switchLanguage:'Switch language',copy:'Copy explanation',pin:'Keep this window open',settings:'Settings',close:'Close',closeTip:'Close · Escape',pronounce:'Pronounce word',readExample:'Read example aloud',
    initialHint:'Point at a word and press Option+S.',context:'IN THIS SENTENCE',source:'Original sentence',examples:'EXAMPLES',similar:'Similar',opposite:'Opposite',retry:'Try again',ready:'Ready when you are',explaining:'Explaining',attention:'Needs attention',english:'Simple English',chinese:'Chinese explanation',loading:'Reading the word in context…',copied:'Copied',
    settingsTitle:'A little help. Keep reading.',settingsIntro:'Point at a word, press ⌥ S, and understand it where you found it.',service:'Explanation service',serviceNote:'Use your own Chat Completions-compatible provider. Your API key stays in macOS Keychain.',provider:'Provider URL',model:'Model',temperature:'Temperature',key:'API key',keyAdd:'Add your API key',keySaved:'Saved in Keychain',keySavedNote:'A key is saved for this server. Leave this field empty to keep it.',keyMissingNote:'No key is saved for this server. Keys stay in macOS Keychain.',
    preferences:'Language and explanation',language:'Language',shortcut:'Lookup shortcut',prompt:'Explanation instructions',reset:'Reset default',promptNote:'Each language has its own editable prompt. Keep the JSON fields so meaning, context, examples and word relations can appear in the card.',
    screenTitle:'Read text under your pointer',screenChecking:'Checking Screen Recording access…',screenAllowed:'Screen Recording is allowed. You can look up words on screen.',screenMissing:'Allow Screen Recording so InlineExplain can read the word and nearby text.',screenOpen:'Open macOS permissions',screenAllow:'Allow Screen Recording',screenReopen:'After allowing InlineExplain, quit and reopen InlineExplain from Applications.',privacy:'Screen capture and OCR run locally when you press the shortcut. Only the word and nearby text go to your provider. Screenshots and reading history are not saved.',
    save:'Save settings',saving:'Saving…',saved:'Saved. Point at a word and press your shortcut.',savedNoKey:'Settings saved. Add an API key before looking up words.',resetDone:'Default prompt restored. Save to apply it.',genericError:'The action could not be completed. Check Settings and try again.'
  };
  const zh = {
    card:'单词解释',controls:'解释工具',switchLanguage:'切换语言',copy:'复制解释',pin:'保持窗口打开',settings:'设置',close:'关闭',closeTip:'关闭 · Escape',pronounce:'朗读单词',readExample:'朗读例句',
    initialHint:'将鼠标放在单词上，按 Option+S。',context:'语境说明',source:'查看原句',examples:'例句',similar:'近义词',opposite:'反义词',retry:'重试',ready:'随时为你解惑',explaining:'正在解释',attention:'需要处理',english:'英文解释',chinese:'中文解释',loading:'正在理解当前语境…',copied:'已复制',
    settingsTitle:'读懂一句，继续阅读。',settingsIntro:'将鼠标放在单词上，按 ⌥ S，就地理解它的意思。',service:'解释服务',serviceNote:'填写兼容 Chat Completions 的服务地址。API Key 保存在 macOS 钥匙串中。',provider:'服务地址',model:'模型',temperature:'温度',key:'API Key',keyAdd:'填写 API Key',keySaved:'已保存到钥匙串',keySavedNote:'此服务已保存 Key。留空即可继续使用。',keyMissingNote:'此服务尚未保存 Key。Key 仅保存在 macOS 钥匙串中。',
    preferences:'语言与解释',language:'语言',shortcut:'取词快捷键',prompt:'解释指令（Prompt）',reset:'恢复默认',promptNote:'中英文分别保存自己的指令，切换语言时会使用对应版本。保留 JSON 字段，释义、语境、例句和近反义词才能正常显示。',
    screenTitle:'读取鼠标下的文字',screenChecking:'正在检查屏幕录制权限…',screenAllowed:'已获得屏幕录制权限，可以开始取词。',screenMissing:'允许屏幕录制后，InlineExplain 才能读取单词和附近文字。',screenOpen:'打开 macOS 权限设置',screenAllow:'允许屏幕录制',screenReopen:'允许 InlineExplain 后，请退出并从“应用程序”重新打开。',privacy:'仅在按下快捷键时截图，并在本机识别。只有单词和附近文字会发送给你选择的服务，不会保存截图或阅读历史。',
    save:'保存设置',saving:'正在保存…',saved:'已保存。将鼠标放在单词上，按取词快捷键即可。',savedNoKey:'设置已保存，请先填写 API Key 再取词。',resetDone:'已恢复默认指令，保存后生效。',genericError:'操作未完成，请检查设置后重试。'
  };
  const errors = {
    'Enter an HTTPS provider URL, or an HTTP localhost URL.':'请填写 HTTPS 服务地址；本机 localhost 服务也可以使用 HTTP。',
    'Enter a valid provider hostname.':'请填写有效的服务域名。',
    'Add your model and API key in Settings.':'请在设置中填写模型和 API Key。',
    'Add your provider API key in Settings.':'请在设置中填写此服务的 API Key。',
    'Temperature must be between 0 and 2.':'温度必须在 0 到 2 之间。',
    'Could not create the provider request.':'无法创建请求，请检查服务设置。',
    'Could not reach the provider. Check your connection and provider URL.':'无法连接解释服务，请检查网络和服务地址。',
    'This provider did not return a streaming response. Check model compatibility.':'服务未返回流式响应，请检查模型是否兼容。',
    'The provider returned an invalid stream. Try again.':'服务返回的流式数据无效，请重试。',
    'The model declined this request.':'模型未接受这次请求。',
    'The provider stopped before completing the explanation. Try again.':'解释尚未完成，服务就停止了响应，请重试。',
    'The explanation is too long. Shorten your prompt.':'解释内容过长，请缩短指令。',
    'The connection was interrupted. Try again.':'连接已中断，请重试。',
    'The connection ended before the explanation was complete. Try again.':'解释尚未完成，连接就结束了，请重试。',
    'The model returned an incomplete explanation. Try again.':'模型返回的解释不完整，请重试。',
    'The model returned extra text. Check your prompt and try again.':'模型返回了额外文字，请检查指令后重试。',
    'The explanation is missing required fields. Check your prompt.':'解释缺少必填字段，请检查指令。',
    'The model returned an empty explanation field. Try again.':'模型返回了空的解释字段，请重试。',
    'The explanation is missing its meaning, context or examples. Try again.':'解释缺少释义、语境或例句，请重试。',
    'The example language does not match your settings. Check your prompt.':'例句语言与设置不一致，请检查指令。',
    'Word relations must be short, single-line text. Check your prompt.':'近反义词应为简短的单行文字，请检查指令。',
    'Choose English or Chinese.':'请选择中文或英文。',
    'Enter a model name.':'请填写模型名称。',
    'Choose a supported shortcut.':'请选择支持的快捷键。',
    'That shortcut is not available. Choose another.':'此快捷键不可用，请选择另一个。',
    'Could not read your settings.':'无法读取设置。',
    'Your settings file could not be read. Open Settings to save it again.':'设置文件无法读取，请打开设置重新保存。',
    'Could not create the settings folder.':'无法创建设置文件夹。',
    'Could not encode settings.':'无法处理设置内容。',
    'Could not save settings.':'无法保存设置。',
    'Could not save your API key to Keychain.':'无法将 API Key 保存到钥匙串。',
    'Speech could not start. Check your macOS voices.':'无法开始朗读，请检查 macOS 的语音设置。',
    'Allow Screen Recording in Settings, then quit and reopen InlineExplain.':'请在设置中允许屏幕录制，然后退出并重新打开 InlineExplain。',
    'The pointer is not on an available display.':'鼠标不在可用的显示器上。',
    'Could not capture the screen. Check Screen Recording access in Settings.':'无法截取屏幕，请在设置中检查屏幕录制权限。',
    'No word found. Point directly at text and try again.':'未找到单词，请将鼠标放在文字上后重试。',
    'No word found. Point directly at readable text and press Option+S again.':'未找到单词，请将鼠标放在清晰的文字上，再按取词快捷键。',
    'Point directly at a word, then press Option+S again.':'请将鼠标放在单词上，再按取词快捷键。',
    'Screen capture failed. Check Screen Recording access and try again.':'截图失败，请检查屏幕录制权限后重试。'
  };
  let language = 'zh';
  const t = (key) => (language === 'zh' ? zh[key] : en[key]) ?? en[key] ?? key;
  const setLanguage = (value) => {
    language = value === 'en' ? 'en' : 'zh';
    if (document.documentElement) document.documentElement.lang = language === 'zh' ? 'zh-Hans' : 'en';
    document.querySelectorAll('[data-i18n]').forEach(node => { node.textContent = t(node.dataset.i18n); });
    document.querySelectorAll('[data-i18n-aria]').forEach(node => { node.setAttribute('aria-label',t(node.dataset.i18nAria)); if(node.tagName==='BUTTON') node.title = t(node.dataset.i18nAria); });
    document.querySelectorAll('[data-i18n-placeholder]').forEach(node => { node.placeholder=t(node.dataset.i18nPlaceholder); });
  };
  const error = (value) => {
    const text = value?.message || String(value);
    if (language === 'en') return text;
    if (errors[text]) return errors[text];
    const status = text.match(/^Provider returned HTTP (\d+)\./);
    return status ? `服务返回 HTTP ${status[1]}，请检查 Key、模型及服务设置。` : zh.genericError;
  };
  window.InlineExplainI18n = {t,setLanguage,error};
})();
