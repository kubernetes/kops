package linodego

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"slices"
	"strconv"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var englishTitle = cases.Title(language.English)

// EventPoller waits for events associated with a given entity and action.
type EventPoller struct {
	EntityID   any
	EntityType EntityType

	// Type is excluded here because it is implicitly determined
	// by the event action.
	SecondaryEntityID any

	Action EventAction

	client         Client
	previousEvents map[int]bool
}

// WaitForInstanceStatus waits for the Linode instance to reach the desired state
// before returning.
func (client Client) WaitForInstanceStatus(ctx context.Context, instanceID int, status InstanceStatus) (*Instance, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*Instance, bool, error) {
			instance, err := client.GetInstance(ctx, instanceID)
			if err != nil {
				return instance, false, err
			}

			return instance, instance.Status == status, nil
		},
		func() error {
			return fmt.Errorf("Error waiting for Instance %d status %s: %w", instanceID, status, ctx.Err())
		},
	)
}

// WaitForInstanceDiskStatus waits for the Linode instance disk to reach the desired state
// before returning.
func (client Client) WaitForInstanceDiskStatus(ctx context.Context, instanceID int, diskID int, status DiskStatus) (*InstanceDisk, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*InstanceDisk, bool, error) {
			// GetInstanceDisk will 404 on newly created disks. Use List instead.
			disks, err := client.ListInstanceDisks(ctx, instanceID, nil)
			if err != nil {
				return nil, false, err
			}

			for _, disk := range disks {
				if disk.ID == diskID {
					if disk.Status == status {
						return &disk, true, nil
					}

					break
				}
			}

			return nil, false, nil
		},
		func() error {
			return fmt.Errorf("Error waiting for Instance %d Disk %d status %s: %w", instanceID, diskID, status, ctx.Err())
		},
	)
}

// WaitForVolumeStatus waits for the Volume to reach the desired state
// before returning.
func (client Client) WaitForVolumeStatus(ctx context.Context, volumeID int, status VolumeStatus) (*Volume, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*Volume, bool, error) {
			volume, err := client.GetVolume(ctx, volumeID)
			if err != nil {
				return volume, false, err
			}

			return volume, volume.Status == status, nil
		},
		func() error {
			return fmt.Errorf("Error waiting for Volume %d status %s: %w", volumeID, status, ctx.Err())
		},
	)
}

// WaitForSnapshotStatus waits for the Snapshot to reach the desired state
// before returning.
func (client Client) WaitForSnapshotStatus(
	ctx context.Context,
	instanceID int,
	snapshotID int,
	status InstanceSnapshotStatus,
) (*InstanceSnapshot, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*InstanceSnapshot, bool, error) {
			snapshot, err := client.GetInstanceSnapshot(ctx, instanceID, snapshotID)
			if err != nil {
				return snapshot, false, err
			}

			return snapshot, snapshot.Status == status, nil
		},
		func() error {
			return fmt.Errorf("Error waiting for Instance %d Snapshot %d status %s: %w", instanceID, snapshotID, status, ctx.Err())
		},
	)
}

// WaitForVolumeLinodeID waits for the Volume to match the desired LinodeID
// before returning. An active Instance will not immediately attach or detach a volume, so
// the LinodeID must be polled to determine volume readiness from the API.
func (client Client) WaitForVolumeLinodeID(ctx context.Context, volumeID int, linodeID *int) (*Volume, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*Volume, bool, error) {
			volume, err := client.GetVolume(ctx, volumeID)
			if err != nil {
				return volume, false, err
			}

			switch {
			case linodeID == nil && volume.LinodeID == nil:
				return volume, true, nil
			case linodeID == nil || volume.LinodeID == nil:
				// Continue waiting.
			case *volume.LinodeID == *linodeID:
				return volume, true, nil
			}

			return volume, false, nil
		},
		func() error {
			return fmt.Errorf("Error waiting for Volume %d to have Instance %v: %w", volumeID, linodeID, ctx.Err())
		},
	)
}

// WaitForLKEClusterStatus waits for the LKECluster to reach the desired state
// before returning.
func (client Client) WaitForLKEClusterStatus(ctx context.Context, clusterID int, status LKEClusterStatus) (*LKECluster, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*LKECluster, bool, error) {
			cluster, err := client.GetLKECluster(ctx, clusterID)
			if err != nil {
				return cluster, false, err
			}

			return cluster, cluster.Status == status, nil
		},
		func() error {
			return fmt.Errorf("Error waiting for Cluster %d status %s: %w", clusterID, status, ctx.Err())
		},
	)
}

