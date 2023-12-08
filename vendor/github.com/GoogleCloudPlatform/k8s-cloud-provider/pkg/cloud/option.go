package cloud

// Option are optional parameters to the generated methods.
type Option interface {
	mergeInto(all *allOptions)
}

// allOptions that can be configured for the generated methods.
type allOptions struct {
	projectID string
}

// ForceProjectID forces the projectID to be used in the call to be the one
// specified. This ignores the default routing done by the ProjectRouter.
func ForceProjectID(projectID string) Option { return projectIDOption(projectID) }

type projectIDOption string

func (opt projectIDOption) mergeInto(all *allOptions) { all.projectID = string(opt) }

func mergeOptions(options []Option) allOptions {
	var ret allOptions
	for _, opt := range options {
		opt.mergeInto(&ret)
	}
	return ret
}
