package enums

type TransStatus string

const (
	TransStatusInit    TransStatus = "Init"
	TransStatusPending TransStatus = "Pending"
	TransStatusSuccess TransStatus = "Success"
	TransStatusFailed  TransStatus = "Failed"
)

func (s TransStatus) String() string { return string(s) }

func (s TransStatus) IsValid() bool {
	switch s {
	case TransStatusInit, TransStatusPending, TransStatusSuccess, TransStatusFailed:
		return true
	default:
		return false
	}
}

func Values() []TransStatus {
	return []TransStatus{
		TransStatusInit,
		TransStatusPending,
		TransStatusSuccess,
		TransStatusFailed,
	}
}
