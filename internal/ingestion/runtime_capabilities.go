package ingestion

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

type RuntimeCapabilities struct {
	FFmpeg              bool
	WhisperCLI          bool
	Python3             bool
	PythonWhisperModule bool
	Tesseract           bool
	ONNXRuntimeLibrary  bool
	ONNXRuntimeLibPath  string
}

func DetectRuntimeCapabilities(clipORTLibraryPath string) RuntimeCapabilities {
	hasPython := hasCommand("python3")
	resolvedORTPath := normalizeONNXRuntimeLibraryPath(clipORTLibraryPath)
	return RuntimeCapabilities{
		FFmpeg:              hasCommand("ffmpeg"),
		WhisperCLI:          hasCommand("whisper"),
		Python3:             hasPython,
		PythonWhisperModule: hasPythonModule("python3", hasPython, "whisper"),
		Tesseract:           hasCommand("tesseract"),
		ONNXRuntimeLibrary:  hasFile(resolvedORTPath),
		ONNXRuntimeLibPath:  resolvedORTPath,
	}
}

func (c RuntimeCapabilities) WhisperFallbackAvailable() bool {
	return c.WhisperCLI || (c.Python3 && c.PythonWhisperModule)
}

func (c RuntimeCapabilities) CLIPVisualAvailable() bool {
	return c.ONNXRuntimeLibrary
}

func (c RuntimeCapabilities) Summary() string {
	return fmt.Sprintf(
		"ffmpeg=%t whisper_cli=%t python3=%t python_whisper_module=%t whisper_fallback=%t tesseract=%t onnxruntime_library=%t onnxruntime_library_path=%s clip_visual=%t",
		c.FFmpeg,
		c.WhisperCLI,
		c.Python3,
		c.PythonWhisperModule,
		c.WhisperFallbackAvailable(),
		c.Tesseract,
		c.ONNXRuntimeLibrary,
		c.ONNXRuntimeLibPath,
		c.CLIPVisualAvailable(),
	)
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func normalizeONNXRuntimeLibraryPath(value string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return "/usr/local/lib/libonnxruntime.so"
}

func hasPythonModule(pythonExec string, hasPython bool, module string) bool {
	if !hasPython {
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	command := exec.CommandContext(ctx, pythonExec, "-c", "import "+module)
	if err := command.Run(); err != nil {
		return false
	}
	return true
}

func hasFile(path string) bool {
	trimmed := strings.TrimSpace(path)
	if trimmed == "" {
		return false
	}
	fileInfo, err := os.Stat(trimmed)
	if err != nil {
		return false
	}
	return !fileInfo.IsDir()
}
