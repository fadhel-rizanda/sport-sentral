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
	SubjectRoleCreated = "identity.role.created"
	SubjectRoleUpdated = "identity.role.updated"
	SubjectRoleDeleted = "identity.role.deleted"
)

const (
	SubjectPermissionCreated = "identity.permission.created"
	SubjectPermissionUpdated = "identity.permission.updated"
	SubjectPermissionDeleted = "identity.permission.deleted"
)
