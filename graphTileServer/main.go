package main

import (
	"fmt"
	"net/http"
	"image"
	"bytes"
	"image/jpeg"
	"strconv"
	//"image/color"
	//"image/draw"
	"os"
)

func main() {
	http.HandleFunc("/", HelloServer)
	http.ListenAndServe(":8080", nil)
}

var ImageTemplate string = `<!DOCTYPE html>
<html lang="en"><head></head>
<body><img src="data:image/jpg;base64,{{.Image}}"></body>`

func HelloServer(w http.ResponseWriter, r *http.Request) {
	//fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])

	fmt.Println(r.URL.Path[1:])

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