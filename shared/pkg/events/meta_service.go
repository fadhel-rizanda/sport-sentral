package events

const (
	MetaStreamName       = "META_EVENTS"
	IdentityConsumerName = "identity-service-consumer"
	IdentityDurableName  = "identity-service-durable"
)

const (
	SubjectStatusCreated = "meta.status.created"
	SubjectStatusUpdated = "meta.status.updated"
	SubjectStatusDeleted = "meta.status.deleted"
)

const (
	SubjectTagCreated = "meta.tag.created"
	SubjectTagUpdated = "meta.tag.updated"
	SubjectTagDeleted = "meta.tag.deleted"
)
