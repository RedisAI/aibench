package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"fmt"
	"github.com/cheggaaa/pb/v3"
	"math"

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
	inputDir         string
	outputFileName   string
	defaultWriteSize = 4 << 20 // 4 MB
)

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixel(r uint32, g uint32, b uint32, a uint32) (R, G, B, A uint8) {
	return uint8(r / 257), uint8(g / 257), uint8(b / 257), uint8(a / 257)
}

// img.At(x, y).RGBA() returns four uint32 values; we want a Pixel
func rgbaToPixelFloat32(r uint32, g uint32, b uint32, a uint32, scale float32) (R, G, B, A float32) {
	return float32(r/257) * scale, float32(g/257) * scale, float32(b/257) * scale, float32(a/257) * scale
}

func getRGBAPos_CxHxW(y int, width int, x int, height int) (int, int, int, int) {
	r_pos := y*width + x
	g_pos := (height * width) + r_pos
	b_pos := (2 * height * width) + r_pos
	a_pos := (3 * height * width) + r_pos
	return r_pos, g_pos, b_pos, a_pos
}

// converts the image to a tensor with a H x W x C layout.
func JPEGImageTo_HxWxC_float32_AiTensor(img image.Image, useAlpha bool, scale float32) ([]float32, *implementations.AITensor) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var numChannels int64 = 4
	if !useAlpha {
		numChannels = 3
	}
	var pixels = make([]float32, 0, height*width*int(numChannels))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			uir, uig, uib, uia := img.At(x, y).RGBA()
			r, g, b, a := rgbaToPixelFloat32(uir, uig, uib, uia, scale)
			if useAlpha {
				pixels = append(pixels, r, g, b, a)
			} else {
				pixels = append(pixels, r, g, b)
			}
		}
	}
	// Build a tensor
	tensor := implementations.NewAiTensor()
	tensor.SetShape([]int64{int64(height), int64(width), numChannels})
	tensor.SetData(pixels)
	return pixels, tensor
}

// converts the image to a tensor with a H x W x C layout.
func JPEGImageTo_HxWxC_uint8_AiTensor(img image.Image, useAlpha bool) ([]uint8, *implementations.AITensor) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var numChannels int64 = 4
	if !useAlpha {
		numChannels = 3
	}
	var pixels = make([]uint8, 0, height*width*int(numChannels))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := rgbaToPixel(img.At(x, y).RGBA())
			if useAlpha {
				pixels = append(pixels, r, g, b, a)
			} else {
				pixels = append(pixels, r, g, b)
			}
		}
	}
	// Build a tensor
	tensor := implementations.NewAiTensor()
	tensor.SetShape([]int64{int64(height), int64(width), numChannels})
	tensor.SetData(pixels)
	return pixels, tensor
}

// converts the image to a tensor with a C x H x W layout.
func JPEGImageTo_CxHxW_uint8_AiTensor(img image.Image, useAlpha bool) ([]uint8, *implementations.AITensor) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var numChannels int64 = 4
	if !useAlpha {
		numChannels = 3
	}
	var pixels = make([]uint8, height*width*int(numChannels), height*width*int(numChannels))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			r, g, b, a := rgbaToPixel(img.At(x, y).RGBA())
			r_pos, g_pos, b_pos, a_pos := getRGBAPos_CxHxW(y, width, x, height)
			pixels[r_pos] = r
			pixels[g_pos] = g
			pixels[b_pos] = b
			if useAlpha {
				pixels[a_pos] = a
			}
		}
	}

	// Build a tensor
	tensor := implementations.NewAiTensor()
	tensor.SetShape([]int64{numChannels, int64(height), int64(width)})
	tensor.SetData(pixels)
	return pixels, tensor
}

// converts the image to a tensor with a C x H x W layout.
func JPEGImageTo_CxHxW_float32_AiTensor(img image.Image, useAlpha bool, scale float32) ([]float32, *implementations.AITensor) {
	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	var numChannels int64 = 4
	if !useAlpha {
		numChannels = 3
	}
	var pixels = make([]float32, height*width*int(numChannels), height*width*int(numChannels))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			uir, uig, uib, uia := img.At(x, y).RGBA()
			r, g, b, a := rgbaToPixelFloat32(uir, uig, uib, uia, scale)
			r_pos, g_pos, b_pos, a_pos := getRGBAPos_CxHxW(y, width, x, height)
			pixels[r_pos] = r
			pixels[g_pos] = g
			pixels[b_pos] = b
			if useAlpha {
				pixels[a_pos] = a
			}
		}
	}

	// Build a tensor
	tensor := implementations.NewAiTensor()
	tensor.SetShape([]int64{numChannels, int64(height), int64(width)})
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
func SerializeTensorData(pixels []byte, w io.Writer) (err error) {
	var buf []byte
	buf = append(buf, pixels...)
	_, err = w.Write(buf)
	return err
}

func Float32bytes(float float32) []byte {
	bits := math.Float32bits(float)
	bytes := make([]byte, 4)
	binary.LittleEndian.PutUint32(bytes, bits)
	return bytes
}

func SerializeTensorDataFloat32(pixels []float32, w io.Writer) (err error) {
	var buf []byte
	for _, value := range pixels {
		buf = append(buf, Float32bytes(value)...)
	}
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
	bar := pb.StartNew(len(items))
	for _, item := range items {
		if !item.IsDir() {

			// Read image from file that already exists
			imageFile, err := os.Open(fmt.Sprintf("%s/%s", inputDir, item.Name()))
			if err != nil {
				log.Fatal(err)
			}
			img, err := jpeg.Decode(imageFile)
			pixels, _ := JPEGImageTo_HxWxC_float32_AiTensor(img, false, 1.0/255.0)
			SerializeTensorDataFloat32(pixels, out)
			bar.Increment()

		}
	}
	bar.Finish()
}