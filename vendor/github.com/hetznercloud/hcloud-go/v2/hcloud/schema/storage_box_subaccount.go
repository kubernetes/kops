package schema

import "time"

// StorageBoxSubaccount defines the schema of a Storage Box subaccount.
type StorageBoxSubaccount struct {
	ID             int64                              `json:"id"`
	Name           string                             `json:"name"`
	Username       string                             `json:"username"`
	HomeDirectory  string                             `json:"home_directory"`
	Server         string                             `json:"server"`
	AccessSettings StorageBoxSubaccountAccessSettings `json:"access_settings"`
	Description    string                             `json:"description"`
	Labels         map[string]string                  `json:"labels"`
	Created        time.Time                          `json:"created"`
	StorageBox     int64                              `json:"storage_box"`
}

// StorageBoxSubaccountAccessSettings defines the schema of a Storage Box subaccount's access settings.
type StorageBoxSubaccountAccessSettings struct {
	ReachableExternally bool `json:"reachable_externally"`
	Readonly            bool `json:"readonly"`
	SambaEnabled        bool `json:"samba_enabled"`
	SSHEnabled          bool `json:"ssh_enabled"`
	WebDAVEnabled       bool `json:"webdav_enabled"`
}

// StorageBoxSubaccountGetResponse defines the schema of the response when retrieving a single Storage Box subaccount.
type StorageBoxSubaccountGetResponse struct {
	Subaccount StorageBoxSubaccount `json:"subaccount"`
}

// StorageBoxSubaccountListResponse defines the schema of the response when listing Storage Box subaccounts.
type StorageBoxSubaccountListResponse struct {
	Subaccounts []StorageBoxSubaccount `json:"subaccounts"`
}

// StorageBoxSubaccountCreateRequest defines the schema of the request when creating a Storage Box subaccount.
type StorageBoxSubaccountCreateRequest struct {
	Name           string                                           `json:"name,omitempty"`
	HomeDirectory  string                                           `json:"home_directory"`
	Password       string                                           `json:"password"`
	Description    string                                           `json:"description,omitempty"`
	AccessSettings *StorageBoxSubaccountCreateRequestAccessSettings `json:"access_settings,omitempty"`
	Labels         map[string]string                                `json:"labels,omitempty"`
}

// StorageBoxSubaccountCreateRequestAccessSettings defines the schema of the access settings in the
// request when creating a Storage Box subaccount.
type StorageBoxSubaccountCreateRequestAccessSettings struct {
	ReachableExternally *bool `json:"reachable_externally,omitempty"`
	Readonly            *bool `json:"readonly,omitempty"`
	SambaEnabled        *bool `json:"samba_enabled,omitempty"`
	SSHEnabled          *bool `json:"ssh_enabled,omitempty"`
	WebDAVEnabled       *bool `json:"webdav_enabled,omitempty"`
}

// StorageBoxSubaccountCreateResponse defines the schema of the response when creating a Storage Box subaccount.
type StorageBoxSubaccountCreateResponse struct {
	Subaccount StorageBoxSubaccountCreateResponseSubaccount `json:"subaccount"`
	Action     Action                                       `json:"action"`
}

// StorageBoxSubaccountCreateResponseSubaccount defines the schema of the subaccount in the response
// when creating a Storage Box subaccount.
type StorageBoxSubaccountCreateResponseSubaccount struct {
	ID         int64 `json:"id"`
	StorageBox int64 `json:"storage_box"`
}

// StorageBoxSubaccountUpdateRequest defines the schema of the request when updating a Storage Box subaccount.
type StorageBoxSubaccountUpdateRequest struct {
	Name        string             `json:"name,omitempty"`
	Description *string            `json:"description,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
}

// StorageBoxSubaccountUpdateResponse defines the schema of the response when updating a Storage Box subaccount.
type StorageBoxSubaccountUpdateResponse struct {
	Subaccount StorageBoxSubaccount `json:"subaccount"`
}

// StorageBoxSubaccountResetPasswordRequest defines the schema of the request when resetting a
// Storage Box subaccount's password.
type StorageBoxSubaccountResetPasswordRequest struct {
	Password string `json:"password"`
}

// StorageBoxSubaccountUpdateAccessSettingsRequest defines the schema of the request when updating
// Storage Box subaccount's access settings.
type StorageBoxSubaccountUpdateAccessSettingsRequest struct {
	ReachableExternally *bool `json:"reachable_externally,omitempty"`
	Readonly            *bool `json:"readonly,omitempty"`
	SambaEnabled        *bool `json:"samba_enabled,omitempty"`
	SSHEnabled          *bool `json:"ssh_enabled,omitempty"`
	WebDAVEnabled       *bool `json:"webdav_enabled,omitempty"`
}

// StorageBoxSubaccountChangeHomeDirectoryRequest defines the schema of the request when changing
// the home directory of a Storage Box subaccount.
type StorageBoxSubaccountChangeHomeDirectoryRequest struct {
	HomeDirectory string `json:"home_directory"`
}
