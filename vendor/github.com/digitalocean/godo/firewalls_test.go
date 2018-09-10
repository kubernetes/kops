package godo

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"reflect"
	"testing"
)

var (
	firewallCreateJSONBody = `
{
  "name": "f-i-r-e-w-a-l-l",
  "inbound_rules": [
    {
      "protocol": "icmp",
      "sources": {
        "addresses": ["0.0.0.0/0"],
        "tags": ["frontend"],
        "droplet_ids": [123, 456],
        "load_balancer_uids": ["lb-uid"]
      }
    },
    {
      "protocol": "tcp",
      "ports": "8000-9000",
      "sources": {
        "addresses": ["0.0.0.0/0"]
      }
    }
  ],
  "outbound_rules": [
    {
      "protocol": "icmp",
      "destinations": {
        "tags": ["frontend"]
      }
    },
    {
      "protocol": "tcp",
      "ports": "8000-9000",
      "destinations": {
        "addresses": ["::/1"]
      }
    }
  ],
  "droplet_ids": [123],
  "tags": ["frontend"]
}
`
	firewallRulesJSONBody = `
{
  "inbound_rules": [
    {
      "protocol": "tcp",
      "ports": "22",
      "sources": {
        "addresses": ["0.0.0.0/0"]
      }
    }
  ],
  "outbound_rules": [
    {
      "protocol": "tcp",
      "ports": "443",
      "destinations": {
        "addresses": ["0.0.0.0/0"]
      }
    }
  ]
}
`
	firewallUpdateJSONBody = `
{
  "name": "f-i-r-e-w-a-l-l",
  "inbound_rules": [
    {
      "protocol": "tcp",
      "ports": "443",
      "sources": {
        "addresses": ["10.0.0.0/8"]
      }
    }
  ],
  "droplet_ids": [123],
  "tags": []
}
`
	firewallUpdateJSONResponse = `
{
  "firewall": {
    "id": "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
    "name": "f-i-r-e-w-a-l-l",
    "inbound_rules": [
      {
        "protocol": "tcp",
        "ports": "443",
        "sources": {
          "addresses": ["10.0.0.0/8"]
        }
      }
    ],
    "outbound_rules": [],
    "created_at": "2017-04-06T13:07:27Z",
    "droplet_ids": [
      123
    ],
    "tags": []
  }
}
`
	firewallJSONResponse = `
{
  "firewall": {
    "id": "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
    "name": "f-i-r-e-w-a-l-l",
    "status": "waiting",
    "inbound_rules": [
      {
        "protocol": "icmp",
        "ports": "0",
        "sources": {
          "tags": ["frontend"]
        }
      },
      {
        "protocol": "tcp",
        "ports": "8000-9000",
        "sources": {
          "addresses": ["0.0.0.0/0"]
        }
      }
    ],
    "outbound_rules": [
      {
        "protocol": "icmp",
        "ports": "0"
      },
      {
        "protocol": "tcp",
        "ports": "8000-9000",
        "destinations": {
          "addresses": ["::/1"]
        }
      }
    ],
    "created_at": "2017-04-06T13:07:27Z",
    "droplet_ids": [
      123
    ],
    "tags": [
      "frontend"
    ],
    "pending_changes": [
      {
        "droplet_id": 123,
        "removing": false,
        "status": "waiting"
      }
    ]
  }
}
`
	firewallListJSONResponse = `
{
  "firewalls": [
    {
      "id": "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
      "name": "f-i-r-e-w-a-l-l",
      "inbound_rules": [
        {
          "protocol": "icmp",
          "ports": "0",
          "sources": {
            "tags": ["frontend"]
          }
        },
        {
          "protocol": "tcp",
          "ports": "8000-9000",
          "sources": {
            "addresses": ["0.0.0.0/0"]
          }
        }
      ],
      "outbound_rules": [
        {
          "protocol": "icmp",
          "ports": "0"
        },
        {
          "protocol": "tcp",
          "ports": "8000-9000",
          "destinations": {
            "addresses": ["::/1"]
          }
        }
      ],
      "created_at": "2017-04-06T13:07:27Z",
      "droplet_ids": [
        123
      ],
      "tags": [
        "frontend"
      ]
    }
  ],
  "links": {},
  "meta": {
    "total": 1
  }
}
`
)

