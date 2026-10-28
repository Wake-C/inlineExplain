# Reading card

Use a compact white card beside the source text, with no dark mode. Establish
hierarchy through typography, spacing and alignment. Use quiet green accents
and a soft tint only behind the explanation of the word in context.

The word leads, with the copy icon at its right. Follow with pronunciation,
meaning and contextual explanation. Keep the original sentence available on
demand, then examples and useful word relations. Do not add an app toolbar or
status footer to the card. Give the word and meaning a reader-like serif face;
use system fonts for controls and supporting text.

Keep controls compact, named for accessibility and visibly focusable. Avoid
entrance and streaming animations. Escape and outside-click dismiss the card.
Settings and language choices belong in the menu-bar app; provider setup,
prompt editing and permissions remain on the separate Settings surface.

One language preference selects the interface, native menus and explanation
prompt. Simplified Chinese is the default; English remains available. Keep
custom prompts separate for each language. Default explanations state the
meaning in context directly, without an opening phrase such as “此处指” or
“It describes”. Prompt wording belongs in
[`internal/explain/prompts/`](internal/explain/prompts/); do not trim model output
to enforce this style.

[`frontend/style.css`](frontend/style.css) owns exact visual tokens and type
sizes. [`main.go`](main.go) owns native window dimensions and placement.
