package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UTC().UnixNano())

	var dimensions string
	flag.StringVar(&dimensions, "dimensions", "800x600", "image dimensions, separate values with an x, single value without an x will be a square image")

	var background string
	flag.StringVar(&background, "background", "0,0,0", "background colour in rgb format (0,0,0 = black, 255,255,255 = white)")

	var pixel int
	flag.IntVar(&pixel, "pixel", 10, `"pixel" size`)

	flag.Parse()

	x, y, err := ParseDimensions(dimensions)
	if err != nil {
		panic(err)
	}

	bg, err := ParseColour(background)
	if err != nil {
		panic(err)
	}

	logo := []image.Point{
		{0, 0}, {4, 0},
		{0, 1}, {1, 1}, {3, 1}, {4, 1},
		{0, 2}, {2, 2}, {4, 2},
		{1, 3}, {3, 3},
	}

	img := image.NewRGBA(image.Rect(0, 0, x, y))
	draw.Draw(img, img.Bounds(), &image.Uniform{bg}, image.Point{0, 0}, draw.Src)

	ib := img.Bounds()
	iw, ih := ib.Max.X, ib.Max.Y

	lb := GetPointListBounds(logo, pixel)
	lw, lh := lb.Max.X, lb.Max.Y

	dx, dy := (iw-lw)/2, (ih-lh)/2

	DrawBackground(img, logo, pixel)

	DrawLogo(img, logo, color.Palette{color.Gray{255}, color.Gray{245}, color.Gray{235}, color.Gray{225}, color.Gray{215}}, dx, dy, pixel)

	f, err := os.Create(fmt.Sprintf("background-%dx%d-%dpx-%x.png", x, y, pixel, md5.Sum(img.Pix)))
	if err != nil {
		panic(err)
	}

	if err := png.Encode(f, img); err != nil {
		panic(err)
	}
}

func DrawLogo(img draw.Image, logo []image.Point, colours color.Palette, x, y, px int) {
	for _, pt := range logo {
		c := colours[rand.Intn(len(colours))]
		x1, x2, y1, y2 := pt.X*px+x, pt.X*px+px+x, pt.Y*px+y, pt.Y*px+px+y

		draw.Draw(img, image.Rect(x1, y1, x2, y2), &image.Uniform{c}, image.Point{0, 0}, draw.Src)
	}
}

func DrawBackground(img *image.RGBA, logo []image.Point, px int) {
	ib := img.Bounds()
	iw, ih := ib.Max.X, ib.Max.Y

	lb := GetPointListBounds(logo, px)
	lw, lh := lb.Max.X, lb.Max.Y

	// get starting x/y position of logo to perfectly center it
	dx, dy := (iw-lw)/2, (ih-lh)/2

	// find number of logo tesselations to cover background (+ a few more for luck)
	wt, ht := IntDivideCeil(iw+dx, lw), IntDivideCeil(ih+dy, lh)
	t := wt * ht

	// spiral out from center logo drawing "background" logos
	x, y := 0, 0
	xd, yd := 0, -1

	palettes := []color.Palette{
		{color.Gray{6}, color.Gray{7}, color.Gray{8}, color.Gray{9}, color.Gray{10}},
		{color.Gray{11}, color.Gray{12}, color.Gray{13}, color.Gray{14}, color.Gray{15}},
	}

	for i := 0; i < t; i++ {
		if (-wt/2 <= x && x <= wt/2) && (-ht/2 <= y && y <= ht/2) {
			lx, ly := dx+(x*lw), dy+(y*lh)
			DrawLogo(img, logo, palettes[i%2], lx, ly, px)
		}

		if x == y || (x < 0 && x == -y) || (x > 0 && x == 1-y) {
			xd, yd = -yd, xd
		}

		x, y = x+xd, y+yd
	}
}

func IntDivideCeil(a, b int) int {
	return int(math.Ceil(float64(a) / float64(b)))
}

func IntMax(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func GetPointListBounds(pts []image.Point, px int) image.Rectangle {
	mx, my := 0, 0
	for _, pt := range pts {
		mx = IntMax(pt.X, mx)
		my = IntMax(pt.Y, my)
	}

	w, h := mx*px, my*px

	return image.Rect(0, 0, w+px, h+px)
}

// ParseDimensions string into int x, y image dimensions
func ParseDimensions(in string) (int, int, error) {
	xy := strings.SplitN(in, "x", 2)

	if len(xy) == 2 {
		x, err := strconv.Atoi(xy[0])
		if err != nil {
			return 0, 0, nil
		}

		y, err := strconv.Atoi(xy[1])
		if err != nil {
			return 0, 0, nil
		}

		return x, y, nil
	}

	x, err := strconv.Atoi(xy[0])
	if err != nil {
		return 0, 0, err
	}

	return x, x, nil
}

// ParseColour string into a color.Color representation
func ParseColour(in string) (color.Color, error) {
	rgb := strings.SplitN(in, ",", 3)

	r, err := strconv.Atoi(rgb[0])
	if err != nil {
		return nil, err
	}

	g, err := strconv.Atoi(rgb[1])
	if err != nil {
		return nil, err
	}

	b, err := strconv.Atoi(rgb[2])
	if err != nil {
		return nil, err
	}

	// default to 100% opaque
	return color.RGBA{uint8(r), uint8(g), uint8(b), 0xff}, nil
}