func TestFirewalls_Get(t *testing.T) {
	setup()
	defer teardown()

	urlStr := "/v2/firewalls"
	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr = path.Join(urlStr, fID)

	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, firewallJSONResponse)
	})

	actualFirewall, _, err := client.Firewalls.Get(ctx, fID)
	if err != nil {
		t.Errorf("Firewalls.Get returned error: %v", err)
	}

	expectedFirewall := &Firewall{
		ID:     "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
		Name:   "f-i-r-e-w-a-l-l",
		Status: "waiting",
		InboundRules: []InboundRule{
			{
				Protocol:  "icmp",
				PortRange: "0",
				Sources: &Sources{
					Tags: []string{"frontend"},
				},
			},
			{
				Protocol:  "tcp",
				PortRange: "8000-9000",
				Sources: &Sources{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
		OutboundRules: []OutboundRule{
			{
				Protocol:  "icmp",
				PortRange: "0",
			},
			{
				Protocol:  "tcp",
				PortRange: "8000-9000",
				Destinations: &Destinations{
					Addresses: []string{"::/1"},
				},
			},
		},
		Created:    "2017-04-06T13:07:27Z",
		DropletIDs: []int{123},
		Tags:       []string{"frontend"},
		PendingChanges: []PendingChange{
			{
				DropletID: 123,
				Removing:  false,
				Status:    "waiting",
			},
		},
	}

	if !reflect.DeepEqual(actualFirewall, expectedFirewall) {
		t.Errorf("Firewalls.Get returned %+v, expected %+v", actualFirewall, expectedFirewall)
	}
}

func TestFirewalls_Create(t *testing.T) {
	setup()
	defer teardown()

	expectedFirewallRequest := &FirewallRequest{
		Name: "f-i-r-e-w-a-l-l",
		InboundRules: []InboundRule{
			{
				Protocol: "icmp",
				Sources: &Sources{
					Addresses:        []string{"0.0.0.0/0"},
					Tags:             []string{"frontend"},
					DropletIDs:       []int{123, 456},
					LoadBalancerUIDs: []string{"lb-uid"},
				},
			},
			{
				Protocol:  "tcp",
				PortRange: "8000-9000",
				Sources: &Sources{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
		OutboundRules: []OutboundRule{
			{
				Protocol: "icmp",
				Destinations: &Destinations{
					Tags: []string{"frontend"},
				},
			},
			{
				Protocol:  "tcp",
				PortRange: "8000-9000",
				Destinations: &Destinations{
					Addresses: []string{"::/1"},
				},
			},
		},
		DropletIDs: []int{123},
		Tags:       []string{"frontend"},
	}

	mux.HandleFunc("/v2/firewalls", func(w http.ResponseWriter, r *http.Request) {
		v := new(FirewallRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, expectedFirewallRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, expectedFirewallRequest)
		}

		var actualFirewallRequest *FirewallRequest
		json.Unmarshal([]byte(firewallCreateJSONBody), &actualFirewallRequest)
		if !reflect.DeepEqual(actualFirewallRequest, expectedFirewallRequest) {
			t.Errorf("Request body = %+v, expected %+v", actualFirewallRequest, expectedFirewallRequest)
		}

		fmt.Fprint(w, firewallJSONResponse)
	})

	actualFirewall, _, err := client.Firewalls.Create(ctx, expectedFirewallRequest)
	if err != nil {
		t.Errorf("Firewalls.Create returned error: %v", err)
	}

	expectedFirewall := &Firewall{
		ID:     "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
		Name:   "f-i-r-e-w-a-l-l",
		Status: "waiting",
		InboundRules: []InboundRule{
			{
				Protocol:  "icmp",
				PortRange: "0",
				Sources: &Sources{
					Tags: []string{"frontend"},
				},
			},
			{
				Protocol:  "tcp",
				PortRange: "8000-9000",
				Sources: &Sources{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
		OutboundRules: []OutboundRule{
			{
				Protocol:  "icmp",
				PortRange: "0",
			},
			{
				Protocol:  "tcp",
				PortRange: "8000-9000",
				Destinations: &Destinations{
					Addresses: []string{"::/1"},
				},
			},
		},
		Created:    "2017-04-06T13:07:27Z",
		DropletIDs: []int{123},
		Tags:       []string{"frontend"},
		PendingChanges: []PendingChange{
			{
				DropletID: 123,
				Removing:  false,
				Status:    "waiting",
			},
		},
	}

	if !reflect.DeepEqual(actualFirewall, expectedFirewall) {
		t.Errorf("Firewalls.Create returned %+v, expected %+v", actualFirewall, expectedFirewall)
	}
}

func TestFirewalls_Update(t *testing.T) {
	setup()
	defer teardown()

	expectedFirewallRequest := &FirewallRequest{
		Name: "f-i-r-e-w-a-l-l",
		InboundRules: []InboundRule{
			{
				Protocol:  "tcp",
				PortRange: "443",
				Sources: &Sources{
					Addresses: []string{"10.0.0.0/8"},
				},
			},
		},
		DropletIDs: []int{123},
		Tags:       []string{},
	}

	urlStr := "/v2/firewalls"
	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr = path.Join(urlStr, fID)
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(FirewallRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, "PUT")
		if !reflect.DeepEqual(v, expectedFirewallRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, expectedFirewallRequest)
		}

		var actualFirewallRequest *FirewallRequest
		json.Unmarshal([]byte(firewallUpdateJSONBody), &actualFirewallRequest)
		if !reflect.DeepEqual(actualFirewallRequest, expectedFirewallRequest) {
			t.Errorf("Request body = %+v, expected %+v", actualFirewallRequest, expectedFirewallRequest)
		}

		fmt.Fprint(w, firewallUpdateJSONResponse)
	})

	actualFirewall, _, err := client.Firewalls.Update(ctx, fID, expectedFirewallRequest)
	if err != nil {
		t.Errorf("Firewalls.Update returned error: %v", err)
	}

	expectedFirewall := &Firewall{
		ID:   "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
		Name: "f-i-r-e-w-a-l-l",
		InboundRules: []InboundRule{
			{
				Protocol:  "tcp",
				PortRange: "443",
				Sources: &Sources{
					Addresses: []string{"10.0.0.0/8"},
				},
			},
		},
		OutboundRules: []OutboundRule{},
		Created:       "2017-04-06T13:07:27Z",
		DropletIDs:    []int{123},
		Tags:          []string{},
	}

	if !reflect.DeepEqual(actualFirewall, expectedFirewall) {
		t.Errorf("Firewalls.Update returned %+v, expected %+v", actualFirewall, expectedFirewall)
	}
}

func TestFirewalls_Delete(t *testing.T) {
	setup()
	defer teardown()

	urlStr := "/v2/firewalls"
	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr = path.Join(urlStr, fID)
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
	})

	_, err := client.Firewalls.Delete(ctx, fID)

	if err != nil {
		t.Errorf("Firewalls.Delete returned error: %v", err)
	}
}

func TestFirewalls_List(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/firewalls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, firewallListJSONResponse)
	})

	actualFirewalls, _, err := client.Firewalls.List(ctx, nil)

	if err != nil {
		t.Errorf("Firewalls.List returned error: %v", err)
	}

	expectedFirewalls := makeExpectedFirewalls()
	if !reflect.DeepEqual(actualFirewalls, expectedFirewalls) {
		t.Errorf("Firewalls.List returned %+v, expected %+v", actualFirewalls, expectedFirewalls)
	}
}

