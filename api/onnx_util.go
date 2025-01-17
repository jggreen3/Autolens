package handler

import (
	"runtime"

	ort "github.com/yalue/onnxruntime_go"
)

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
		return "./third_party/onnxruntime.so"
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
