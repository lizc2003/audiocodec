package audiocodec

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	WavHeaderSize = 44
)

func Int32toInt24LEBytes(n int32) []byte {
	bytes := make([]byte, 3)
	if (n & 0x800000) > 0 {
		n |= ^0xffffff
	}
	bytes[2] = byte(n >> 16)
	bytes[1] = byte(n >> 8)
	bytes[0] = byte(n)
	return bytes
}

func GenerateWavHeader(pcmSize int, sampleRate int, numChannels int, bitsPerSample int) []byte {
	header := make([]byte, WavHeaderSize)
	byteRate := sampleRate * numChannels * bitsPerSample / 8
	blockAlign := numChannels * bitsPerSample / 8

	// RIFF
	copy(header[0:4], []byte("RIFF"))
	binary.LittleEndian.PutUint32(header[4:8], uint32(36+pcmSize))
	copy(header[8:12], []byte("WAVE"))

	// fmt
	copy(header[12:16], []byte("fmt "))
	binary.LittleEndian.PutUint32(header[16:20], 16) // Subchunk1Size for PCM
	binary.LittleEndian.PutUint16(header[20:22], 1)  // AudioFormat 1 = PCM
	binary.LittleEndian.PutUint16(header[22:24], uint16(numChannels))
	binary.LittleEndian.PutUint32(header[24:28], uint32(sampleRate))
	binary.LittleEndian.PutUint32(header[28:32], uint32(byteRate))
	binary.LittleEndian.PutUint16(header[32:34], uint16(blockAlign))
	binary.LittleEndian.PutUint16(header[34:36], uint16(bitsPerSample))

	// data
	copy(header[36:40], []byte("data"))
	binary.LittleEndian.PutUint32(header[40:44], uint32(pcmSize))

	return header
}

func ParseWavHeader(wavStream io.Reader) (pcmSize int, sampleRate int, numChannels int, bitsPerSample int, err error) {
	var (
		riffHeader    [12]byte
		chunkHeader   [8]byte
		fmtChunkFound bool
	)

	// Read RIFF header
	if _, err := io.ReadFull(wavStream, riffHeader[:]); err != nil {
		return 0, 0, 0, 0, fmt.Errorf("read RIFF header failed: %w", err)
	}
	if string(riffHeader[0:4]) != "RIFF" || string(riffHeader[8:12]) != "WAVE" {
		return 0, 0, 0, 0, errors.New("invalid WAV header: missing RIFF/WAVE")
	}

	// Loop chunks
	for {
		if _, err := io.ReadFull(wavStream, chunkHeader[:]); err != nil {
			return 0, 0, 0, 0, fmt.Errorf("read chunk header failed: %w", err)
		}
		chunkID := string(chunkHeader[0:4])
		chunkSize := binary.LittleEndian.Uint32(chunkHeader[4:8])

		if chunkID == "fmt " {
			if chunkSize < 16 {
				return 0, 0, 0, 0, fmt.Errorf("invalid fmt chunk size: %d", chunkSize)
			}
			fmtData := make([]byte, chunkSize)
			if _, err := io.ReadFull(wavStream, fmtData); err != nil {
				return 0, 0, 0, 0, fmt.Errorf("read fmt chunk failed: %w", err)
			}

			audioFormat := binary.LittleEndian.Uint16(fmtData[0:2])
			numChannels = int(binary.LittleEndian.Uint16(fmtData[2:4]))
			sampleRate = int(binary.LittleEndian.Uint32(fmtData[4:8]))
			bitsPerSample = int(binary.LittleEndian.Uint16(fmtData[14:16]))

			if audioFormat != 1 {
				return 0, 0, 0, 0, fmt.Errorf("unsupported audio format: %d (only PCM supported)", audioFormat)
			}
			fmtChunkFound = true
		} else if chunkID == "data" {
			if !fmtChunkFound {
				return 0, 0, 0, 0, errors.New("data chunk found before fmt chunk")
			}
			// We found data chunk, stop parsing.
			pcmSize = int(chunkSize)
			break
		} else {
			// Skip other chunks
			if _, err := io.CopyN(io.Discard, wavStream, int64(chunkSize)); err != nil {
				return 0, 0, 0, 0, fmt.Errorf("skip chunk %s failed: %w", chunkID, err)
			}
		}
	}
	return pcmSize, sampleRate, numChannels, bitsPerSample, nil
}
