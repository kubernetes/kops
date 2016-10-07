package tags

const (
	TagOSFamilyCentos = "_centos_family"
	TagOSFamilyDebian = "_debian_family"

	TagSystemd = "_systemd"
)

type HasTags interface {
	HasTag(tag string) bool
}
