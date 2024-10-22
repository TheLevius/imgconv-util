package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/davidbyttow/govips/v2/vips"
)

type Target struct {
	height int
	widthAspectRatio int
	heightAspectRatio int
}
type Frame struct {
	cropWidth int
	cropHeight int
	xOffset int
	yOffset int
}

func main() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter the file extension without dots (e.g. heic, png): ")
	extName, _ := reader.ReadString('\n')
	extName = strings.TrimSpace(extName)
	fmt.Printf("Вы выбрали расширение: %s\n", extName)

	ext := "." + extName
	outExtName := "webp"
	outExt := "." + outExtName
	
	fmt.Print("Enter the input and output folder separated by a space: ")
	dirsLn, _ := reader.ReadString('\n')
	dirs := strings.Split(strings.TrimSpace(dirsLn), " ")
	dirInput := dirs[0]
	dirOutput := dirs[1]
	err := os.MkdirAll(dirOutput, 0755)
	if err != nil {
		fmt.Printf("mkdir error: %v\n", err)
		return
	}

	fmt.Print("Enter the height and aspect ratio via a space (e.g. 1280 4 3): ")
	paramsInput, _ := reader.ReadString('\n')
	strNumParams := strings.Split(strings.TrimSpace(paramsInput), " ")
	
	numericParams := make([]int, len(strNumParams))
	for i, str := range strNumParams {
		num, err := strconv.Atoi(str)
		if err != nil {
			fmt.Printf("string conversion error '%s': %v\n", str, err)
			return
		}
		numericParams[i] = num
	}

	target := new(Target)
	target.height = numericParams[0]
	target.widthAspectRatio = numericParams[1]
	if len(numericParams) == 2 {
		target.heightAspectRatio = target.widthAspectRatio
	} else {
		target.heightAspectRatio = numericParams[2]
	}

	vips.Startup(nil)
	defer vips.Shutdown()

	files, err := os.ReadDir(dirInput)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		baseName := file.Name()
		inputPath := filepath.Join(dirInput, baseName);
		nameWithoutExt := strings.TrimSuffix(baseName, ext)
		outFileName := nameWithoutExt + outExt
		outputFilePath := filepath.Join(dirOutput, outFileName)
		image, err := vips.NewImageFromFile(inputPath)
		if err != nil {
			fmt.Printf("download image error: %v\n", err)
			return
		}
		frame := computeTarget(target, image.Width(), image.Height())
		err = image.ExtractArea(frame.xOffset, frame.yOffset, frame.cropWidth, frame.cropHeight)
		if err != nil {
			log.Fatalf("crop image error: %v", err)
		}
		scaleFactor := float64(target.height) / float64(frame.cropHeight)

		err = image.Resize(scaleFactor, vips.KernelLanczos3)
		if err != nil {
			log.Fatalf("scale image error: %v", err)
		}
		buf, _, err := image.ExportWebp(vips.NewWebpExportParams())
		if err != nil {
			fmt.Printf("export image error: %v\n", err)
			return
		}
		err = os.WriteFile(outputFilePath, buf, 0644)
		if err != nil {
			fmt.Printf("writing image error: %v\n", err)
			return
		}
	}

	fmt.Println("Images successfully processed and saved")
}

func computeTarget(target *Target, originalWidth, originalHeight int) Frame {
	cropWidth := 0
	cropHeight := 0
	if originalWidth*target.heightAspectRatio > originalHeight*target.widthAspectRatio {
		cropWidth = originalHeight * target.widthAspectRatio / target.heightAspectRatio
		cropHeight = originalHeight
	} else {
		cropWidth = originalWidth
		cropHeight = originalWidth * target.heightAspectRatio / target.widthAspectRatio
	}

	return Frame{
		cropWidth: cropWidth,
		cropHeight: cropHeight,
		xOffset: (originalWidth - cropWidth) / 2,
		yOffset: (originalHeight - cropHeight) / 2,
	}
}
