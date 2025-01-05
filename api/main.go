package main

import (
	"encoding/json"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/rs/cors"
)

// Main function that defines
// a web service endpoints a starts
// the web service
func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("/detect", detect)

	// Setup CORS middleware
	handler := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type"},
		AllowCredentials: true,
	}).Handler(mux)

	log.Fatal(http.ListenAndServe(":8080", handler))
}

// Site main page handler function.
// Returns Content of index.html file
func index(w http.ResponseWriter, _ *http.Request) {
	file, _ := os.Open("index.html")
	buf, _ := io.ReadAll(file)
	w.Write(buf)
}

// Handler of /detect POST endpoint
// Receives uploaded file with a name "image_file", passes it
// through YOLOv8 object detection network and returns and array
// of bounding boxes.
// Returns a JSON array of objects bounding boxes in format [[x1,y1,x2,y2,object_type,probability],..]
func detect(w http.ResponseWriter, r *http.Request) {
	// Parse with a reasonable max memory limit
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image_file")
	if err != nil {
		http.Error(w, "Error retrieving file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	boxes, err := detect_objects_on_image(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf, err := json.Marshal(&boxes)
	if err != nil {
		http.Error(w, "Error encoding response", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(buf)
}

// Function receives an image,
// passes it through YOLOv8 neural network
// and returns an array of detected objects
// and their bounding boxes
// Returns Array of bounding boxes in format [[x1,y1,x2,y2,object_type,probability],..]
func detect_objects_on_image(buf io.Reader) ([][]interface{}, error) {
	input, img_width, img_height := prepare_input(buf)
	output, err := run_model(input)
	if err != nil {
		return nil, err
	}

	data := process_output(output, img_width, img_height)

	return data, nil
}

// Function used to pass provided input tensor to
// YOLOv8 neural network and return result
// Returns raw output of YOLOv8 network as a single dimension
// array
func run_model(input []float32) ([]float32, error) {

	var err error

	if Yolo8Model.Session == nil {
		Yolo8Model, err = InitYolo8Session(input)
		if err != nil {
			return nil, err
		}
	}

	return runInference(Yolo8Model, input)

}