// LKEClusterPollOptions configures polls against LKE Clusters.
type LKEClusterPollOptions struct {
	// Retry will cause the Poll to ignore interimittent errors
	Retry bool

	// TansportWrapper allows adding a transport middleware function that will
	// wrap the LKE Cluster client's underlying http.RoundTripper.
	TransportWrapper func(http.RoundTripper) http.RoundTripper
}

// ClusterConditionOptions configures LKE cluster condition checks.
type ClusterConditionOptions struct {
	LKEClusterKubeconfig *LKEClusterKubeconfig
	TransportWrapper     func(http.RoundTripper) http.RoundTripper
}

// ClusterConditionFunc represents a function that tests a condition against an LKE cluster,
// returns true if the condition has been reached, false if it has not yet been reached.
type ClusterConditionFunc func(context.Context, ClusterConditionOptions) (bool, error)

// WaitForLKEClusterConditions waits for the given LKE conditions to be true
func (client Client) WaitForLKEClusterConditions(
	ctx context.Context,
	clusterID int,
	options LKEClusterPollOptions,
	conditions ...ClusterConditionFunc,
) error {
	lkeKubeConfig, err := client.GetLKEClusterKubeconfig(ctx, clusterID)
	if err != nil {
		return fmt.Errorf("failed to get Kubeconfig for LKE cluster %d: %w", clusterID, err)
	}

	ticker := newTicker(&client)
	defer ticker.Stop()

	conditionOptions := ClusterConditionOptions{LKEClusterKubeconfig: lkeKubeConfig, TransportWrapper: options.TransportWrapper}

	for _, condition := range conditions {
	ConditionSucceeded:
		for {
			select {
			case <-ticker.C:
				result, err := condition(ctx, conditionOptions)
				if err != nil {
					log.Printf("[WARN] Ignoring WaitForLKEClusterConditions conditional error: %s", err)

					if !options.Retry {
						return err
					}
				}

				if result {
					break ConditionSucceeded
				}

			case <-ctx.Done():
				return fmt.Errorf("Error waiting for cluster %d conditions: %w", clusterID, ctx.Err())
			}
		}
	}

	return nil
}

// WaitForEventFinished waits for an entity action to reach the 'finished' state
// before returning.
// If the event indicates a failure both the failed event and the error will be returned.
// nolint
func (client Client) WaitForEventFinished(
	ctx context.Context,
	id any,
	entityType EntityType,
	action EventAction,
	minStart time.Time,
) (*Event, error) {
	titledEntityType := englishTitle.String(string(entityType))
	filter := Filter{
		Order:   Descending,
		OrderBy: "created",
	}
	filter.AddField(Eq, "action", action)
	filter.AddField(Gte, "created", minStart.UTC().Format("2006-01-02T15:04:05"))

	// Optimistically restrict results to page 1.  We should remove this when more
	// precise filtering options exist.
	pages := 1

	// The API has limitted filtering support for Event ID and Event Type
	// Optimize the list, if possible
	switch entityType {
	case EntityDisk, EntityDatabase, EntityLinode, EntityDomain, EntityNodebalancer:
		// All of the filter supported types have int ids
		filterableEntityID, err := strconv.Atoi(fmt.Sprintf("%v", id))
		if err != nil {
			return nil, fmt.Errorf("error parsing Entity ID %q for optimized "+
				"WaitForEventFinished EventType %q: %w", id, entityType, err)
		}
		filter.AddField(Eq, "entity.id", filterableEntityID)
		filter.AddField(Eq, "entity.type", entityType)
	}

	if deadline, ok := ctx.Deadline(); ok {
		log.Printf("[INFO] Waiting %d seconds for %s events since %v for %s %v", int(time.Until(deadline).Seconds()), action, minStart, titledEntityType, id)
	}

	ticker := newTicker(&client)

	// avoid repeating log messages
	nextLog := ""
	lastLog := ""
	lastEventID := 0

	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if lastEventID > 0 {
				filter.AddField(Gte, "id", lastEventID)
			}

			filterStr, err := filter.MarshalJSON()
			if err != nil {
				return nil, err
			}

			listOptions := NewListOptions(pages, string(filterStr))

			events, err := client.ListEvents(ctx, listOptions)
			if err != nil {
				return nil, err
			}

			// If there are events for this instance + action, inspect them
			for _, event := range events {
				if event.Entity == nil || event.Entity.Type != entityType {
					// log.Println("type mismatch", event.Entity.Type, entityType)
					continue
				}

				var entID string

				switch id := event.Entity.ID.(type) {
				case float64, float32:
					entID = fmt.Sprintf("%.f", id)
				case int:
					entID = strconv.Itoa(id)
				default:
					entID = fmt.Sprintf("%v", id)
				}

				var findID string

				switch id := id.(type) {
				case float64, float32:
					findID = fmt.Sprintf("%.f", id)
				case int:
					findID = strconv.Itoa(id)
				default:
					findID = fmt.Sprintf("%v", id)
				}

				if entID != findID {
					// log.Println("id mismatch", entID, findID)
					continue
				}

				if event.Created == nil {
					log.Printf("[WARN] event.Created is nil when API returned: %#+v", event.Created)
				}

				// This is the event we are looking for. Save our place.
				if lastEventID == 0 {
					lastEventID = event.ID
				}

				switch event.Status {
				case EventFailed:
					return &event, fmt.Errorf("%s %v action %s failed", titledEntityType, id, action)
				case EventFinished:
					log.Printf("[INFO] %s %v action %s is finished", titledEntityType, id, action)
					return &event, nil
				}

				nextLog = fmt.Sprintf("[INFO] %s %v action %s is %s", titledEntityType, id, action, event.Status)
			}

			// de-dupe logging statements
			if nextLog != lastLog {
				log.Print(nextLog)
				lastLog = nextLog
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("Error waiting for Event Status '%s' of %s %v action '%s': %w", EventFinished, titledEntityType, id, action, ctx.Err())
		}
	}
}

