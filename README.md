# desmosinator

Turn MIDI files into playable music in Desmos. Also can draw images pixel-by-pixel because why not.

## What does it do?

- **MIDI files** → Converts notes to Desmos `tone()` expressions with a time slider. Animate the slider and watch your music play!
- **Images** → Plots colored points for each pixel. It's slow but looks cool.
- **MuseScore links** → Paste a MuseScore URL and it'll download the MIDI for you (needs `dl-librescore` via npx).

## Install

```bash
go install
```

Or just:

```bash
go build
```

## Usage

### Play a MIDI file

```bash
desmosinator song.mid
```

**Tips:**

- Set the `t` slider to play indefinitely (click the play button on the slider)
- For best viewing: set Y range to `0` to `140`, X range to `t-100` to `t+100`

### From a MuseScore link

```bash
desmosinator https://musescore.com/some/song
```

### Draw an image

```bash
desmosinator image.png
```

## Flags

| Flag | What it does | Default |
|------|--------------|---------|
| `-cs<N>` | Chunk size (notes per visual group) | 500 |
| `-step<N>` | Pixel step for images (lower = more detail, slower) | 10 |
| `-nv` | No visualization for MIDI (just audio, no note dots) | off |

### Examples

```bash
# Smaller chunks
desmosinator -cs200 song.mid

# Higher res image (will take forever)
desmosinator -step5 image.png

# MIDI without the visual note plot
desmosinator -nv song.mid
```

## How it works

Opens Desmos in a browser (via [rod](https://github.com/go-rod/rod)), then injects expressions using `Calc.setExpression()`. For MIDI, it creates a `t` slider and a bunch of `tone(freq, volume{start<t<end})` expressions. Animate the slider and you get music.

## Requirements

- Go 1.21+
- Chrome/Chromium (rod will find it)
- `npx` if you want to use MuseScore links

## Notes

- Drums (MIDI channel 10) are filtered out because Desmos can't really do percussion
- Large MIDI files = lots of expressions = Desmos might chug a bit
- Images are plotted point by point so... be patient

---

*yes i named it desmosinator*
