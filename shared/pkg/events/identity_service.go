package events

const (
	IdentityStreamName = "IDENTITY_EVENTS"
	MetaConsumerName   = "meta-service-consumer"
	MetaDurableName    = "meta-service-durable"
)

const (
	SubjectUserCreated = "identity.user.created"
	SubjectUserUpdated = "identity.user.updated"
	SubjectUserDeleted = "identity.user.deleted"
)

const (
	SubjectRoleCreated = "meta.status.created"
	SubjectRoleUpdated = "meta.status.updated"
	SubjectRoleDeleted = "meta.status.deleted"
)

const (
	SubjectPermissionCreated = "meta.status.created"
	SubjectPermissionUpdated = "meta.status.updated"
	SubjectPermissionDeleted = "meta.status.deleted"
)
