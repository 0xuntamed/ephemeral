package environment

import "testing"

func TestLifecycleRejectsReadyWithoutHealthCheck(t *testing.T) {
	if CanTransition(StatusProvisioning, StatusReady) {
		t.Fatal("PROVISIONING -> READY skipped the health-check gate")
	}
	if !CanTransition(StatusHealthCheck, StatusReady) {
		t.Fatal("HEALTH_CHECK -> READY should be allowed")
	}
}
