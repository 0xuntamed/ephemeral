package config

import "testing"

func TestValidateAPIRequiresKeyWhenProvisioningRealInfrastructure(t *testing.T) {
	cases := []struct {
		name     string
		provider string
		apiKey   string
		wantErr  bool
	}{
		{name: "terraform without key", provider: "terraform", apiKey: "", wantErr: true},
		{name: "terraform with key", provider: "terraform", apiKey: "secret", wantErr: false},
		{name: "demo without key", provider: "demo", apiKey: "", wantErr: false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := Config{Provider: tc.provider, APIKey: tc.apiKey}.ValidateAPI()
			if (err != nil) != tc.wantErr {
				t.Fatalf("ValidateAPI() error = %v, wantErr %v", err, tc.wantErr)
			}
		})
	}
}

func TestLoadDoesNotRequireAPIKeyForStandaloneWorker(t *testing.T) {
	t.Setenv("PROVIDER", "terraform")
	t.Setenv("EPHEMERAL_API_KEY", "")
	if _, err := Load(); err != nil {
		t.Fatalf("Load() = %v; the worker serves no HTTP and must start without an API key", err)
	}
}
