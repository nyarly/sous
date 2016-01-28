package yaml

import (
	"reflect"
	"testing"
)

func TestInjectTemplatePipelineMap(t *testing.T) {
	template := map[string]string{
		"k1":        "v1",
		"k2":        "{{.Value2}}",
		"{{.Key3}}": "{{.Value3}}",
	}
	pipeline := map[string]string{
		"Value2": "v2",
		"Value3": "v3",
		"Key3":   "k3",
	}
	expectedResult := map[string]string{
		"k1": "v1",
		"k2": "v2",
		"k3": "v3",
	}

	var result map[string]string

	if err := InjectTemplatePipeline(template, &result, pipeline); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Got result: % +v \n\nWant:\n\n% +v", result, expectedResult)
	}
}

type TestStruct struct {
	K1, K2 string
}

func TestInjectTemplatePipelineStruct(t *testing.T) {
	template := TestStruct{"v1", "{{.Value2}}"}
	pipeline := map[string]string{"Value2": "v2"}
	expectedResult := TestStruct{"v1", "v2"}
	var result TestStruct

	if err := InjectTemplatePipeline(template, &result, pipeline); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(result, expectedResult) {
		t.Fatalf("Got result: % +v \n\nWant:\n\n% +v", result, expectedResult)
	}
}
