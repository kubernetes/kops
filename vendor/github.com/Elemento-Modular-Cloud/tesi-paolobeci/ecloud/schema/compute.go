package schema

import (
	"time"
)

// -------- HEALTH --------
type HealthCheckComputeResponse struct {
	Message string `json:"message"`
}

type CanAllocateComputeResponse struct {
	Slots         	int `json:"slots"`
	Overprovision 	bool `json:"overprovision"`
	AllowSMT      	bool `json:"allowSMT"`
	Archs         	[]string `json:"archs"`
	Flags         	[]string `json:"flags"`
	Ramsize       	int `json:"ramSize"` // MB
	ReqECC        	bool `json:"reqECC"`
	Misc          	[]string `json:"misc"`
	Pci           	[]string `json:"pci"`
}

// -------- CREATE COMPUTE --------
type CreateComputeRequest struct {
	Name   			string `json:"vm_name"`
	Slots         	int `json:"slots"`
	Overprovision 	bool `json:"overprovision"`
	AllowSMT      	bool `json:"allowSMT"`
	Archs         	[]string `json:"archs"`
	Flags         	[]string `json:"flags"`
	Ramsize       	int `json:"ramSize"`
	ReqECC        	bool `json:"reqECC"`
	Misc          	[]string `json:"misc"`
	Pci          	[]string `json:"pci"`
	Volumes   	  	[]string `json:"volumes"`
	Netdevs			[]string `json:"netdevs"`
}

type CreateComputeResponse struct {}


// -------- COMPUTE STATUS --------
type ComputeStatusResponse struct {
	Servers []Server `json:"servers"`
}

type Server struct {
	Name		  string        `json:"name"`
	ServerURL     string        `json:"serverurl"`
	IsGateway     bool          `json:"is_gateway"`
	UniqueID      string        `json:"uniqueID"`
	ReqJSON       RequestConfig `json:"req_json"`
	Volumes       []Volume      `json:"volumes"`
	CreationDate  time.Time     `json:"creation_date"`
	NetworkConfig NetworkConfig `json:"network_config"`
	Status 	  	  string        `json:"status"`
	Labels 	      map[string]string `json:"labels"`
}

type RequestConfig struct {
	Slots         int      `json:"slots"`
	Overprovision int      `json:"overprovision"`
	AllowSMT      bool     `json:"allowSMT"`
	Arch          string   `json:"arch"`
	Flags         []string `json:"flags"`
	RamSize       int      `json:"ramsize"`
	ReqECC        bool     `json:"reqECC"`
	Volumes       []Volume `json:"volumes"`
	PciDevs       []string `json:"pcidevs"`
	NetDevs       []string `json:"netdevs"`
	OSFamily      string   `json:"os_family"`
	OSFlavour     string   `json:"os_flavour"`
	VMName        string   `json:"vm_name"`
}

type Volume struct {
	Bootable        bool     `json:"bootable"`
	CreatorID       string   `json:"creatorID"`
	Name            string   `json:"name"`
	NumServers      int      `json:"nservers"`
	Own             bool     `json:"own"`
	Private         bool     `json:"private"`
	ReadOnly        bool     `json:"readonly"`
	Server          string   `json:"server"`
	Servers         []string `json:"servers"`
	ServerURL       string   `json:"serverurl"`
	Shareable       bool     `json:"shareable"`
	Size            int64    `json:"size"`
	VolumeID        string   `json:"volumeID"`
	Vid             string   `json:"vid"`
	SelectedServer  string   `json:"selected_server"`
	ISCSIName       string   `json:"iscsi_name"`
	Driver          string   `json:"driver"`
}

type NetworkConfig struct {
	Name       string       `json:"name"`
	Interface  string       `json:"interface"`
	Type       string       `json:"type"`
	Source     string       `json:"source"`
	Model      string       `json:"model"`
	MAC        string       `json:"mac"`
	DomDisplay NetworkDisplay `json:"dom_display"`
	IPv4       *string      `json:"ipv4"` // Assuming IPv4 could be null
}

type NetworkDisplay struct {
	Protocol string `json:"protocol"`
	Port     int    `json:"port"`
}

// -------- COMPUTE TEMPLATES --------
type ComputeTemplatesResponse struct {
	Templates []ComputeTemplate `json:"templates"`
}

type ComputeTemplate struct {
	Info struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		// TODO ...
	}
}

// -------- COMPUTE DELETE --------
type DeleteComputeRequest struct {
	Name string `json:"vm_name"`
}
type DeleteComputeResponse struct {}