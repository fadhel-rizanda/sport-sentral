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

const (
	SubjectCountryCreated = "meta.country.created"
	SubjectCountryUpdated = "meta.country.updated"
	SubjectCountryDeleted = "meta.country.deleted"
)

const (
	SubjectAdministrativeDivisionCreated = "meta.administrative_division.created"
	SubjectAdministrativeDivisionUpdated = "meta.administrative_division.updated"
	SubjectAdministrativeDivisionDeleted = "meta.administrative_division.deleted"
)
