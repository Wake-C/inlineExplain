// Sample data for the README prototype. The production frontend stays unchanged.
(() => {
  const preview = parent.InlineExplainPreview;
  const sample = {
    id: 1,
    language: 'zh',
    busy: false,
    error: '',
    passage: {
      selection: 'grasp',
      before: 'The guide helps readers',
      after: 'unfamiliar ideas without breaking their flow.'
    },
    answer: {
      term: 'grasp',
      partOfSpeech: '动词',
      ukIpa: '/ɡrɑːsp/',
      usIpa: '/ɡræsp/',
      meaning: '理解；领会',
      context: '把原本陌生的想法理解透彻，读懂它们的要点。',
      examples: [
        {text: 'She quickly grasped the main idea.', translation: '她很快领会了主要意思。'},
        {text: 'A simple example can help you grasp the concept.', translation: '一个简单的例子能帮助你理解这个概念。'}
      ],
      synonyms: 'understand, comprehend',
      antonyms: 'misunderstand'
    }
  };
  window.mygo = {
    on() {},
    async call(method, ...args) {
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
        case 'Desktop.Speak': {
          if (!('speechSynthesis' in window)) return;
          speechSynthesis.cancel();
          const utterance = new SpeechSynthesisUtterance(args[0]);
          utterance.lang = 'en-US';
          speechSynthesis.speak(utterance);
          return;
        }
        default: throw new Error('This control is unavailable in the preview.');
      }
    }
  };
  document.addEventListener('keydown', preview.keydown);
})();
