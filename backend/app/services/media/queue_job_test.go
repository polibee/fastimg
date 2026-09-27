package media

import (
	"testing"
	"time"
)

func TestRecoverUploadJobContract(t *testing.T) {
	job := &RecoverUploadJob{}
	if job.Signature() != RecoverUploadJobSignature {
		t.Fatalf("signature = %q", job.Signature())
	}
	if err := job.Handle(1); err == nil {
		t.Fatal("expected missing session id error")
	}
	if retry, delay := job.ShouldRetry(nil, 3); retry || delay != 0 {
		t.Fatalf("final retry = %v, %s", retry, delay)
	}
	if retry, delay := job.ShouldRetry(nil, 1); !retry || delay != 2*time.Minute {
		t.Fatalf("second attempt retry = %v, %s", retry, delay)
	}
}

func TestDispatchUploadRecoveryRejectsEmptyIdentifiers(t *testing.T) {
	if err := DispatchUploadRecovery(0, 1); err == nil {
		t.Fatal("expected empty user id error")
	}
	if err := DispatchUploadRecovery(1, 0); err == nil {
		t.Fatal("expected empty session id error")
	}
}