func TestFirewalls_ListByDroplet(t *testing.T) {
	setup()
	defer teardown()

	mux.HandleFunc("/v2/droplets/123/firewalls", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, firewallListJSONResponse)
	})

	actualFirewalls, _, err := client.Firewalls.ListByDroplet(ctx, 123, nil)

	if err != nil {
		t.Errorf("Firewalls.List returned error: %v", err)
	}

	expectedFirewalls := makeExpectedFirewalls()
	if !reflect.DeepEqual(actualFirewalls, expectedFirewalls) {
		t.Errorf("Firewalls.List returned %+v, expected %+v", actualFirewalls, expectedFirewalls)
	}
}

func TestFirewalls_AddDroplets(t *testing.T) {
	setup()
	defer teardown()

	dRequest := &dropletsRequest{
		IDs: []int{123},
	}

	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr := path.Join("/v2/firewalls", fID, "droplets")
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(dropletsRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, dRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, dRequest)
		}

		expectedJSONBody := `{"droplet_ids": [123]}`
		var actualDropletsRequest *dropletsRequest
		json.Unmarshal([]byte(expectedJSONBody), &actualDropletsRequest)
		if !reflect.DeepEqual(actualDropletsRequest, dRequest) {
			t.Errorf("Request body = %+v, expected %+v", actualDropletsRequest, dRequest)
		}

		fmt.Fprint(w, nil)
	})

	_, err := client.Firewalls.AddDroplets(ctx, fID, dRequest.IDs...)

	if err != nil {
		t.Errorf("Firewalls.AddDroplets returned error: %v", err)
	}
}

