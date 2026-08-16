package events

const (
	AttachmentStreamName   = "ATTACHMENT_EVENTS"
	AttachmentConsumerName = "attachment-service-consumer"
	AttachmentDurableName  = "attachment-service-durable"
)

const (
	SubjectAttachmentCreated = "attachment.created"
	SubjectAttachmentUpdated = "attachment.updated"
	SubjectAttachmentDeleted = "attachment.deleted"
)