// WaitForImageStatus waits for the Image to reach the desired state
// before returning.
func (client Client) WaitForImageStatus(ctx context.Context, imageID string, status ImageStatus) (*Image, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*Image, bool, error) {
			image, err := client.GetImage(ctx, imageID)
			if err != nil {
				return image, false, err
			}

			return image, image.Status == status, nil
		},
		func() error {
			return fmt.Errorf("failed to wait for Image %s status %s: %w", imageID, status, ctx.Err())
		},
	)
}

// WaitForImageRegionStatus waits for an Image's replica to reach the desired state
// before returning.
func (client Client) WaitForImageRegionStatus(ctx context.Context, imageID, region string, status ImageRegionStatus) (*Image, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*Image, bool, error) {
			image, err := client.GetImage(ctx, imageID)
			if err != nil {
				return image, false, err
			}

			replicaIdx := slices.IndexFunc(
				image.Regions,
				func(r ImageRegion) bool {
					return r.Region == region
				},
			)

			if replicaIdx < 0 || image.Regions[replicaIdx].Status != status {
				return image, false, nil
			}

			return image, true, nil
		},
		func() error {
			return fmt.Errorf("failed to wait for Image %s status %s: %w", imageID, status, ctx.Err())
		},
	)
}

type databaseStatusFunc func(ctx context.Context, client Client, dbID int) (DatabaseStatus, error)

var databaseStatusHandlers = map[DatabaseEngineType]databaseStatusFunc{
	DatabaseEngineTypeMySQL: func(ctx context.Context, client Client, dbID int) (DatabaseStatus, error) {
		db, err := client.GetMySQLDatabase(ctx, dbID)
		if err != nil {
			return "", err
		}

		return db.Status, nil
	},
	DatabaseEngineTypePostgres: func(ctx context.Context, client Client, dbID int) (DatabaseStatus, error) {
		db, err := client.GetPostgresDatabase(ctx, dbID)
		if err != nil {
			return "", err
		}

		return db.Status, nil
	},
}

// WaitForDatabaseStatus waits for the provided database to have the given status.
func (client Client) WaitForDatabaseStatus(
	ctx context.Context, dbID int, dbEngine DatabaseEngineType, status DatabaseStatus,
) error {
	_, err := poll(ctx, &client,
		func(ctx context.Context) (struct{}, bool, error) {
			statusHandler, ok := databaseStatusHandlers[dbEngine]
			if !ok {
				return struct{}{}, false, fmt.Errorf("invalid db engine: %s", dbEngine)
			}

			currentStatus, err := statusHandler(ctx, client, dbID)
			if err != nil {
				return struct{}{}, false, fmt.Errorf("failed to get db status: %w", err)
			}

			return struct{}{}, currentStatus == status, nil
		},
		func() error {
			return fmt.Errorf("failed to wait for database %d status: %w", dbID, ctx.Err())
		},
	)

	return err
}