func TestFirewalls_RemoveDroplets(t *testing.T) {
	setup()
	defer teardown()

	dRequest := &dropletsRequest{
		IDs: []int{123, 345},
	}

	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr := path.Join("/v2/firewalls", fID, "droplets")
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(dropletsRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodDelete)
		if !reflect.DeepEqual(v, dRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, dRequest)
		}

		expectedJSONBody := `{"droplet_ids": [123, 345]}`
		var actualDropletsRequest *dropletsRequest
		json.Unmarshal([]byte(expectedJSONBody), &actualDropletsRequest)
		if !reflect.DeepEqual(actualDropletsRequest, dRequest) {
			t.Errorf("Request body = %+v, expected %+v", actualDropletsRequest, dRequest)
		}

		fmt.Fprint(w, nil)
	})

	_, err := client.Firewalls.RemoveDroplets(ctx, fID, dRequest.IDs...)

	if err != nil {
		t.Errorf("Firewalls.RemoveDroplets returned error: %v", err)
	}
}

func TestFirewalls_AddTags(t *testing.T) {
	setup()
	defer teardown()

	tRequest := &tagsRequest{
		Tags: []string{"frontend"},
	}

	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr := path.Join("/v2/firewalls", fID, "tags")
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(tagsRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, tRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, tRequest)
		}

		var actualTagsRequest *tagsRequest
		json.Unmarshal([]byte(`{"tags": ["frontend"]}`), &actualTagsRequest)
		if !reflect.DeepEqual(actualTagsRequest, tRequest) {
			t.Errorf("Request body = %+v, expected %+v", actualTagsRequest, tRequest)
		}

		fmt.Fprint(w, nil)
	})

	_, err := client.Firewalls.AddTags(ctx, fID, tRequest.Tags...)

	if err != nil {
		t.Errorf("Firewalls.AddTags returned error: %v", err)
	}
}

func TestFirewalls_RemoveTags(t *testing.T) {
	setup()
	defer teardown()

	tRequest := &tagsRequest{
		Tags: []string{"frontend", "backend"},
	}

	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr := path.Join("/v2/firewalls", fID, "tags")
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(tagsRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodDelete)
		if !reflect.DeepEqual(v, tRequest) {
			t.Errorf("Request body = %+v, expected %+v", v, tRequest)
		}

		var actualTagsRequest *tagsRequest
		json.Unmarshal([]byte(`{"tags": ["frontend", "backend"]}`), &actualTagsRequest)
		if !reflect.DeepEqual(actualTagsRequest, tRequest) {
			t.Errorf("Request body = %+v, expected %+v", actualTagsRequest, tRequest)
		}

		fmt.Fprint(w, nil)
	})

	_, err := client.Firewalls.RemoveTags(ctx, fID, tRequest.Tags...)

	if err != nil {
		t.Errorf("Firewalls.RemoveTags returned error: %v", err)
	}
}

