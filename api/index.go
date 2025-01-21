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

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/nfnt/resize"
	ort "github.com/yalue/onnxruntime_go"
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
)

type ModelSession struct {
	Session *ort.AdvancedSession
	Input   *ort.Tensor[float32]
	Output  *ort.Tensor[float32]
}

func initializeFiles() error {
	if initialized {
		return nil
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion(regionName))
	if err != nil {
		return fmt.Errorf("unable to load SDK config, %v", err)
	}

	s3Client := s3.NewFromConfig(cfg)

	if !fileExists(runtimePath) {
		if err := downloadFromS3(s3Client, bucketName, runtimeKey, runtimePath); err != nil {
			return err
		}
	}

	if !fileExists(ModelPath) {
		if err := downloadFromS3(s3Client, bucketName, modelKey, ModelPath); err != nil {
			return err
		}
	}

	initialized = true
	return nil
}

func downloadFromS3(client *s3.Client, bucket, key, filepath string) error {
	resp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to get object, %v", err)
	}
	defer resp.Body.Close()

	outFile, err := os.Create(filepath)
	if err != nil {
		return fmt.Errorf("failed to create file, %v", err)
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to copy object to file, %v", err)
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
	if err := initializeFiles(); err != nil {
		log.Printf("Error initializing files: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	defer func() {
		if err := recover(); err != nil {
			fmt.Printf("Panic recovered: %v\n", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}()

	// Log each step
	fmt.Printf("Request received: %s %s\n", r.Method, r.URL.Path)

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil {
		fmt.Printf("Form parsing error: %v\n", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image_file")
	if err != nil {
		fmt.Printf("File retrieval error: %v\n", err)
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fmt.Println("File retrieved successfully")

	boxes, err := DetectObjectsOnImage(file)
	if err != nil {
		fmt.Printf("Object detection error: %v\n", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	fmt.Println("Object detection completed")

	buf, err := json.Marshal(&boxes)
	if err != nil {
		fmt.Printf("JSON encoding error: %v\n", err)
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	fmt.Println("Response encoded successfully")

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)

	fmt.Println("Response sent successfully")
}
func DetectObjectsOnImage(buf io.Reader) ([][]interface{}, error) {
	input, img_width, img_height := prepare_input(buf)
	output, err := RunModel(input)
	if err != nil {
		return nil, err
	}

	data := process_output(output, img_width, img_height)
	return data, nil
}

func RunModel(input []float32) ([]float32, error) {
	var err error

	if Yolo8Model.Session == nil {
		Yolo8Model, err = InitYolo8Session(input)
		if err != nil {
			return nil, err
		}
	}

	return runInference(Yolo8Model, input)
}

// Function used to convert RAW output from YOLOv8 to an array
// of detected objects. Each object contain the bounding box of
// this object, the type of object and the probability
// Returns array of detected objects in a format [[x1,y1,x2,y2,object_type,probability],..]
func process_output(output []float32, img_width, img_height int64) [][]interface{} {
	boxes := [][]interface{}{}
	for index := 0; index < 8400; index++ {
		class_id, prob := 0, float32(0.0)
		for col := 0; col < 50; col++ {
			if output[8400*(col+4)+index] > prob {
				prob = output[8400*(col+4)+index]
				class_id = col
			}
		}
		if prob < 0.5 {
			continue
		}
		label := yolo_classes[class_id]
		xc := output[index]
		yc := output[8400+index]
		w := output[2*8400+index]
		h := output[3*8400+index]
		x1 := (xc - w/2) / 640 * float32(img_width)
		y1 := (yc - h/2) / 640 * float32(img_height)
		x2 := (xc + w/2) / 640 * float32(img_width)
		y2 := (yc + h/2) / 640 * float32(img_height)
		boxes = append(boxes, []interface{}{float64(x1), float64(y1), float64(x2), float64(y2), label, prob})
	}

	sort.Slice(boxes, func(i, j int) bool {
		return boxes[i][5].(float32) < boxes[j][5].(float32)
	})
	result := [][]interface{}{}
	for len(boxes) > 0 {
		result = append(result, boxes[0])
		tmp := [][]interface{}{}
		for _, box := range boxes {
			if iou(boxes[0], box) < 0.7 {
				tmp = append(tmp, box)
			}
		}
		boxes = tmp
	}
	return result
}

// Function calculates "Intersection-over-union" coefficient for specified two boxes
// https://pyimagesearch.com/2016/11/07/intersection-over-union-iou-for-object-detection/.
// Returns Intersection over union ratio as a float number
func iou(box1, box2 []interface{}) float64 {
	return intersection(box1, box2) / union(box1, box2)
}

// Function calculates union area of two boxes
// Returns Area of the boxes union as a float number
func union(box1, box2 []interface{}) float64 {
	box1_x1, box1_y1, box1_x2, box1_y2 := box1[0].(float64), box1[1].(float64), box1[2].(float64), box1[3].(float64)
	box2_x1, box2_y1, box2_x2, box2_y2 := box2[0].(float64), box2[1].(float64), box2[2].(float64), box2[3].(float64)
	box1_area := (box1_x2 - box1_x1) * (box1_y2 - box1_y1)
	box2_area := (box2_x2 - box2_x1) * (box2_y2 - box2_y1)
	return box1_area + box2_area - intersection(box1, box2)
}

// Function calculates intersection area of two boxes
// Returns Area of intersection of the boxes as a float number
func intersection(box1, box2 []interface{}) float64 {
	box1_x1, box1_y1, box1_x2, box1_y2 := box1[0].(float64), box1[1].(float64), box1[2].(float64), box1[3].(float64)
	box2_x1, box2_y1, box2_x2, box2_y2 := box2[0].(float64), box2[1].(float64), box2[2].(float64), box2[3].(float64)
	x1 := math.Max(box1_x1, box2_x1)
	y1 := math.Max(box1_y1, box2_y1)
	x2 := math.Min(box1_x2, box2_x2)
	y2 := math.Min(box1_y2, box2_y2)
	return (x2 - x1) * (y2 - y1)
}

// Function used to convert input image to tensor,
// required as an input to YOLOv8 object detection
// network.
// Returns the input tensor, original image width and height
func prepare_input(buf io.Reader) ([]float32, int64, int64) {
	img, _, _ := image.Decode(buf)
	size := img.Bounds().Size()
	img_width, img_height := int64(size.X), int64(size.Y)
	img = resize.Resize(640, 640, img, resize.Lanczos3)
	red := []float32{}
	green := []float32{}
	blue := []float32{}
	for y := 0; y < 640; y++ {
		for x := 0; x < 640; x++ {
			r, g, b, _ := img.At(x, y).RGBA()
			red = append(red, float32(r/257)/255.0)
			green = append(green, float32(g/257)/255.0)
			blue = append(blue, float32(b/257)/255.0)
		}
	}
	input := append(red, green...)
	input = append(input, blue...)
	return input, img_width, img_height
}

func InitYolo8Session(input []float32) (ModelSession, error) {
	ort.SetSharedLibraryPath(getSharedLibPath())
	err := ort.InitializeEnvironment()
	if err != nil {
		return ModelSession{}, err
	}

	inputShape := ort.NewShape(1, 3, 640, 640)
	inputTensor, err := ort.NewTensor(inputShape, input)
	if err != nil {
		return ModelSession{}, err
	}

	outputShape := ort.NewShape(1, 54, 8400)
	outputTensor, err := ort.NewEmptyTensor[float32](outputShape)
	if err != nil {
		return ModelSession{}, err
	}

	options, e := ort.NewSessionOptions()
	if e != nil {
		return ModelSession{}, err
	}

	if UseCoreML { // If CoreML is enabled, append the CoreML execution provider
		e = options.AppendExecutionProviderCoreML(0)
		if e != nil {
			options.Destroy()
			return ModelSession{}, err
		}
		defer options.Destroy()
	}

	session, err := ort.NewAdvancedSession(ModelPath,
		[]string{"images"}, []string{"output0"},
		[]ort.ArbitraryTensor{inputTensor}, []ort.ArbitraryTensor{outputTensor}, options)

	if err != nil {
		return ModelSession{}, err
	}

	modelSes := ModelSession{
		Session: session,
		Input:   inputTensor,
		Output:  outputTensor,
	}

	return modelSes, err
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
		return nil, err
	}
	return modelSes.Output.GetData(), nil
}
