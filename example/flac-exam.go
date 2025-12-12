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
	if hex.EncodeToString(md5sum[:]) != "24d62c157971605e42defb9954d65e33" {
		fmt.Println("wav md5sum check wrong")
		return
	}

	md5sum = md5.Sum(bytes[audiocodec.WavHeaderSize:])
	if hex.EncodeToString(md5sum[:]) != "372fd5b0a07a0a78b92311ccdca4cc81" {
		fmt.Println("pcm md5sum check wrong")
		return
	}

	fmt.Printf("Write to Memory, Decoded %d bytes, totalSamples: %d, sampleRate: %d\n", totalBytes, totalSamples, sampleRate)
}