func TestFirewalls_AddRules(t *testing.T) {
	setup()
	defer teardown()

	rr := &FirewallRulesRequest{
		InboundRules: []InboundRule{
			{
				Protocol:  "tcp",
				PortRange: "22",
				Sources: &Sources{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
		OutboundRules: []OutboundRule{
			{
				Protocol:  "tcp",
				PortRange: "443",
				Destinations: &Destinations{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
	}

	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr := path.Join("/v2/firewalls", fID, "rules")
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(FirewallRulesRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodPost)
		if !reflect.DeepEqual(v, rr) {
			t.Errorf("Request body = %+v, expected %+v", v, rr)
		}

		var actualFirewallRulesRequest *FirewallRulesRequest
		json.Unmarshal([]byte(firewallRulesJSONBody), &actualFirewallRulesRequest)
		if !reflect.DeepEqual(actualFirewallRulesRequest, rr) {
			t.Errorf("Request body = %+v, expected %+v", actualFirewallRulesRequest, rr)
		}

		fmt.Fprint(w, nil)
	})

	_, err := client.Firewalls.AddRules(ctx, fID, rr)

	if err != nil {
		t.Errorf("Firewalls.AddRules returned error: %v", err)
	}
}

func TestFirewalls_RemoveRules(t *testing.T) {
	setup()
	defer teardown()

	rr := &FirewallRulesRequest{
		InboundRules: []InboundRule{
			{
				Protocol:  "tcp",
				PortRange: "22",
				Sources: &Sources{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
		OutboundRules: []OutboundRule{
			{
				Protocol:  "tcp",
				PortRange: "443",
				Destinations: &Destinations{
					Addresses: []string{"0.0.0.0/0"},
				},
			},
		},
	}

	fID := "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0"
	urlStr := path.Join("/v2/firewalls", fID, "rules")
	mux.HandleFunc(urlStr, func(w http.ResponseWriter, r *http.Request) {
		v := new(FirewallRulesRequest)
		err := json.NewDecoder(r.Body).Decode(v)
		if err != nil {
			t.Fatal(err)
		}

		testMethod(t, r, http.MethodDelete)
		if !reflect.DeepEqual(v, rr) {
			t.Errorf("Request body = %+v, expected %+v", v, rr)
		}

		var actualFirewallRulesRequest *FirewallRulesRequest
		json.Unmarshal([]byte(firewallRulesJSONBody), &actualFirewallRulesRequest)
		if !reflect.DeepEqual(actualFirewallRulesRequest, rr) {
			t.Errorf("Request body = %+v, expected %+v", actualFirewallRulesRequest, rr)
		}

		fmt.Fprint(w, nil)
	})

	_, err := client.Firewalls.RemoveRules(ctx, fID, rr)

	if err != nil {
		t.Errorf("Firewalls.RemoveRules returned error: %v", err)
	}
}

func makeExpectedFirewalls() []Firewall {
	return []Firewall{
		Firewall{
			ID:   "fe6b88f2-b42b-4bf7-bbd3-5ae20208f0b0",
			Name: "f-i-r-e-w-a-l-l",
			InboundRules: []InboundRule{
				{
					Protocol:  "icmp",
					PortRange: "0",
					Sources: &Sources{
						Tags: []string{"frontend"},
					},
				},
				{
					Protocol:  "tcp",
					PortRange: "8000-9000",
					Sources: &Sources{
						Addresses: []string{"0.0.0.0/0"},
					},
				},
			},
			OutboundRules: []OutboundRule{
				{
					Protocol:  "icmp",
					PortRange: "0",
				},
				{
					Protocol:  "tcp",
					PortRange: "8000-9000",
					Destinations: &Destinations{
						Addresses: []string{"::/1"},
					},
				},
			},
			DropletIDs: []int{123},
			Tags:       []string{"frontend"},
			Created:    "2017-04-06T13:07:27Z",
		},
	}
}
