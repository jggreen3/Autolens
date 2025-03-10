package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nfnt/resize"
	ort "github.com/yalue/onnxruntime_go"
)

// Configuration constants
const (
	MaxFileSize         = 10 * 1024 * 1024 // 10MB
	ConfidenceThreshold = 0.5
	IoUThreshold        = 0.7
	InputWidth          = 640
	InputHeight         = 640
)

var (
	UseCoreML   = false
	Blank       []float32
	Yolo8Model  ModelSession
	runtimePath = "/tmp/onnxruntime.so"
	ModelPath   = "/tmp/best.onnx"
	initialized = false
	bucketName  = os.Getenv("S3_BUCKET_NAME")
	regionName  = os.Getenv("S3_REGION_NAME")
	runtimeKey  = "onnxruntime.so"
	modelKey    = "best.onnx"
	initMutex   = &sync.Mutex{} // Mutex to prevent race conditions during initialization
)

// ModelSession holds the ONNX runtime session and tensors
type ModelSession struct {
	Session *ort.AdvancedSession
	Input   *ort.Tensor[float32]
	Output  *ort.Tensor[float32]
}

// DetectionResult represents a single detection result
type DetectionResult struct {
	X1         float64 `json:"x1"`
	Y1         float64 `json:"y1"`
	X2         float64 `json:"x2"`
	Y2         float64 `json:"y2"`
	Label      string  `json:"label"`
	Confidence float32 `json:"confidence"`
}

func initializeFiles() error {
	initMutex.Lock()
	defer initMutex.Unlock()

	if initialized {
		return nil
	}

	startTime := time.Now()
	defer func() {
		log.Printf("Initialization took %v", time.Since(startTime))
	}()

	if bucketName == "" || regionName == "" {
		return fmt.Errorf("S3_BUCKET_NAME or S3_REGION_NAME environment variables not set")
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(regionName))
	if err != nil {
		return fmt.Errorf("unable to load SDK config: %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	if !fileExists(runtimePath) {
		log.Printf("Downloading runtime from S3: %s", runtimeKey)
		if err := downloadFromS3(s3Client, bucketName, runtimeKey, runtimePath); err != nil {
			return fmt.Errorf("failed to download runtime: %v", err)
		}
	}

	if !fileExists(ModelPath) {
		log.Printf("Downloading model from S3: %s", modelKey)
		if err := downloadFromS3(s3Client, bucketName, modelKey, ModelPath); err != nil {
			return fmt.Errorf("failed to download model: %v", err)
		}
	}

	initialized = true
	return nil
}

func downloadFromS3(client *s3.Client, bucket, key, filepath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	resp, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object: %v", err)
	}
	defer resp.Body.Close()

	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy object to file: %v", err)
	}

	return nil
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