// NewEventPoller initializes a new Linode event poller. This should be run before the event is triggered as it stores
// the previous state of the entity's events.
func (client Client) NewEventPoller(
	ctx context.Context, id any, entityType EntityType, action EventAction,
) (*EventPoller, error) {
	result := EventPoller{
		EntityID:   id,
		EntityType: entityType,
		Action:     action,

		client: client,
	}

	if err := result.preTask(ctx); err != nil {
		return nil, fmt.Errorf("failed to run pretask: %w", err)
	}

	return &result, nil
}

// NewEventPollerWithSecondary initializes a new Linode event poller with for events with a
// specific secondary entity.
func (client Client) NewEventPollerWithSecondary(
	ctx context.Context, id any, primaryEntityType EntityType, secondaryID int, action EventAction,
) (*EventPoller, error) {
	poller, err := client.NewEventPoller(ctx, id, primaryEntityType, action)
	if err != nil {
		return nil, err
	}

	poller.SecondaryEntityID = secondaryID

	return poller, nil
}

// NewEventPollerWithoutEntity initializes a new Linode event poller without a target entity ID.
// This is useful for create events where the ID of the entity is not yet known.
// For example:
// p, _ := client.NewEventPollerWithoutEntity(...)
// inst, _ := client.CreateInstance(...)
// p.EntityID = inst.ID
// ...
func (client Client) NewEventPollerWithoutEntity(entityType EntityType, action EventAction) (*EventPoller, error) {
	result := EventPoller{
		EntityType:     entityType,
		Action:         action,
		EntityID:       0,
		previousEvents: make(map[int]bool, 0),

		client: client,
	}

	return &result, nil
}

// WaitForLatestUnknownEvent waits for the next event not observed by this poller.
func (p *EventPoller) WaitForLatestUnknownEvent(ctx context.Context) (*Event, error) {
	ticker := newTicker(&p.client)
	defer ticker.Stop()

	f := Filter{
		OrderBy: "created",
		Order:   Descending,
	}
	f.AddField(Eq, "entity.type", p.EntityType)
	f.AddField(Eq, "entity.id", p.EntityID)
	f.AddField(Eq, "action", p.Action)

	fBytes, err := f.MarshalJSON()
	if err != nil {
		return nil, err
	}

	listOpts := ListOptions{
		Filter:      string(fBytes),
		PageOptions: &PageOptions{Page: 1},
	}

	for {
		select {
		case <-ticker.C:
			events, err := p.client.ListEvents(ctx, &listOpts)
			if err != nil {
				return nil, fmt.Errorf("failed to list events: %w", err)
			}

			for _, event := range events {
				if p.SecondaryEntityID != nil && !eventMatchesSecondary(p.SecondaryEntityID, event) {
					continue
				}

				if _, ok := p.previousEvents[event.ID]; !ok {
					// Store this event so it is no longer picked up
					// on subsequent jobs
					p.previousEvents[event.ID] = true

					return &event, nil
				}
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to wait for event: %w", ctx.Err())
		}
	}
}

// WaitForFinished waits for a new event to be finished.
func (p *EventPoller) WaitForFinished(ctx context.Context) (*Event, error) {
	ticker := newTicker(&p.client)
	defer ticker.Stop()

	event, err := p.WaitForLatestUnknownEvent(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for event: %w", err)
	}

	for {
		select {
		case <-ticker.C:
			event, err = p.client.GetEvent(ctx, event.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get event: %w", err)
			}

			switch event.Status {
			case EventFinished:
				return event, nil
			case EventFailed:
				return nil, fmt.Errorf("event %d has failed", event.ID)
			case EventScheduled, EventStarted, EventNotification:
				continue
			}
		case <-ctx.Done():
			return nil, fmt.Errorf("failed to wait for event finished: %w", ctx.Err())
		}
	}
}

// WaitForResourceFree waits for a resource to have no running events.
func (client Client) WaitForResourceFree(
	ctx context.Context, entityType EntityType, entityID any,
) error {
	apiFilter := Filter{
		Order:   Descending,
		OrderBy: "created",
	}
	apiFilter.AddField(Eq, "entity.id", entityID)
	apiFilter.AddField(Eq, "entity.type", entityType)

	filterStr, err := apiFilter.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to create filter: %s", err)
	}

	ticker := newTicker(&client)
	defer ticker.Stop()

	// A helper function to determine whether a resource is busy
	checkIsBusy := func(events []Event) bool {
		for _, event := range events {
			if event.Status == EventStarted || event.Status == EventScheduled {
				return true
			}
		}

		return false
	}

	for {
		select {
		case <-ticker.C:
			events, err := client.ListEvents(ctx, &ListOptions{
				Filter: string(filterStr),
			})
			if err != nil {
				return fmt.Errorf("failed to list events: %s", err)
			}

			if !checkIsBusy(events) {
				return nil
			}

		case <-ctx.Done():
			return fmt.Errorf("failed to wait for resource free: %s", ctx.Err())
		}
	}
}

