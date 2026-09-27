package media

import (
	"context"
	"fmt"
	"time"

	"github.com/goravel/framework/contracts/queue"

	"goravel/app/facades"
)

const RecoverUploadJobSignature = "fastimg.media.recover-upload"

type RecoverUploadJob struct{}

func (*RecoverUploadJob) Signature() string { return RecoverUploadJobSignature }

func (*RecoverUploadJob) Handle(args ...any) error {
	if len(args) < 2 {
		return fmt.Errorf("user id and upload session id are required")
	}
	userID, userOK := args[0].(uint)
	sessionID, sessionOK := args[1].(uint)
	if !userOK || !sessionOK || userID == 0 || sessionID == 0 {
		return fmt.Errorf("invalid upload recovery identifiers")
	}
	_, err := NewDatabaseUploadService().Retry(context.Background(), userID, sessionID)
	return err
}

func (*RecoverUploadJob) ShouldRetry(_ error, attempt int) (bool, time.Duration) {
	if attempt >= 3 {
		return false, 0
	}
	return true, time.Duration(attempt+1) * time.Minute
}

func DispatchUploadRecovery(userID, sessionID uint) error {
	if userID == 0 || sessionID == 0 {
		return ErrInvalidUploadInput
	}
	return facades.Queue().Job(&RecoverUploadJob{}, []queue.Arg{{Value: userID, Type: "uint"}, {Value: sessionID, Type: "uint"}}).OnQueue("media").Dispatch()
}
