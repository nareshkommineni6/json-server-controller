package v1

import (
	"testing"
)

func TestValidateName_Valid(t *testing.T) {
	js := &JsonServer{}
	js.Name = "app-myserver"
	js.Spec.JsonConfig = `{"data": []}`

	_, err := js.ValidateCreate()
	if err != nil {
		t.Errorf("expected valid name to pass: %v", err)
	}
}

func TestValidateName_Invalid(t *testing.T) {
	js := &JsonServer{}
	js.Name = "myserver"
	js.Spec.JsonConfig = `{"data": []}`

	_, err := js.ValidateCreate()
	if err == nil {
		t.Error("expected invalid name to fail")
	}
}

func TestValidateJsonConfig_Valid(t *testing.T) {
	js := &JsonServer{}
	js.Name = "app-test"
	js.Spec.JsonConfig = `{"users": [{"id": 1}]}`

	_, err := js.ValidateCreate()
	if err != nil {
		t.Errorf("expected valid json to pass: %v", err)
	}
}

func TestValidateJsonConfig_Invalid(t *testing.T) {
	js := &JsonServer{}
	js.Name = "app-test"
	js.Spec.JsonConfig = `{invalid json}`

	_, err := js.ValidateCreate()
	if err == nil {
		t.Error("expected invalid json to fail")
	}
}
