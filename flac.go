package audiocodec

import (
	"errors"
	"fmt"
	"io"

	"github.com/mewkiz/flac"
)

func FlacToWav(flacStream io.Reader, writer io.WriteSeeker) (totalBytes int, totalSamples int, sampleRate int, err error) {
	stream, err := flac.New(flacStream)
	if err != nil {
		return 0, 0, 0, err
	}
	//defer stream.Close()

	info := stream.Info
	switch info.BitsPerSample {
	case 8, 16, 24, 32:
	default:
		return 0, 0, 0, fmt.Errorf("unsupported BitsPerSample: %d", info.BitsPerSample)
	}

	bufSize := int(info.BlockSizeMax) * int(info.NChannels) * int(info.BitsPerSample) / 8
	if bufSize > 8192*16*4 {
		return 0, 0, 0, fmt.Errorf("buffer size too large: %d bytes", bufSize)
	}
	buf := make([]byte, 0, bufSize)
	for {
		frame, err := stream.ParseNext()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0, 0, 0, err
		}

		blockSize := int(frame.BlockSize)
		for _, sub := range frame.Subframes {
			if len(sub.Samples) < blockSize {
				return 0, 0, 0, fmt.Errorf("subframe contains only %d samples, expected %d", sub.NSamples, frame.BlockSize)
			}
		}

		for i := 0; i < blockSize; i++ {
			for _, sub := range frame.Subframes {
				switch frame.BitsPerSample {
				case 8:
					buf = append(buf, byte(sub.Samples[i]))
				case 16:
					n := int16(sub.Samples[i])
					buf = append(buf, byte(n), byte(n>>8))
				case 24:
					buf = append(buf, Int32toInt24LEBytes(sub.Samples[i])...)
				case 32:
					n := sub.Samples[i]
					buf = append(buf, byte(n), byte(n>>8), byte(n>>16), byte(n>>24))
				default:
					return 0, 0, 0, fmt.Errorf("unsupported BitsPerSample: %d", info.BitsPerSample)
				}
			}
		}

		if totalSamples == 0 {
			// Write placeholder WAV header
			headerBuf := make([]byte, WavHeaderSize)
			if _, err := writer.Write(headerBuf); err != nil {
				return 0, 0, 0, fmt.Errorf("write placeholder header failed: %w", err)
			}
		}

		if _, wErr := writer.Write(buf); wErr != nil {
			return 0, 0, 0, wErr
		}

		totalBytes += len(buf)
		totalSamples += blockSize
		buf = buf[:0]
	}

	if totalSamples == 0 {
		return 0, 0, 0, errors.New("no audio frames decoded")
	}

	// Update WAV header if seeker
	if _, err := writer.Seek(0, io.SeekStart); err != nil {
		// Can't seek, maybe log warning? return error?
		// If we can't seek, the file will have invalid header.
		return 0, 0, 0, fmt.Errorf("seek to start failed: %w", err)
	}

	header := GenerateWavHeader(totalBytes, int(info.SampleRate), int(info.NChannels), int(info.BitsPerSample))
	if _, err := writer.Write(header); err != nil {
		return 0, 0, 0, fmt.Errorf("write real header failed: %w", err)
	}

	// Seek back to end? Not strictly necessary but good practice.
	writer.Seek(0, io.SeekEnd)
	return totalBytes + WavHeaderSize, totalSamples, int(info.SampleRate), nil
}
