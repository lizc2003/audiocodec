package main

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/lizc2003/audiocodec"
	"os"
)

func main() {
	flacToWavFile()
	flacToWavMemory()
}

func flacToWavFile() {
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

	fmt.Printf("Write to File, Decoded %d bytes, totalSamples: %d, sampleRate: %d\n", totalBytes, totalSamples, sampleRate)
}

func flacToWavMemory() {
	inFile, err := os.Open("samples/sample.flac")
	if err != nil {
		fmt.Println(err)
		return
	}
	defer inFile.Close()

	writer := &audiocodec.WriterSeeker{}
	totalBytes, totalSamples, sampleRate, err := audiocodec.FlacToWav(inFile, writer)
	if err != nil {
		fmt.Println(err)
		return
	}
	bytes := writer.Bytes()

	md5sum := md5.Sum(bytes)
	fmt.Println("wav md5sum", hex.EncodeToString(md5sum[:]))

	md5sum = md5.Sum(bytes[audiocodec.WavHeaderSize:])
	fmt.Println("pcm md5sum", hex.EncodeToString(md5sum[:]))

	fmt.Printf("Write to Memory, Decoded %d bytes, totalSamples: %d, sampleRate: %d\n", totalBytes, totalSamples, sampleRate)
}
