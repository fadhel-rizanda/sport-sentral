package events

const (
	AcademyStreamName   = "ACADEMY_EVENTS"
	AcademyConsumerName = "academy-service-consumer"
	AcademyDurableName  = "academy-service-durable"
)

const (
	SubjectAcademyHoldingCreated = "academy.holding.created"
	SubjectAcademyHoldingUpdated = "academy.holding.updated"
	SubjectAcademyHoldingDeleted = "academy.holding.deleted"

	SubjectAcademyBranchCreated = "academy.branch.created"
	SubjectAcademyBranchUpdated = "academy.branch.updated"
	SubjectAcademyBranchDeleted = "academy.branch.deleted"

	SubjectAcademyAdminCreated = "academy.admin.created"
	SubjectAcademyAdminUpdated = "academy.admin.updated"
	SubjectAcademyAdminDeleted = "academy.admin.deleted"

	SubjectEnrollmentCreated = "academy.enrollment.created"
	SubjectEnrollmentUpdated = "academy.enrollment.updated"
	SubjectEnrollmentDeleted = "academy.enrollment.deleted"

	SubjectRosterCreated = "academy.roster.created"
	SubjectRosterUpdated = "academy.roster.updated"
	SubjectRosterDeleted = "academy.roster.deleted"
)
