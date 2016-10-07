package tags

const (
	TagOSFamilyRHEL   = "_rhel_family"
	TagOSFamilyDebian = "_debian_family"

	TagSystemd = "_systemd"
)

type HasTags interface {
	HasTag(tag string) bool
}
