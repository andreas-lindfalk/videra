package ingestion

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRuntimeCapabilitiesWhisperFallbackAvailable(t *testing.T) {
	testCases := []struct {
		name         string
		capabilities RuntimeCapabilities
		expected     bool
	}{
		{
			name: "whisper cli available",
			capabilities: RuntimeCapabilities{
				WhisperCLI: true,
			},
			expected: true,
		},
		{
			name: "python module path available",
			capabilities: RuntimeCapabilities{
				Python3:             true,
				PythonWhisperModule: true,
			},
			expected: true,
		},
		{
			name: "python without module not available",
			capabilities: RuntimeCapabilities{
				Python3:             true,
				PythonWhisperModule: false,
			},
			expected: false,
		},
		{
			name:         "no whisper runtime",
			capabilities: RuntimeCapabilities{},
			expected:     false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			require.Equal(t, testCase.expected, testCase.capabilities.WhisperFallbackAvailable())
		})
	}
}

func TestRuntimeCapabilitiesSummaryIncludesAllFlags(t *testing.T) {
	capabilities := RuntimeCapabilities{
		FFmpeg:              true,
		WhisperCLI:          false,
		Python3:             true,
		PythonWhisperModule: true,
		Tesseract:           false,
		ONNXRuntimeLibrary:  true,
		ONNXRuntimeLibPath:  "/usr/local/lib/libonnxruntime.so",
	}

	summary := capabilities.Summary()
	require.Contains(t, summary, "ffmpeg=true")
	require.Contains(t, summary, "whisper_cli=false")
	require.Contains(t, summary, "python3=true")
	require.Contains(t, summary, "python_whisper_module=true")
	require.Contains(t, summary, "whisper_fallback=true")
	require.Contains(t, summary, "tesseract=false")
	require.Contains(t, summary, "onnxruntime_library=true")
	require.Contains(t, summary, "onnxruntime_library_path=/usr/local/lib/libonnxruntime.so")
	require.Contains(t, summary, "clip_visual=true")
}

func TestRuntimeCapabilitiesCLIPVisualAvailableDependsOnONNXLibrary(t *testing.T) {
	require.False(t, RuntimeCapabilities{ONNXRuntimeLibrary: false}.CLIPVisualAvailable())
	require.True(t, RuntimeCapabilities{ONNXRuntimeLibrary: true}.CLIPVisualAvailable())
}
