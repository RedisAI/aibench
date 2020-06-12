package main

import (
	"bufio"
	"flag"
	"fmt"
	//"github.com/RedisAI/redisai-go/redisai"
	"github.com/RedisAI/redisai-go/redisai/implementations"
	"image"
	"image/jpeg"
	"io"
	"io/ioutil"
	"log"
	"os"
)

// Program option vars:
var (
	inputDir                  string
	outputFileName	string
	defaultWriteSize = 4 << 20 // 4 MB
)

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) (R, G, B, A uint8) {
	return uint8(r / 257), uint8(g / 257), uint8(b / 257), uint8(a / 257)
}

func JPEGImageToAiTensor(img image.Image) ([]uint8, *implementations.AITensor) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var pixels []uint8
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := rgbaToPixel(img.At(x, y).RGBA())
			pixels = append(pixels, r, g, b, a)
		}
	}
	// Build a tensor
	tensor := implementations.NewAiTensor()
	tensor.SetShape([]int64{int64(width), int64(height), 4})
	tensor.SetData(pixels)
	return pixels, tensor
}


// GetBufferedWriter returns the buffered Writer that should be used for generated output
func GetBufferedWriter(fileName string) *bufio.Writer {
	// Prepare output file/STDOUT
	if len(fileName) > 0 {
		// Write output to file
		file, err := os.Create(fileName)
		if err != nil {
			log.Fatalf("cannot open file for write %s: %v", fileName, err)
		}
		return bufio.NewWriterSize(file, defaultWriteSize)
	}

	// Write output to STDOUT
	return bufio.NewWriterSize(os.Stdout, defaultWriteSize)
}



// Serialize writes Transaction data to the given writer, in a format that will be easy to create a RedisAI command
func  SerializeTensorData(pixels []byte, w io.Writer) (err error) {
	var buf []byte
	buf = append(buf, pixels...)
	_, err = w.Write(buf)
	return err
}


func main() {
	flag.StringVar(&inputDir, "input-val-dir", ".", fmt.Sprintf(""))
	flag.StringVar(&outputFileName, "output-file", "", "File name to write generated data to")
	flag.Parse()

	// Get output writer
	out := GetBufferedWriter(outputFileName)
	defer func() {
		err := out.Flush()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()

	items, _ := ioutil.ReadDir(inputDir)
	for _, item := range items {
		if !item.IsDir() {
			fmt.Println(item.Name())

			// Read image from file that already exists
			imageFile, err := os.Open(fmt.Sprintf("%s/%s",inputDir,item.Name()))
			if err != nil {
				log.Fatal(err)
			}
			img, err := jpeg.Decode(imageFile)
			pixels, _ := JPEGImageToAiTensor(img)
			SerializeTensorData(pixels,out)
			//
			//fmt.Println(len(pixels))
			//
			//// Create a simple client.
			//client := redisai.Connect("redis://localhost:6379", nil)
			//
			//// Set a tensor
			//// AI.TENSORSET foo UINT8 224 224 4 BLOB ....
			//err = client.TensorSetFromTensor(item.Name(), tensor)

		}
	}
}
