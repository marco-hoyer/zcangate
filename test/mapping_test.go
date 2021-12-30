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
