//go:build cgo

package ingestion

import (
	"context"
	"fmt"
	"image"
	"image/draw"
	"math"
	"os"
	"strings"
	"sync"

	ort "github.com/yalue/onnxruntime_go"
	xdraw "golang.org/x/image/draw"

	_ "image/jpeg"
	_ "image/png"
)

var clipMean = [3]float32{0.48145466, 0.4578275, 0.40821073}
var clipStd = [3]float32{0.26862954, 0.26130258, 0.27577711}

var ortEnvironmentMu sync.Mutex
var ortEnvironmentInitialized bool
var ortEnvironmentLibraryPath string

type NativeCLIPRunner struct {
	mu         sync.Mutex
	modelPath  string
	inputName  string
	outputName string
	session    *ort.DynamicAdvancedSession
}

func NewNativeCLIPRunner() *NativeCLIPRunner {
	return &NativeCLIPRunner{}
}

func (r *NativeCLIPRunner) Prepare(modelPath, ortLibraryPath string) error {
	trimmedModelPath := strings.TrimSpace(modelPath)
	if trimmedModelPath == "" {
		return fmt.Errorf("clip model path is required")
	}
	if _, err := os.Stat(trimmedModelPath); err != nil {
		return fmt.Errorf("clip model path is not readable: %w", err)
	}

	trimmedORTLibraryPath := strings.TrimSpace(ortLibraryPath)
	if trimmedORTLibraryPath == "" {
		return fmt.Errorf("clip onnxruntime shared library path is required")
	}
	if _, err := os.Stat(trimmedORTLibraryPath); err != nil {
		return fmt.Errorf("clip onnxruntime shared library is not readable: %w", err)
	}

	if err := initializeONNXRuntimeEnvironment(trimmedORTLibraryPath); err != nil {
		return err
	}

	inputs, outputs, err := ort.GetInputOutputInfo(trimmedModelPath)
	if err != nil {
		return fmt.Errorf("inspect clip model input/output metadata: %w", err)
	}
	if len(inputs) == 0 {
		return fmt.Errorf("clip model has no inputs")
	}
	if len(outputs) == 0 {
		return fmt.Errorf("clip model has no outputs")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.session != nil {
		if r.modelPath != trimmedModelPath {
			return fmt.Errorf("clip runner already initialized for model %s", r.modelPath)
		}
		return nil
	}

	session, err := ort.NewDynamicAdvancedSession(trimmedModelPath, []string{inputs[0].Name}, []string{outputs[0].Name}, nil)
	if err != nil {
		return fmt.Errorf("create clip inference session: %w", err)
	}

	r.modelPath = trimmedModelPath
	r.inputName = inputs[0].Name
	r.outputName = outputs[0].Name
	r.session = session
	return nil
}

func initializeONNXRuntimeEnvironment(libraryPath string) error {
	ortEnvironmentMu.Lock()
	defer ortEnvironmentMu.Unlock()

	if ortEnvironmentInitialized {
		if ortEnvironmentLibraryPath != libraryPath {
			return fmt.Errorf("onnxruntime already initialized with %s, cannot switch to %s", ortEnvironmentLibraryPath, libraryPath)
		}
		return nil
	}

	ort.SetSharedLibraryPath(libraryPath)
	if err := ort.InitializeEnvironment(); err != nil {
		return fmt.Errorf("initialize onnxruntime environment: %w", err)
	}

	ortEnvironmentInitialized = true
	ortEnvironmentLibraryPath = libraryPath
	return nil
}

func (r *NativeCLIPRunner) EmbedFrame(_ context.Context, modelPath, framePath string, inputSize int) ([]float32, error) {
	trimmedFramePath := strings.TrimSpace(framePath)
	if trimmedFramePath == "" {
		return nil, fmt.Errorf("frame path is required")
	}
	if inputSize <= 0 {
		return nil, fmt.Errorf("clip input size must be > 0")
	}

	r.mu.Lock()
	session := r.session
	initializedModelPath := r.modelPath
	r.mu.Unlock()

	if session == nil {
		return nil, fmt.Errorf("clip runner is not initialized")
	}
	if strings.TrimSpace(modelPath) != initializedModelPath {
		return nil, fmt.Errorf("clip runner initialized for model %s but got %s", initializedModelPath, strings.TrimSpace(modelPath))
	}

	inputData, err := preprocessCLIPFrame(trimmedFramePath, inputSize)
	if err != nil {
		return nil, err
	}

	inputTensor, err := ort.NewTensor(ort.NewShape(1, 3, int64(inputSize), int64(inputSize)), inputData)
	if err != nil {
		return nil, fmt.Errorf("create clip input tensor: %w", err)
	}
	defer inputTensor.Destroy()

	outputs := []ort.Value{nil}

	r.mu.Lock()
	err = r.session.Run([]ort.Value{inputTensor}, outputs)
	r.mu.Unlock()
	if err != nil {
		return nil, fmt.Errorf("run clip inference: %w", err)
	}
	if len(outputs) == 0 || outputs[0] == nil {
		return nil, fmt.Errorf("clip inference returned no outputs")
	}
	defer outputs[0].Destroy()

	outputTensor, ok := outputs[0].(*ort.Tensor[float32])
	if !ok {
		return nil, fmt.Errorf("clip output tensor type is not float32")
	}

	raw := outputTensor.GetData()
	if len(raw) == 0 {
		return nil, fmt.Errorf("clip inference returned empty embedding")
	}
	embedding := append([]float32(nil), raw...)
	normalizeCLIPEmbedding(embedding)
	return embedding, nil
}

func preprocessCLIPFrame(framePath string, inputSize int) ([]float32, error) {
	file, err := os.Open(framePath)
	if err != nil {
		return nil, fmt.Errorf("open frame image: %w", err)
	}
	defer file.Close()

	decodedImage, _, err := image.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("decode frame image: %w", err)
	}

	resizedImage := image.NewRGBA(image.Rect(0, 0, inputSize, inputSize))
	xdraw.CatmullRom.Scale(resizedImage, resizedImage.Bounds(), decodedImage, decodedImage.Bounds(), draw.Over, nil)

	pixelCount := inputSize * inputSize
	channels := 3
	inputData := make([]float32, channels*pixelCount)

	for y := 0; y < inputSize; y++ {
		for x := 0; x < inputSize; x++ {
			red, green, blue, _ := resizedImage.At(x, y).RGBA()
			redValue := float32(red) / 65535.0
			greenValue := float32(green) / 65535.0
			blueValue := float32(blue) / 65535.0

			pixelIndex := y*inputSize + x
			inputData[pixelIndex] = (redValue - clipMean[0]) / clipStd[0]
			inputData[pixelCount+pixelIndex] = (greenValue - clipMean[1]) / clipStd[1]
			inputData[(2*pixelCount)+pixelIndex] = (blueValue - clipMean[2]) / clipStd[2]
		}
	}

	return inputData, nil
}

func normalizeCLIPEmbedding(vector []float32) {
	var squaredSum float64
	for _, value := range vector {
		squaredSum += float64(value * value)
	}
	if squaredSum <= 0 {
		return
	}

	norm := float32(math.Sqrt(squaredSum))
	for index := range vector {
		vector[index] = vector[index] / norm
	}
}
