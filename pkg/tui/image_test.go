package tui

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/png"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImageProtocol(t *testing.T) {
	// The protocol is the terminal's, except in tmux, or else characters in
	// 24-bit colour when the terminal has it, or in 256 colours
	for _, test := range []struct {
		env  map[string]string
		want imageProtocol
	}{
		{map[string]string{"TERM": "xterm-kitty"}, imageKitty},
		{map[string]string{"TERM_PROGRAM": "ghostty", "COLORTERM": "truecolor"}, imageKitty},
		{map[string]string{"TERM_PROGRAM": "iTerm.app"}, imageITerm},
		{map[string]string{"LC_TERMINAL": "iTerm2"}, imageITerm},
		{map[string]string{"TERM_PROGRAM": "iTerm.app", "TMUX": "/tmp/tmux", "COLORTERM": "truecolor"}, imageBlocks},
		{map[string]string{"TERM": "xterm-256color", "KITTY_WINDOW_ID": "1", "STY": "1234.pts-0.host"}, imageBlocks256},
		{map[string]string{"TERM_PROGRAM": "Apple_Terminal", "TERM": "xterm-256color"}, imageBlocks256},
		{map[string]string{"TERM_PROGRAM": "vscode", "COLORTERM": "truecolor"}, imageBlocks},
	} {
		assert.Equal(t, test.want, detectImageProtocol(func(k string) string { return test.env[k] }), test.env)
	}
}

func TestImageFit(t *testing.T) {
	// An image fits in the columns and rows, keeping its aspect, and isn't
	// made larger
	for _, test := range []struct {
		width, height, cols, rows, wantCols, wantRows int
	}{
		{10, 10, 80, 48, 10, 10},
		{1000, 500, 80, 48, 80, 40},
		{500, 1000, 80, 48, 24, 48},
		{1, 10000, 80, 48, 1, 48},
	} {
		c, r := fit(test.width, test.height, test.cols, test.rows, 1, 1)
		assert.Equal(t, []int{test.wantCols, test.wantRows}, []int{c, r}, "%dx%d in %dx%d", test.width, test.height, test.cols, test.rows)
	}
}

func TestImageBlocks(t *testing.T) {
	// Two pixels are a character, in their colours, and a transparent pixel
	// is the terminal's background
	img := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	img.SetNRGBA(0, 0, color.NRGBA{255, 0, 0, 255})
	img.SetNRGBA(0, 1, color.NRGBA{0, 0, 255, 255})
	img.SetNRGBA(1, 1, color.NRGBA{0, 255, 0, 255})
	var b strings.Builder
	n, err := (&imageFormatter{width: 80, height: 24, protocol: imageBlocks}).Write(&b, img)
	require.NoError(t, err)
	assert.Equal(t, "\x1b[48;2;0;0;255m\x1b[38;2;255;0;0m▀\x1b[49m\x1b[38;2;0;255;0m▄\x1b[0m\n", b.String())
	assert.Equal(t, b.Len(), n)
}

func TestImageBlocksSize(t *testing.T) {
	// As characters, a pixel is at most half a character, so a small image
	// is drawn with a character for every two of its pixels, and a large one
	// fits in the width and height
	for _, test := range []struct {
		width, height, wantLines, wantCols int
	}{
		{16, 32, 16, 16},
		{1000, 500, 10, 40},
	} {
		var b strings.Builder
		_, err := (&imageFormatter{width: 40, height: 24, protocol: imageBlocks256}).Write(&b, image.NewNRGBA(image.Rect(0, 0, test.width, test.height)))
		require.NoError(t, err)
		lines := strings.Split(strings.TrimSuffix(b.String(), "\n"), "\n")
		assert.Len(t, lines, test.wantLines, "%dx%d", test.width, test.height)
		assert.Equal(t, test.wantCols, strings.Count(lines[0], " "), "%dx%d", test.width, test.height)
	}
}

func TestImage256(t *testing.T) {
	// The nearest of the 256 colours is of the cube, or a grey
	for _, test := range []struct {
		c    color.NRGBA
		want int
	}{
		{color.NRGBA{0, 0, 0, 255}, 16},
		{color.NRGBA{255, 255, 255, 255}, 231},
		{color.NRGBA{255, 0, 0, 255}, 196},
		{color.NRGBA{128, 128, 128, 255}, 244},
	} {
		assert.Equal(t, test.want, index256(test.c), test.c)
	}
}

func TestImageKitty(t *testing.T) {
	// An image is sent with kitty's protocol in parts, which are the PNG,
	// of noise, so it's larger than a part
	img := image.NewNRGBA(image.Rect(0, 0, 200, 200))
	r := rand.New(rand.NewPCG(1, 2))
	for i := range img.Pix {
		img.Pix[i] = byte(r.Uint32())
	}
	var b strings.Builder
	_, err := (&imageFormatter{width: 80, height: 24, protocol: imageKitty}).Write(&b, img)
	require.NoError(t, err)

	parts := strings.Split(strings.TrimSuffix(b.String(), "\n"), "\x1b\\")
	require.Greater(t, len(parts), 2)
	assert.True(t, strings.HasPrefix(parts[0], "\x1b_Ga=T,f=100,q=2,c=25,r=13,m=1;"))
	assert.True(t, strings.HasPrefix(parts[len(parts)-2], "\x1b_Gm=0;"))

	var payload strings.Builder
	for _, part := range parts[:len(parts)-1] {
		payload.WriteString(part[strings.IndexByte(part, ';')+1:])
	}
	data, err := base64.StdEncoding.DecodeString(payload.String())
	require.NoError(t, err)
	decoded, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	assert.Equal(t, img.Bounds(), decoded.Bounds())
}

func TestImageITerm(t *testing.T) {
	// An image is sent with iTerm2's inline images, as the PNG, in the
	// characters it's displayed in
	img := image.NewNRGBA(image.Rect(0, 0, 16, 32))
	var b strings.Builder
	_, err := (&imageFormatter{width: 80, height: 24, protocol: imageITerm}).Write(&b, img)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(b.String(), "\x1b]1337;File=inline=1;size="))
	assert.Contains(t, b.String(), ";width=2;height=2;preserveAspectRatio=1:")
	assert.True(t, strings.HasSuffix(b.String(), "\a\n"))
}