// eventMatchesSecondary returns whether the given event's secondary entity
// matches the configured secondary ID.
// This logic has been broken out to improve readability.
func eventMatchesSecondary(configuredID any, e Event) bool {
	// We should return false if the event has no secondary entity.
	// e.g. A previous disk deletion has completed.
	if e.SecondaryEntity == nil {
		return false
	}

	secondaryID := e.SecondaryEntity.ID

	// Evil hack to correct IDs parsed as floats
	if value, ok := secondaryID.(float64); ok {
		secondaryID = int(value)
	}

	return secondaryID == configuredID
}

// WaitForAlertDefinitionStatus waits for the Alert Definition to reach the specified status
func (client Client) WaitForAlertDefinitionStatus(
	ctx context.Context,
	status AlertDefinitionStatus,
	serviceType string,
	alertID int,
) (*AlertDefinition, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*AlertDefinition, bool, error) {
			alertDef, err := client.GetMonitorAlertDefinition(ctx, serviceType, alertID)
			if err != nil {
				return alertDef, false, err
			}

			return alertDef, alertDef.Status == status, nil
		},
		func() error {
			return fmt.Errorf("failed to wait for AlertDefinition %d status %s: %w", alertID, status, ctx.Err())
		},
	)
}

// WaitForVolumeIOReadyStatus waits for the io_ready status to verify whether the volume is
// successfully attached to a Linode instance and ready for read and write operations
func (client Client) WaitForVolumeIOReadyStatus(
	ctx context.Context,
	volumeID int,
	status bool,
) (*Volume, error) {
	return poll(ctx, &client,
		func(ctx context.Context) (*Volume, bool, error) {
			volume, err := client.GetVolume(ctx, volumeID)
			if err != nil {
				return volume, false, fmt.Errorf("failed to get volume: %w", err)
			}

			return volume, volume.IOReady == status, nil
		},
		func() error {
			return fmt.Errorf("failed to wait for Volume %d IO Ready status %t: %w", volumeID, status, ctx.Err())
		},
	)
}

// preTask stores all current events for the given entity to prevent them from being
// processed on subsequent runs.
func (p *EventPoller) preTask(ctx context.Context) error {
	f := Filter{
		OrderBy: "created",
		Order:   Descending,
	}
	f.AddField(Eq, "entity.type", p.EntityType)
	f.AddField(Eq, "entity.id", p.EntityID)
	f.AddField(Eq, "action", p.Action)

	fBytes, err := f.MarshalJSON()
	if err != nil {
		return err
	}

	events, err := p.client.ListEvents(ctx, &ListOptions{
		Filter:      string(fBytes),
		PageOptions: &PageOptions{Page: 1},
	})
	if err != nil {
		return fmt.Errorf("failed to list events: %w", err)
	}

	eventIDs := make(map[int]bool, len(events))
	for _, event := range events {
		eventIDs[event.ID] = true
	}

	p.previousEvents = eventIDs

	return nil
}

// poll runs check on each tick until check reports done, returns an error, or ctx is canceled.
//
//nolint:ireturn // false positive: returning a generic concrete type, not an interface
func poll[T any](
	ctx context.Context,
	client *Client,
	check func(context.Context) (T, bool, error),
	timeoutErr func() error,
) (T, error) {
	ticker := newTicker(client)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			result, done, err := check(ctx)
			if err != nil {
				return result, err
			}

			if done {
				return result, nil
			}
		case <-ctx.Done():
			var zero T
			return zero, timeoutErr()
		}
	}
}

func newTicker(client *Client) *time.Ticker {
	return time.NewTicker(client.GetPollDelay())
}
