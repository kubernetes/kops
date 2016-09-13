package terraform

type Lifecycle struct {
	PreventDestroy      *bool `json:"prevent_destroy,omitempty"`
	CreateBeforeDestroy *bool `json:"create_before_destroy,omitempty"`
}
