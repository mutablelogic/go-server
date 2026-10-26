package tui

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"os"
	"strings"

	// Packages
	draw "golang.org/x/image/draw"
)

///////////////////////////////////////////////////////////////////////////////
// TYPES

type imageFormatter struct {
	width, height int
	protocol      imageProtocol
}

// imageProtocol is how an image is written to a terminal
type imageProtocol int

// countingWriter counts the bytes written
type countingWriter struct {
	w io.Writer
	n int
}

const (
	// imageBlocks writes the image as characters, two pixels to a
	// character, with the upper half block, in 24-bit colour
	imageBlocks imageProtocol = iota

	// imageBlocks256 writes the image as characters in the 256 colours of
	// xterm
	imageBlocks256

	// imageKitty writes the image with kitty's graphics protocol, which
	// Ghostty also has
	imageKitty

	// imageITerm writes the image with iTerm2's inline images, which
	// WezTerm also has
	imageITerm
)

const (
	imageDefaultWidth = 80

	// cellWidth and cellHeight are the size of a character, in pixels, which
	// is assumed, as it isn't known, so a small image isn't made larger
	cellWidth, cellHeight = 8, 16

	// maxPixels is the largest width or height of an image which is sent
	// to the terminal, as pixels, which is otherwise twice the size of the
	// characters it's displayed in, for a screen of high resolution
	maxPixels = 2048

	// chunkSize is the size of a part of the image sent with kitty's
	// protocol, which it limits
	chunkSize = 4096
)

///////////////////////////////////////////////////////////////////////////////
// LIFECYCLE

// Image returns a formatter which writes images to a terminal, to
// fit its width and height, which are characters: as pixels, with the
// graphics protocols of the terminals which have one, such as kitty's and
// iTerm2's, or else as characters, two pixels to a character, in colour.
// The protocol is chosen from the terminal's environment variables.
func Image(options ...Opt) *imageFormatter {
	var opts opts
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	width := imageDefaultWidth
	if opts.width > 0 {
		width = opts.width
	}
	height := opts.height
	if height <= 0 {
		height = width
	}

	return &imageFormatter{width: width, height: height, protocol: detectImageProtocol(os.Getenv)}
}

///////////////////////////////////////////////////////////////////////////////
// RENDER

// Write writes the image to fit in the width and height, without making it
// larger than it is, followed by a line break. With a graphics protocol, an
// image isn't larger than its pixels, in characters of the assumed size. As
// characters, a pixel is at most half a character, which is the smallest
// which can be drawn, so a small image is larger than it is with a graphics
// protocol, rather than losing its detail.
func (f *imageFormatter) Write(w io.Writer, img image.Image) (int, error) {
	cw := &countingWriter{w: w}
	err := f.write(cw, img)
	return cw.n, err
}

func (f *imageFormatter) write(w io.Writer, img image.Image) error {
	bounds := img.Bounds()
	if bounds.Empty() {
		return nil
	}
	switch f.protocol {
	case imageKitty, imageITerm:
		c, r := fit(bounds.Dx(), bounds.Dy(), f.width, f.height, cellWidth, cellHeight)
		data, err := encode(img, min(maxPixels, 2*c*cellWidth), min(maxPixels, 2*r*cellHeight))
		if err != nil {
			return err
		}
		if f.protocol == imageKitty {
			return writeKitty(w, data, c, r)
		}
		return writeITerm(w, data, c, r)
	default:
		// Two pixels to a character, which is twice as high as it's wide,
		// so each pixel is square, and the image isn't sampled with more
		// pixels than it has
		c, r := fit(bounds.Dx(), bounds.Dy(), f.width, f.height*2, 1, 1)
		return writeBlocks(w, scale(img, c, r), f.protocol == imageBlocks256)
	}
}

func (w *countingWriter) Write(p []byte) (int, error) {
	n, err := w.w.Write(p)
	w.n += n
	return n, err
}

///////////////////////////////////////////////////////////////////////////////
// PRIVATE METHODS

// detectImageProtocol returns the protocol of the terminal, from its
// environment variables: kitty's or iTerm2's, which aren't used in tmux or
// screen, as they don't pass them through, including in screen with another
// terminal type, which STY shows, or else characters in 24-bit colour, when
// the terminal has it, or in 256 colours
func detectImageProtocol(getenv func(string) string) imageProtocol {
	multiplexed := getenv("TMUX") != "" || getenv("STY") != "" || strings.HasPrefix(getenv("TERM"), "screen") || strings.HasPrefix(getenv("TERM"), "tmux")
	switch {
	case multiplexed:
	case getenv("KITTY_WINDOW_ID") != "", getenv("TERM") == "xterm-kitty", getenv("TERM_PROGRAM") == "ghostty", getenv("TERM") == "xterm-ghostty":
		return imageKitty
	case getenv("TERM_PROGRAM") == "iTerm.app", getenv("LC_TERMINAL") == "iTerm2", getenv("TERM_PROGRAM") == "WezTerm":
		return imageITerm
	}
	switch strings.ToLower(getenv("COLORTERM")) {
	case "truecolor", "24bit":
		return imageBlocks
	}
	return imageBlocks256
}

