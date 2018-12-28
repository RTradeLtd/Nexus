package delegator

// contextKey is used to denote context keys
type contextKey string

func (c contextKey) String() string { return string(c) }

const (
	keyNetwork contextKey = "network_id"
	keyFeature contextKey = "feature"
)
