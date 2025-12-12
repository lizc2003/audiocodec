package audiocodec_test

import (
	"crypto/md5"
	"encoding/hex"
	"os"
	"testing"

	"github.com/lizc2003/audiocodec"
)

func TestFlac2Wav(t *testing.T) {
	t.Run("flac to wav", func(t *testing.T) {
		inFile, err := os.Open("samples/sample.flac")
		if err != nil {
			t.Fatalf("open file failed: %v", err)
		}
		defer inFile.Close()

		writer := &audiocodec.WriterSeeker{}
		totalBytes, totalSamples, sampleRate, err := audiocodec.FlacToWav(inFile, writer)
		if err != nil {
			t.Fatalf("flac to wav failed: %v", err)
		}
		if totalBytes != 2465836 {
			t.Errorf("expected decoded bytes 2465836, got %d", totalBytes)
		}
		if totalSamples != 616448 {
			t.Errorf("expected samples 616448, got %d", totalSamples)
		}
		if sampleRate != 44100 {
			t.Errorf("expected sample rate 44100, got %d", sampleRate)
		}

		bytes := writer.Bytes()

		md5sum := md5.Sum(bytes)
		if hex.EncodeToString(md5sum[:]) != "24d62c157971605e42defb9954d65e33" {
			t.Error("wav md5sum check wrong")
		}

		md5sum = md5.Sum(bytes[audiocodec.WavHeaderSize:])
		if hex.EncodeToString(md5sum[:]) != "372fd5b0a07a0a78b92311ccdca4cc81" {
			t.Error("pcm md5sum check wrong")
		}
	})
}
