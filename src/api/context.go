package api

type contextKeyCodecT struct{}
type contextKeyDatabaseT struct{}
type contextKeyUserIdT struct{}

var (
	ContextKeyCodec    = contextKeyCodecT{}
	ContextKeyDatabase = contextKeyDatabaseT{}
	ContextKeyUserId   = contextKeyUserIdT{}
)
