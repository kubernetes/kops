package schema

import "time"

type PlacementGroup struct {
	ID      int               `json:"id"`
	Name    string            `json:"name"`
	Labels  map[string]string `json:"labels"`
	Created time.Time         `json:"created"`
	Servers []int             `json:"servers"`
	Type    string            `json:"type"`
}

type PlacementGroupListResponse struct {
	PlacementGroups []PlacementGroup `json:"placement_groups"`
}

type PlacementGroupGetResponse struct {
	PlacementGroup PlacementGroup `json:"placement_group"`
}

type PlacementGroupCreateRequest struct {
	Name   string             `json:"name"`
	Labels *map[string]string `json:"labels,omitempty"`
	Type   string             `json:"type"`
}

type PlacementGroupCreateResponse struct {
	PlacementGroup PlacementGroup `json:"placement_group"`
	Action         *Action        `json:"action"`
}

type PlacementGroupUpdateRequest struct {
	Name   *string            `json:"name,omitempty"`
	Labels *map[string]string `json:"labels,omitempty"`
}

type PlacementGroupUpdateResponse struct {
	PlacementGroup PlacementGroup `json:"placement_group"`
}
