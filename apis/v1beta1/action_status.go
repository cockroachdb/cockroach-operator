package v1beta1

type ActionStatus int

const (
	// Failed status indicates the action taken by the operator completed
	// unsuccessfully.
	Failed = iota
	// Starting status indicates the operator identified the action needed
	// to be taken and will start performing the action.
	Starting
	// Finished status indicates the action taken by the operator completed
	// successfully.
	Finished
	// Unknown status is used when the operator cannot determine if the
	// action is complete or in progress.
	Unknown
)

var statuses []string = []string{
	"Failed",
	"Starting",
	"Finished",
	"Unknown",
}

func (a ActionStatus) String() string {
	if a < Failed || a > Unknown {
		return "Unknown"
	}
	return statuses[a]
}