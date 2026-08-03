package test

import (
	"fmt"
	"github.com/marco-hoyer/zcangate/can"
	"testing"
)

func TestTransformTemperatureForZero(t *testing.T) {

	result := can.TransformTemperature("FFFF")
	fmt.Println(result)
	if result != 0.0 {
		t.Errorf("TransformTemperature('FFFF') = %.2f; want 0.0", result)
	}
}

func TestTransformTemperatureForNegativeValue(t *testing.T) {

	result := can.TransformTemperature("FBFF")
	fmt.Println(result)
	if result != -0.4 {
		t.Errorf("TransformTemperature('FBFF') = %.2f; want -0.4", result)
	}
}

func TestTransformTemperatureForPositiveValue(t *testing.T) {

	result := can.TransformTemperature("0201")
	fmt.Println(result)
	if result != 25.7 {
		t.Errorf("TransformTemperature('0201') = %.2f; want 25.7", result)
	}
}

func TestToPdo(t *testing.T) {
	input := uint64(0x00454041)
	result := can.ToPdo(input, 1)
	if result != 277 {
		t.Errorf("00454041 should be transformed to 277 but was %d", result)
	}

	input = uint64(0x00458041)
	result = can.ToPdo(input, 1)
	if result != 278 {
		t.Errorf("00454041 should be transformed to 278 but was %d", result)
	}
}

func TestFromPdo(t *testing.T) {
	input := uint64(277)
	result := can.FromPdo(input, 1)
	if result != 0x00454041 {
		t.Errorf("277 should be transformed to 0x00454041 but was %#x", result)
	}
}

func TestToMeasurementDecodesVentilationControlMode(t *testing.T) {
	autoFrame := can.BusFrame{Pdu: 72, Data: "00"}
	autoResult := can.ToMeasurement(autoFrame)
	if autoResult.Name != "ventilation_control_mode" {
		t.Errorf("expected name 'ventilation_control_mode', got '%s'", autoResult.Name)
	}
	if autoResult.Value != 0.0 {
		t.Errorf("expected value 0.0 for auto mode, got %.2f", autoResult.Value)
	}

	manualFrame := can.BusFrame{Pdu: 72, Data: "01"}
	manualResult := can.ToMeasurement(manualFrame)
	if manualResult.Value != 1.0 {
		t.Errorf("expected value 1.0 for manual mode, got %.2f", manualResult.Value)
	}
}
