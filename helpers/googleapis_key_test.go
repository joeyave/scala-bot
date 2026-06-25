package helpers

import "testing"

func TestGoogleAPIsClientEmailFromJSON(t *testing.T) {
	email, err := GoogleAPIsClientEmailFromJSON(`{"client_email":"bot@example.iam.gserviceaccount.com"}`)
	if err != nil {
		t.Fatalf("GoogleAPIsClientEmailFromJSON returned error: %v", err)
	}

	if email != "bot@example.iam.gserviceaccount.com" {
		t.Fatalf("expected client email, got %q", email)
	}
}

func TestGoogleAPIsClientEmailFromJSONReturnsErrorForMissingEmail(t *testing.T) {
	if _, err := GoogleAPIsClientEmailFromJSON(`{"type":"service_account"}`); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGoogleAPIsClientEmailFromEnv(t *testing.T) {
	t.Setenv(googleAPIsKeyEnv, `{"client_email":"bot@example.iam.gserviceaccount.com"}`)

	email, err := GoogleAPIsClientEmailFromEnv()
	if err != nil {
		t.Fatalf("GoogleAPIsClientEmailFromEnv returned error: %v", err)
	}

	if email != "bot@example.iam.gserviceaccount.com" {
		t.Fatalf("expected client email, got %q", email)
	}
}
