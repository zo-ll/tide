# Issues

## Image paste support

Ctrl+V should capture an image from the system clipboard and attach it to
the current prompt as an image part, not text.

Detection order:
- `xclip -selection clipboard -t image/png -o`
- `xsel --clipboard --input` (image)
- `wl-paste --type image/png`
- `pbpaste` (macOS, if ever relevant)

If no image is found, fall back to current text-paste behavior.

## Large paste collapse

When a large multi-line text is pasted into the input, the editor should
show `[ +N lines pasted ]` instead of rendering all pasted content inline.

Threshold: if pasted text contains more than 3 lines, collapse it.

The full pasted text should still be stored in the input buffer and
submitted on Enter; only the display is collapsed.

## Picker UI: simpler scrollable inline style

The current fullscreen box picker should be replaced with the simpler
inline scrollable style used in `oi`'s line editor (`Editor.Select`):
- Items render below the prompt line, not in a centered box.
- Shows 5 visible items at a time with `>` marker on the selected row.
- Up/down navigates, type to filter, Enter selects, Ctrl+C cancels.
- No border box, no title bar chrome.
