package ingestion

import (
	"context"
	"fmt"
	"os/exec"
	"time"
)

type RuntimeCapabilities struct {
	FFmpeg              bool
	WhisperCLI          bool
	Python3             bool
	PythonWhisperModule bool
	Tesseract           bool
}

func DetectRuntimeCapabilities() RuntimeCapabilities {
	python3 := hasCommand("python3")
	return RuntimeCapabilities{
		FFmpeg:              hasCommand("ffmpeg"),
		WhisperCLI:          hasCommand("whisper"),
		Python3:             python3,
		PythonWhisperModule: hasPythonWhisperModule(python3),
		Tesseract:           hasCommand("tesseract"),
	}
}

func (c RuntimeCapabilities) WhisperFallbackAvailable() bool {
	return c.WhisperCLI || (c.Python3 && c.PythonWhisperModule)
}

func (c RuntimeCapabilities) Summary() string {
	return fmt.Sprintf(
		"ffmpeg=%t whisper_cli=%t python3=%t python_whisper_module=%t whisper_fallback=%t tesseract=%t",
		c.FFmpeg,
		c.WhisperCLI,
		c.Python3,
		c.PythonWhisperModule,
		c.WhisperFallbackAvailable(),
		c.Tesseract,
	)
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func hasPythonWhisperModule(hasPython3 bool) bool {
	if !hasPython3 {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, "python3", "-c", "import whisper")
	if err := command.Run(); err != nil {
		return false
	}
	return true
}