// fit returns the width and height, of at least one, to fit an image of
// width by height pixels in cols by rows, of cellW by cellH pixels each,
// keeping its aspect, without making it larger than it is
func fit(width, height, cols, rows, cellW, cellH int) (int, int) {
	w := float64(width) / float64(cellW)
	h := float64(height) / float64(cellH)
	s := min(1, float64(cols)/w, float64(rows)/h)
	return max(1, int(w*s+0.5)), max(1, int(h*s+0.5))
}

// scale returns the image at width by height pixels
func scale(img image.Image, width, height int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Src, nil)
	return dst
}

// encode returns the image as PNG, made smaller when it's larger than
// width by height pixels
func encode(img image.Image, width, height int) ([]byte, error) {
	b := img.Bounds()
	if b.Dx() > width || b.Dy() > height {
		w, h := fit(b.Dx(), b.Dy(), width, height, 1, 1)
		img = scale(img, w, h)
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeKitty writes the PNG with kitty's protocol, in parts, displayed in
// cols by rows, and without the terminal's responses
func writeKitty(w io.Writer, data []byte, cols, rows int) error {
	payload := base64.StdEncoding.EncodeToString(data)
	for i := 0; i < len(payload); i += chunkSize {
		chunk := payload[i:min(i+chunkSize, len(payload))]
		more := 0
		if i+chunkSize < len(payload) {
			more = 1
		}
		var err error
		if i == 0 {
			_, err = fmt.Fprintf(w, "\x1b_Ga=T,f=100,q=2,c=%d,r=%d,m=%d;%s\x1b\\", cols, rows, more, chunk)
		} else {
			_, err = fmt.Fprintf(w, "\x1b_Gm=%d;%s\x1b\\", more, chunk)
		}
		if err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "\n")
	return err
}

// writeITerm writes the PNG with iTerm2's inline images, displayed in cols
// by rows
func writeITerm(w io.Writer, data []byte, cols, rows int) error {
	_, err := fmt.Fprintf(w, "\x1b]1337;File=inline=1;size=%d;width=%d;height=%d;preserveAspectRatio=1:%s\a\n", len(data), cols, rows, base64.StdEncoding.EncodeToString(data))
	return err
}

// writeBlocks writes the image as characters, the upper half block in the
// colour of a pixel on the background of the colour of the pixel beneath
// it, where a pixel which is mostly transparent is the terminal's
// background, and the colour is only written when it changes
func writeBlocks(w io.Writer, img *image.NRGBA, palette bool) error {
	var b strings.Builder
	b.Grow(img.Bounds().Dx() * 40)
	for y := 0; y < img.Bounds().Dy(); y += 2 {
		fg, bg := "", ""
		for x := 0; x < img.Bounds().Dx(); x++ {
			top, topOK := pixel(img, x, y)
			bottom, bottomOK := pixel(img, x, y+1)
			ch, f, g := "▀", sgr(top, false, palette), sgr(bottom, true, palette)
			switch {
			case !topOK && !bottomOK:
				ch, f, g = " ", "", "49"
			case !topOK:
				ch, f, g = "▄", sgr(bottom, false, palette), "49"
			case !bottomOK:
				g = "49"
			}
			if g != bg {
				b.WriteString("\x1b[" + g + "m")
				bg = g
			}
			if f != "" && f != fg {
				b.WriteString("\x1b[" + f + "m")
				fg = f
			}
			b.WriteString(ch)
		}
		b.WriteString("\x1b[0m\n")
		if _, err := io.WriteString(w, b.String()); err != nil {
			return err
		}
		b.Reset()
	}
	return nil
}

// pixel returns the colour of a pixel, and false when it's mostly
// transparent, or beyond the image
func pixel(img *image.NRGBA, x, y int) (color.NRGBA, bool) {
	if !(image.Point{x, y}.In(img.Bounds())) {
		return color.NRGBA{}, false
	}
	c := img.NRGBAAt(x, y)
	return c, c.A >= 128
}

// sgr returns the parameters which set the foreground, or background,
// colour
func sgr(c color.NRGBA, background, palette bool) string {
	switch {
	case palette && background:
		return fmt.Sprintf("48;5;%d", index256(c))
	case palette:
		return fmt.Sprintf("38;5;%d", index256(c))
	case background:
		return fmt.Sprintf("48;2;%d;%d;%d", c.R, c.G, c.B)
	default:
		return fmt.Sprintf("38;2;%d;%d;%d", c.R, c.G, c.B)
	}
}

// index256 returns the nearest of xterm's 256 colours, of its cube of 216
// colours and its 24 greys
func index256(c color.NRGBA) int {
	levels := [6]int{0, 95, 135, 175, 215, 255}
	nearest := func(v uint8) int {
		best := 0
		for i, l := range levels {
			if abs(int(v)-l) < abs(int(v)-levels[best]) {
				best = i
			}
		}
		return best
	}
	r, g, b := nearest(c.R), nearest(c.G), nearest(c.B)
	cube := 16 + 36*r + 6*g + b
	cubeDist := sq(int(c.R)-levels[r]) + sq(int(c.G)-levels[g]) + sq(int(c.B)-levels[b])

	// The nearest grey, from 8 to 238 in steps of 10
	grey := min(23, max(0, ((int(c.R)+int(c.G)+int(c.B))/3-3)/10))
	l := 8 + 10*grey
	if sq(int(c.R)-l)+sq(int(c.G)-l)+sq(int(c.B)-l) < cubeDist {
		return 232 + grey
	}
	return cube
}

func abs(v int) int { return max(v, -v) }
func sq(v int) int  { return v * v }
