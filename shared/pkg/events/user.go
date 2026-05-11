package events

const (
	SubjectUserCreated = "identity.user.created"
	SubjectUserUpdated = "identity.user.updated"
	SubjectUserDeleted = "identity.user.deleted"
)

const (
	IdentityStreamName = "IDENTITY_EVENTS"
	MetaConsumerName   = "meta-service-consumer"
	MetaDurableName    = "meta-service-durable"
)
