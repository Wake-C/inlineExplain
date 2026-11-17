// Sample data for the README prototype. The production frontend stays unchanged.
(() => {
  const preview = parent.InlineExplainPreview;
  const sample = {
    id: 1,
    language: preview.language,
    busy: false,
    error: '',
    passage: {
      selection: 'bank',
      before: 'She sat on the',
      after: 'of the river and watched the water.'
    },
    answer: {
      term: 'bank',
      partOfSpeech: preview.language === 'en' ? 'noun' : '名词',
      ukIpa: '/bæŋk/',
      usIpa: '/bæŋk/',
      meaning: preview.language === 'en' ? 'The land along the side of a river.' : '河岸；河边的陆地。',
      context: preview.language === 'en' ? 'She is sitting on the land beside the river, watching the water.' : '她坐在河边的陆地上，看着河水流过。',
      examples: [
        {text: 'We walked along the river bank.', translation: preview.language === 'en' ? '' : '我们沿着河岸散步。'},
        {text: 'Trees grow on both banks.', translation: preview.language === 'en' ? '' : '两岸都长着树。'}
      ],
      synonyms: 'riverside, shore',
      antonyms: ''
    }
  };
  window.mygo = {
    on() {},
    async call(method, ...args) {
      const speak = (text) => {
        if (!('speechSynthesis' in window)) return;
        speechSynthesis.cancel();
        const utterance = new SpeechSynthesisUtterance(text);
        utterance.lang = 'en-US';
        speechSynthesis.speak(utterance);
      };
      switch (method) {
        case 'Desktop.Read': return sample;
        case 'Desktop.Resize':
          preview.resize(args[0]);
          preview.ready();
          return;
        case 'Desktop.Hide': preview.hide(); return;
        case 'Desktop.Copy':
          return navigator.clipboard.writeText([
            sample.answer.term, sample.answer.meaning, sample.answer.context,
            ...sample.answer.examples.flatMap((example) => [example.text, example.translation])
          ].join('\n\n'));
        case 'Desktop.Speak':
        case 'Desktop.SpeakWord': { speak(args[0]); return; }
        default: throw new Error('This control is unavailable in the preview.');
      }
    }
  };
  document.addEventListener('keydown', preview.keydown);
})();
