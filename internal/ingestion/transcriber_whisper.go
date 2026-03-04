package ingestion

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/andreas-lindfalk/videra/internal/storage"
)

type WhisperCLITranscriber struct {
	model    string
	language string
}

func NewWhisperCLITranscriber() *WhisperCLITranscriber {
	model := strings.TrimSpace(os.Getenv("VIDERA_WHISPER_MODEL"))
	if model == "" {
		model = "tiny"
	}
	language := strings.TrimSpace(os.Getenv("VIDERA_WHISPER_LANGUAGE"))
	return &WhisperCLITranscriber{model: model, language: language}
}

func (t *WhisperCLITranscriber) Transcribe(ctx context.Context, audioPath string) ([]storage.Segment, error) {
	if strings.TrimSpace(audioPath) == "" {
		return nil, fmt.Errorf("audio path is required")
	}
	if _, err := os.Stat(audioPath); err != nil {
		return nil, fmt.Errorf("audio path not found: %w", err)
	}

	outputDir, err := os.MkdirTemp("", "videra-whisper-*")
	if err != nil {
		return nil, fmt.Errorf("create whisper output dir: %w", err)
	}
	defer os.RemoveAll(outputDir)

	if err := runWhisperCLI(ctx, audioPath, outputDir, t.model, t.language); err != nil {
		return nil, err
	}

	srtPath := filepath.Join(outputDir, strings.TrimSuffix(filepath.Base(audioPath), filepath.Ext(audioPath))+".srt")
	payload, err := os.ReadFile(srtPath)
	if err != nil {
		return nil, fmt.Errorf("read whisper transcript output: %w", err)
	}

	segments, err := parseTimedCaptionSegments(string(payload), 5)
	if err != nil {
		return nil, err
	}
	for index := range segments {
		segments[index].Type = storage.SegmentTypeAudio
	}
	return segments, nil
}

func runWhisperCLI(ctx context.Context, audioPath, outputDir, model, language string) error {
	args := []string{
		audioPath,
		"--task", "transcribe",
		"--output_format", "srt",
		"--output_dir", outputDir,
		"--model", model,
		"--fp16", "False",
	}
	if language != "" {
		args = append(args, "--language", language)
	}

	var lastErr error
	if _, err := exec.LookPath("whisper"); err == nil {
		cmd := exec.CommandContext(ctx, "whisper", args...)
		if output, runErr := cmd.CombinedOutput(); runErr != nil {
			lastErr = fmt.Errorf("whisper transcription failed: %w: %s", runErr, strings.TrimSpace(string(output)))
		} else {
			return nil
		}
	}

	if _, err := exec.LookPath("python3"); err == nil {
		pythonArgs := append([]string{"-m", "whisper"}, args...)
		cmd := exec.CommandContext(ctx, "python3", pythonArgs...)
		if output, runErr := cmd.CombinedOutput(); runErr != nil {
			lastErr = fmt.Errorf("python whisper transcription failed: %w: %s", runErr, strings.TrimSpace(string(output)))
		} else {
			return nil
		}
	}

	if lastErr != nil {
		return lastErr
	}
	return fmt.Errorf("whisper CLI not found; install `whisper` (or `python3 -m whisper`) or provide a sidecar transcript")
}

var _ Transcriber = (*WhisperCLITranscriber)(nil)
