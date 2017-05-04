package azuredns


const (
	ProviderName = "azure-azuredns"
)

// NewAzureDns creates a new instance of an Azure DNS Interface. Exported because
// we can't use the "k8s" way of doing things. If we ever port this, then we can
// change this to follow suit.
//func NewAzureDns(config io.Reader) (*Interface, error) {
func NewAzureDns() (*Interface, error) {
	i := &Interface{}
	return i, nil
}