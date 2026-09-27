package media

const (
	VisibilityPrivate = "private"
	VisibilityLink    = "link"
	VisibilityPublic  = "public"

	ModerationPending      = "pending"
	ModerationApproved     = "approved"
	ModerationManualReview = "manual_review"
	ModerationRejected     = "rejected"
)

// DefaultVisibility is applied to newly completed uploads. Users can switch
// their own ready media to private or unlisted from the member media detail.
const DefaultVisibility = VisibilityPublic

// DefaultModerationStatus means a successfully uploaded image is immediately
// usable and eligible for the public discovery feed. Administrators perform
// moderation after publication through reports and media operations.
const DefaultModerationStatus = ModerationApproved

func ValidVisibility(value string) bool {
	return value == VisibilityPrivate || value == VisibilityLink || value == VisibilityPublic
}

// AllowsPublicDelivery is the common gate for stable external media URLs.
// Normal uploads are available immediately; rejected media is never delivered
// through a public path. Discovery applies the same public/non-rejected rule.
func AllowsPublicDelivery(visibility, moderationStatus string) bool {
	return (visibility == VisibilityPublic || visibility == VisibilityLink) && moderationStatus != ModerationRejected
}

// AllowsSignedDelivery permits a deliberately created, finite signed URL for
// private media, but never lets a rejected asset bypass moderation.
func AllowsSignedDelivery(visibility, moderationStatus string, stable, temporary bool) bool {
	if moderationStatus == ModerationRejected {
		return false
	}
	if AllowsPublicDelivery(visibility, moderationStatus) {
		return true
	}
	return visibility == VisibilityPrivate && temporary && !stable
}
