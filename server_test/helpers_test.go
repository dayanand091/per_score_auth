package identity_server_test

import (
	"testing"
)

func CheckLength(t *testing.T, name string, actualLength int, expectedLength int) {
	if actualLength != expectedLength {
		t.Fatalf("Invalid number of %s, expected %d, got, %d", name, expectedLength, actualLength)
	}
}

func CheckStatus(t *testing.T, status, expectedStatus interface{}) {
	if status != expectedStatus {
		t.Fatalf("Invalid Status, expected %s, got, %s", expectedStatus, status)
	}
}
