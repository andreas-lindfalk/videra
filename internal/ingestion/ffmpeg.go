package ingestion

import (
	"context"
	"fmt"
	"os/exec"
)

type FFmpegRunner interface {
	ExtractAudio(ctx context.Context, videoPath, outputPath string) error
	ExtractKeyframes(ctx context.Context, videoPath, outputDir string, intervalSec int) error
}

type ExecFFmpeg struct{}

func (ExecFFmpeg) ExtractAudio(ctx context.Context, videoPath, outputPath string) error {
	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", videoPath,
		"-vn",
		"-acodec", "mp3",
		outputPath,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg extract audio failed: %w: %s", err, string(output))
	}
	return nil
}

func (ExecFFmpeg) ExtractKeyframes(ctx context.Context, videoPath, outputDir string, intervalSec int) error {
	if intervalSec <= 0 {
		intervalSec = 5
	}

	outputPattern := fmt.Sprintf("%s/frame-%%05d.jpg", outputDir)
	filter := fmt.Sprintf("fps=1/%d", intervalSec)

	cmd := exec.CommandContext(
		ctx,
		"ffmpeg",
		"-y",
		"-i", videoPath,
		"-vf", filter,
		outputPattern,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("ffmpeg extract keyframes failed: %w: %s", err, string(output))
	}

	return nil
}

var _ FFmpegRunner = (*ExecFFmpeg)(nil)
