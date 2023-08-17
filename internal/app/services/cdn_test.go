package services_test

import (
	"testing"

	"github.com/andrei0427/go-changediff/internal/app/services"
)

func TestCDNService_GetFileExtForFile(t *testing.T) {
	cdn := services.NewCDNService()
	result := cdn.GetFileExt("file.jpg")
	expected := "jpg"

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestCDNService_GetFileExtForFileWithNoName(t *testing.T) {
	cdn := services.NewCDNService()
	result := cdn.GetFileExt(".jpg")
	expected := "jpg"

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestCDNService_GetFileExtForFileWithNoExt(t *testing.T) {
	cdn := services.NewCDNService()
	result := cdn.GetFileExt("lol")
	expected := ""

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}
