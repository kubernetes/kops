package schema

import "time"

// StorageBoxSnapshot defines the schema of a Storage Box snapshot.
type StorageBoxSnapshot struct {
	ID          int64                   `json:"id"`
	Name        string                  `json:"name"`
	Description string                  `json:"description"`
	Stats       StorageBoxSnapshotStats `json:"stats"`
	IsAutomatic bool                    `json:"is_automatic"`
	Labels      map[string]string       `json:"labels"`
	Created     time.Time               `json:"created"`
	StorageBox  int64                   `json:"storage_box"`
}

// StorageBoxSnapshotStats defines the schema of a Storage Box snapshot's size statistics.
type StorageBoxSnapshotStats struct {
	Size           uint64 `json:"size"`
	SizeFilesystem uint64 `json:"size_filesystem"`
}

// StorageBoxSnapshotGetResponse defines the schema of the response when retrieving a single Storage Box snapshot.
type StorageBoxSnapshotGetResponse struct {
	Snapshot StorageBoxSnapshot `json:"snapshot"`
}

// StorageBoxSnapshotListResponse defines the schema of the response when listing Storage Box snapshots.
type StorageBoxSnapshotListResponse struct {
	Snapshots []StorageBoxSnapshot `json:"snapshots"`
}

// StorageBoxSnapshotCreateRequest defines the schema of the request to create a Storage Box snapshot.
type StorageBoxSnapshotCreateRequest struct {
	Description string            `json:"description,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
}

// StorageBoxSnapshotCreateResponse defines the schema of the response when creating a Storage Box snapshot.
type StorageBoxSnapshotCreateResponse struct {
	Snapshot StorageBoxSnapshot `json:"snapshot"`
	Action   Action             `json:"action"`
}

// StorageBoxSnapshotUpdateRequest defines the schema of the request to update a Storage Box snapshot.
type StorageBoxSnapshotUpdateRequest struct {
	Description *string            `json:"description,omitempty"`
	Labels      *map[string]string `json:"labels,omitempty"`
}

// StorageBoxSnapshotUpdateResponse defines the schema of the response when updating a Storage Box snapshot.
type StorageBoxSnapshotUpdateResponse struct {
	Snapshot StorageBoxSnapshot `json:"snapshot"`
}
