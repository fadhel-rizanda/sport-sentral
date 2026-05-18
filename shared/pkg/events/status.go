package events

const (
	SubjectStatusCreated = "meta.status.created"
	SubjectStatusUpdated = "meta.status.updated"
	SubjectStatusDeleted = "meta.status.deleted"
)

const (
	MetaStreamName       = "META_EVENTS"
	IdentityConsumerName = "identity-service-consumer"
	IdentityDurableName  = "identity-service-durable"
)
