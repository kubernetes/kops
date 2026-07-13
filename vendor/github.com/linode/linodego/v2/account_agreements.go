package linodego

import "context"

// AccountAgreements represents the agreements and their acceptance status for an Account
type AccountAgreements struct {
	EUModel                bool `json:"eu_model"`
	MasterServiceAgreement bool `json:"master_service_agreement"`
	PrivacyPolicy          bool `json:"privacy_policy"`
}

// AccountAgreementsUpdateOptions fields are those accepted by UpdateAccountAgreements
type AccountAgreementsUpdateOptions struct {
	EUModel                bool `json:"eu_model,omitzero"`
	MasterServiceAgreement bool `json:"master_service_agreement,omitzero"`
	PrivacyPolicy          bool `json:"privacy_policy,omitzero"`
}

// GetUpdateOptions converts an AccountAgreements to AccountAgreementsUpdateOptions for use in UpdateAccountAgreements
func (i AccountAgreements) GetUpdateOptions() (o AccountAgreementsUpdateOptions) {
	o.EUModel = i.EUModel
	o.MasterServiceAgreement = i.MasterServiceAgreement
	o.PrivacyPolicy = i.PrivacyPolicy

	return o
}

// GetAccountAgreements gets all agreements and their acceptance status for the Account.
func (c *Client) GetAccountAgreements(ctx context.Context) (*AccountAgreements, error) {
	return doGETRequest[AccountAgreements](ctx, c, "account/agreements")
}

// AcknowledgeAccountAgreements acknowledges account agreements for the Account
func (c *Client) AcknowledgeAccountAgreements(ctx context.Context, opts AccountAgreementsUpdateOptions) error {
	return doPOSTRequestNoResponseBody(ctx, c, "account/agreements", opts)
}
