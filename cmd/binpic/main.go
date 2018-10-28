package main

import (
	"bufio"
	"flag"
	"image"
	"image/color"
	"image/png"
	"io"
	"log"
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/disintegration/imaging"
)

var (
	dims   = flag.String("resize", "0x0", "resize, if set")
	output = flag.String("o", "output.png", "output file, will be a PNG")
)

// parseDims parses dimensions or returns 0, if there was an error parsing.
func parseDims(s string) (width, height int) {
	parts := strings.Split(s, "x")
	if len(parts) != 2 {
		return 0, 0
	}
	var err error
	width, err = strconv.Atoi(strings.TrimSpace(parts[0]))
	if err != nil {
		return 0, 0
	}
	height, err = strconv.Atoi(strings.TrimSpace(parts[1]))
	if err != nil {
		return 0, 0
	}
	return width, height
}

// dimsFromSize returns image dimensions given number of pixels.
func dimsFromSize(size int64, pct float64) (width, height int) {
	sizef := float64(size)
	sq := math.Sqrt(sizef)
	h := math.Ceil(sq - sq*pct)
	w := math.Ceil(sizef / h)
	return int(w), int(h)
}

func main() {
	flag.Parse()
	if flag.NArg() == 0 {
		log.Fatal("input file required")
	}

	filename := flag.Arg(0)
	fi, err := os.Stat(filename)
	if err != nil {
		log.Fatal(err)
	}
	size := fi.Size()
	w, h := dimsFromSize(size, 0.15)
	log.Printf("%d, %dx%d, %d\n", size, w, h, w*h)

	// A Rectangle contains the points with Min.X <= X < Max.X, Min.Y <= Y <
	// Max.Y. It is well-formed if Min.X <= Max.X and likewise for Y. Points
	// are always well-formed. A rectangle's methods always return well-formed
	// outputs for well-formed inputs.
	rect := image.Rectangle{
		Min: image.Point{X: 0, Y: 0}, // up left
		Max: image.Point{X: w, Y: h}, // down right
	}
	img := image.NewGray(rect)

	f, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	br := bufio.NewReader(f)

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			b, err := br.ReadByte()
			if err == io.EOF {
				img.Set(x, y, color.Gray{255})
				continue
			}
			if err != nil {
				log.Fatal(err)
			}
			img.Set(x, y, color.Gray{b})
		}
	}

	// Write out.
	fout, err := os.Create(*output)
	if err != nil {
		log.Fatal(err)
	}
	defer fout.Close()

	// Resize, if requested.
	resizeWidth, resizeHeight := parseDims(*dims)
	switch {
	case resizeWidth > 0 && resizeHeight > 0:
		resized := imaging.Resize(img, resizeWidth, resizeHeight, imaging.Lanczos)
		if err := png.Encode(fout, resized); err != nil {
			log.Fatal(err)
		}
	default:
		if err := png.Encode(fout, img); err != nil {
			log.Fatal(err)
		}
	}

}