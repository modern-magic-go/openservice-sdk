package enums

type NotifyStatus string

const (
	NotifyStatusInit    NotifyStatus = "Init"
	NotifyStatusRetry   NotifyStatus = "Retry"
	NotifyStatusSuccess NotifyStatus = "Success"
	NotifyStatusFailed  NotifyStatus = "Failed"
)

func (s NotifyStatus) String() string { return string(s) }

func (s NotifyStatus) IsValid() bool {
	switch s {
	case NotifyStatusInit, NotifyStatusRetry, NotifyStatusSuccess, NotifyStatusFailed:
		return true
	default:
		return false
	}
}

func NotifyValues() []NotifyStatus {
	return []NotifyStatus{
		NotifyStatusInit,
		NotifyStatusRetry,
		NotifyStatusSuccess,
		NotifyStatusFailed,
	}
}
