package entity

type SeverityCount struct {
	Severity string `json:"severity"`
	Count    int64  `json:"count"`
}

type ActionCount struct {
	Action string `json:"action"`
	Count  int64  `json:"count"`
}

type ServiceCount struct {
	ServiceName string `json:"service_name"`
	Count       int64  `json:"count"`
}

type LogStats struct {
	TotalAuditLogs    int64           `json:"total_audit_logs"`
	TotalActivityLogs int64           `json:"total_activity_logs"`
	CountBySeverity   []SeverityCount `json:"count_by_severity"`
	CountByAction     []ActionCount   `json:"count_by_action"`
	CountByService    []ServiceCount  `json:"count_by_service"`
}
