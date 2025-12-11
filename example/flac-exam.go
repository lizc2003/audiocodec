package main

import (
	"fmt"
	"os"

	"github.com/lizc2003/audiocodec"
)

func main() {
	inFile, err := os.Open("samples/sample.flac")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inFile.Close()

	wavFile, err := os.Create("output.wav")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer wavFile.Close()

	totalBytes, totalSamples, sampleRate, err := audiocodec.FlacToWav(inFile, wavFile)
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("Decoded %d bytes, totalSamples: %d, sampleRate: %d\n", totalBytes, totalSamples, sampleRate)
}
