package main

// inspired by: https://wiki.openstreetmap.org/wiki/Slippy_map_tilenames

import (
	"fmt"
	"net/http"
	"image"
	"bytes"
	"image/jpeg"
	"strconv"
	"strings"
	//"image/color"
	//"image/draw"
	"os"
	"math"
)

type Tile struct {
	Z    int
	X    int
	Y    int
	Lat  float64
	Long float64
}

type Conversion interface {
	deg2num(t *Tile) (x int, y int)
	num2deg(t *Tile) (lat float64, long float64)
}

// not needed
// func (*Tile) Deg2num(t *Tile) (x int, y int) {
// 	x = int(math.Floor((t.Long + 180.0) / 360.0 * (math.Exp2(float64(t.Z)))))
// 	y = int(math.Floor((1.0 - math.Log(math.Tan(t.Lat*math.Pi/180.0)+1.0/math.Cos(t.Lat*math.Pi/180.0))/math.Pi) / 2.0 * (math.Exp2(float64(t.Z)))))
// 	return
// }

func (*Tile) Num2deg(t *Tile) (lat float64, long float64) {
	n := math.Pi - 2.0*math.Pi*float64(t.Y)/math.Exp2(float64(t.Z))
	lat = 180.0 / math.Pi * math.Atan(0.5*(math.Exp(n)-math.Exp(-n)))
	long = float64(t.X)/math.Exp2(float64(t.Z))*360.0 - 180.0
	return lat, long
}

func main() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

var ImageTemplate string = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><img src="data:image/jpg;base64,{{.Image}}"></body>`

func HelloServer(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])

	s := strings.Split(r.URL.Path, "/")

	// fmt.Println(s[1])
	// fmt.Println(s[2])
	// fmt.Println(s[3])

	// fmt.Println("")

	zoom, err := strconv.Atoi(s[1])
	x, err := strconv.Atoi(s[2])
	y, err := strconv.Atoi(s[3])
	

	if err != nil {
		panic(err)
	}



	tile := Tile{X: x, Y: y, Z: zoom}

	lat, lon := tile.Num2deg(&tile)


	// fmt.Printf("x: %d, y: %d, zoom: %d\n", x, y, zoom)
	
	
	fmt.Printf("Lat: %f, Lon: %f, zoom: %d\n", lat, lon, zoom)


	//m := image.NewRGBA(image.Rect(0, 0, 240, 240))
	//blue := color.RGBA{0, 0, 255, 255}
	//draw.Draw(m, m.Bounds(), &image.Uniform{blue}, image.ZP, draw.Src)



	//var img image.Image = m
	writeImage(w, getImage())



}

func getImage() *image.Image {

	imgfile, err := os.Open("funnycat.jpeg")
	if err != nil {
		fmt.Println("img.jpg file not found!")
		os.Exit(1)
	}
	defer imgfile.Close()

	img, _, err := image.Decode(imgfile)
	//bounds := img.Bounds()
	//canvas := image.NewAlpha(bounds)

	// is this image opaque
	//op := canvas.Opaque()

	//fmt.Println(op)

	return &img
}

func writeImage(w http.ResponseWriter, img *image.Image) {

	buffer := new(bytes.Buffer)
	if err := jpeg.Encode(buffer, *img, nil); err != nil {
		fmt.Println("unable to encode image.")
	}

	w.Header().Set("Content-Type", "image/jpeg")
	w.Header().Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))
	if _, err := w.Write(buffer.Bytes()); err != nil {
		fmt.Println("unable to write image.")
	}
}