// Handler function for Vercel serverless deployment
func Handler(w http.ResponseWriter, r *http.Request) {
	startTime := time.Now()
	defer func() {
		log.Printf("Request processed in %v", time.Since(startTime))
	}()

	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Recover from panics
	defer func() {
		if err := recover(); err != nil {
			log.Printf("Panic recovered: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	// Initialize files if needed
	if err := initializeFiles(); err != nil {
		log.Printf("Error initializing files: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Validate request method
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form with size limit
	if err := r.ParseMultipartForm(MaxFileSize); err != nil {
		log.Printf("Form parsing error: %v", err)
		http.Error(w, "Unable to parse form or file too large", http.StatusBadRequest)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("image_file")
	if err != nil {
		log.Printf("File retrieval error: %v", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate file size
	if header.Size > MaxFileSize {
		http.Error(w, "File too large", http.StatusBadRequest)
		return
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if contentType != "image/jpeg" && contentType != "image/png" && contentType != "image/gif" {
		http.Error(w, "Invalid file type. Only JPEG, PNG, and GIF are supported", http.StatusBadRequest)
		return
	}

	log.Println("File retrieved successfully")

	// Detect objects
	boxes, err := DetectObjectsOnImage(file)
	if err != nil {
		log.Printf("Object detection error: %v", err)
		http.Error(w, "Error processing image", http.StatusInternalServerError)
		return
	}

	log.Println("Object detection completed")

	// Marshal response
	buf, err := json.Marshal(&boxes)
	if err != nil {
		log.Printf("JSON encoding error: %v", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	log.Println("Response encoded successfully")

	// Send response
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)

	log.Println("Response sent successfully")
}

func DetectObjectsOnImage(buf io.Reader) ([][]interface{}, error) {
	input, img_width, img_height, err := prepareInput(buf)
	if err != nil {
		return nil, fmt.Errorf("failed to prepare input: %v", err)
	}

	output, err := RunModel(input)
	if err != nil {
		return nil, fmt.Errorf("failed to run model: %v", err)
	}

	data := processOutput(output, img_width, img_height)
	return data, nil
}

func RunModel(input []float32) ([]float32, error) {
	var err error

	if Yolo8Model.Session == nil {
		Yolo8Model, err = InitYolo8Session(input)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize model: %v", err)
		}
	}

	return runInference(Yolo8Model, input)
}

// processOutput converts RAW output from YOLOv8 to an array of detected objects.
// Each object contains the bounding box, object type, and probability.
// Returns array of detected objects in format [[x1,y1,x2,y2,object_type,probability],..]
func processOutput(output []float32, img_width, img_height int64) [][]interface{} {
	boxes := [][]interface{}{}
	for index := 0; index < 8400; index++ {
		class_id, prob := 0, float32(0.0)
		for col := 0; col < 50; col++ {
			if output[8400*(col+4)+index] > prob {
				prob = output[8400*(col+4)+index]
				class_id = col
			}
		}
		if prob < ConfidenceThreshold {
			continue
		}
		label := yolo_classes[class_id]
		xc := output[index]
		yc := output[8400+index]
		w := output[2*8400+index]
		h := output[3*8400+index]
		x1 := (xc - w/2) / float32(InputWidth) * float32(img_width)
		y1 := (yc - h/2) / float32(InputHeight) * float32(img_height)
		x2 := (xc + w/2) / float32(InputWidth) * float32(img_width)
		y2 := (yc + h/2) / float32(InputHeight) * float32(img_height)
		boxes = append(boxes, []interface{}{float64(x1), float64(y1), float64(x2), float64(y2), label, prob})
	}

	// Sort by confidence (highest first)
	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i][5].(float32) > boxes[j][5].(float32)
	})

	// Apply non-maximum suppression
	result := [][]interface{}{}
	for len(boxes) > 0 {
		result = append(result, boxes[0])
		tmp := [][]interface{}{}
		for _, box := range boxes {
			if iou(boxes[0], box) < IoUThreshold {
				tmp = append(tmp, box)
			}
		}
		boxes = tmp
	}
	return result
}

// iou calculates "Intersection-over-union" coefficient for specified two boxes
// Returns Intersection over union ratio as a float number
func iou(box1, box2 []interface{}) float64 {
	return intersection(box1, box2) / union(box1, box2)
}

// union calculates union area of two boxes
// Returns Area of the boxes union as a float number
func union(box1, box2 []interface{}) float64 {
	box1_x1, box1_y1, box1_x2, box1_y2 := box1[0].(float64), box1[1].(float64), box1[2].(float64), box1[3].(float64)
	box2_x1, box2_y1, box2_x2, box2_y2 := box2[0].(float64), box2[1].(float64), box2[2].(float64), box2[3].(float64)
	box1_area := (box1_x2 - box1_x1) * (box1_y2 - box1_y1)
	box2_area := (box2_x2 - box2_x1) * (box2_y2 - box2_y1)
	return box1_area + box2_area - intersection(box1, box2)
}

// intersection calculates intersection area of two boxes
// Returns Area of intersection of the boxes as a float number
func intersection(box1, box2 []interface{}) float64 {
	box1_x1, box1_y1, box1_x2, box1_y2 := box1[0].(float64), box1[1].(float64), box1[2].(float64), box1[3].(float64)
	box2_x1, box2_y1, box2_x2, box2_y2 := box2[0].(float64), box2[1].(float64), box2[2].(float64), box2[3].(float64)
	x1 := math.Max(box1_x1, box2_x1)
	y1 := math.Max(box1_y1, box2_y1)
	x2 := math.Min(box1_x2, box2_x2)
	y2 := math.Min(box1_y2, box2_y2)

	// Handle case where there is no intersection
	if x2 < x1 || y2 < y1 {
		return 0
	}

	return (x2 - x1) * (y2 - y1)
}

// prepareInput converts input image to tensor
// Returns the input tensor, original image width and height
func prepareInput(buf io.Reader) ([]float32, int64, int64, error) {
	img, _, err := image.Decode(buf)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to decode image: %v", err)
	}

	size := img.Bounds().Size()
	img_width, img_height := int64(size.X), int64(size.Y)

	// Resize image to model input dimensions
	img = resize.Resize(uint(InputWidth), uint(InputHeight), img, resize.Lanczos3)

	// Pre-allocate slices for better performance
	pixelCount := InputWidth * InputHeight
	input := make([]float32, 3*pixelCount)

	// Extract RGB channels
	for y := 0; y < InputHeight; y++ {
		for x := 0; x < InputWidth; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			idx := y*InputWidth + x
			input[idx] = float32(r/257) / 255.0
			input[idx+pixelCount] = float32(g/257) / 255.0
			input[idx+2*pixelCount] = float32(b/257) / 255.0
		}
	}

	return input, img_width, img_height, nil
}

func InitYolo8Session(input []float32) (ModelSession, error) {
	ort.SetSharedLibraryPath(getSharedLibPath())
	err := ort.InitializeEnvironment()
	if err != nil {
		return ModelSession{}, fmt.Errorf("failed to initialize ONNX environment: %v", err)
	}

	inputShape := ort.NewShape(1, 3, InputWidth, InputHeight)
	inputTensor, err := ort.NewTensor(inputShape, input)
	if err != nil {
		return ModelSession{}, fmt.Errorf("failed to create input tensor: %v", err)
	}

	outputShape := ort.NewShape(1, 54, 8400)
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return ModelSession{}, fmt.Errorf("failed to create output tensor: %v", err)
	}

	options, e := ort.NewSessionOptions()
	if e != nil {
		return ModelSession{}, fmt.Errorf("failed to create session options: %v", e)
	}
	defer options.Destroy()

	if UseCoreML { // If CoreML is enabled, append the CoreML execution provider
		e = options.AppendExecutionProviderCoreML(0)
		if e != nil {
			return ModelSession{}, fmt.Errorf("failed to append CoreML execution provider: %v", e)
		}
	}

	session, err := ort.NewAdvancedSession(ModelPath,
		[]string{"images"}, []string{"output0"},
		[]ort.ArbitraryTensor{inputTensor}, []ort.ArbitraryTensor{outputTensor}, options)

	if err != nil {
		return ModelSession{}, fmt.Errorf("failed to create ONNX session: %v", err)
	}

	modelSes := ModelSession{
		Session: session,
		Input:   inputTensor,
		Output:  outputTensor,
	}

	return modelSes, nil
}

func getSharedLibPath() string {
	if runtime.GOOS == "windows" {
		if runtime.GOARCH == "amd64" {
			return "./third_party/onnxruntime.dll"
		}
	}
	if runtime.GOOS == "darwin" {
		if runtime.GOARCH == "arm64" {
			return "./third_party/onnxruntime_arm64.dylib"
		}
	}
	if runtime.GOOS == "linux" {
		if runtime.GOARCH == "arm64" {
			return "../third_party/onnxruntime_arm64.so"
		}
		return "/tmp/onnxruntime.so"
	}
	panic("Unable to find a version of the onnxruntime library supporting this system.")
}

var yolo_classes = []string{
	"AIR COMPRESSOR",
	"ALTERNATOR",
	"BATTERY",
	"BRAKE CALIPER",
	"BRAKE PAD",
	"BRAKE ROTOR",
	"CAMSHAFT",
	"CARBERATOR",
	"CLUTCH PLATE",
	"COIL SPRING",
	"CRANKSHAFT",
	"CYLINDER HEAD",
	"DISTRIBUTOR",
	"ENGINE BLOCK",
	"ENGINE VALVE",
	"FUEL INJECTOR",
	"FUSE BOX",
	"GAS CAP",
	"HEADLIGHTS",
	"IDLER ARM",
	"IGNITION COIL",
	"INSTRUMENT CLUSTER",
	"LEAF SPRING",
	"LOWER CONTROL ARM",
	"MUFFLER",
	"OIL FILTER",
	"OIL PAN",
	"OIL PRESSURE SENSOR",
	"OVERFLOW TANK",
	"OXYGEN SENSOR",
	"PISTON",
	"PRESSURE PLATE",
	"RADIATOR",
	"RADIATOR FAN",
	"RADIATOR HOSE",
	"RADIO",
	"RIM",
	"SHIFT KNOB",
	"SIDE MIRROR",
	"SPARK PLUG",
	"SPOILER",
	"STARTER",
	"TAILLIGHTS",
	"THERMOSTAT",
	"TORQUE CONVERTER",
	"TRANSMISSION",
	"VACUUM BRAKE BOOSTER",
	"VALVE LIFTER",
	"WATER PUMP",
	"WINDOW REGULATOR",
}

func runInference(modelSes ModelSession, input []float32) ([]float32, error) {
	inTensor := modelSes.Input.GetData()
	copy(inTensor, input)

	err := modelSes.Session.Run()
	if err != nil {
		return nil, fmt.Errorf("inference failed: %v", err)
	}

	return modelSes.Output.GetData(), nil
}
