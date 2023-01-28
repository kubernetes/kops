package discovery

import (
	"context"
	"fmt"
	"time"

	"google.golang.org/protobuf/encoding/prototext"
	"k8s.io/klog/v2"
	"k8s.io/kops/pkg/kopscontrollerclient"
	"k8s.io/kops/protokube/pkg/gossip/dns/hosts"

	pb "k8s.io/kops/proto/generated/kops/kopscontroller/v1"
)

type HostsWriter struct {
	Client *kopscontrollerclient.Client
}

func (h *HostsWriter) RunForever(ctx context.Context) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := h.RunOnce(ctx); err != nil {
			klog.Warningf("error running host discovery: %v", err)
		}

		if ctx.Err() == nil {
			time.Sleep(15 * time.Second)
		}
	}
}

func (h *HostsWriter) RunOnce(ctx context.Context) error {
	req := &pb.DiscoverHostsRequest{}
	klog.Infof("Sending DiscoverHostsRequest")

	stream, err := h.Client.DiscoverHosts(ctx, req)
	if err != nil {
		return err
	}

	records := make(map[string]*pb.HostRecord)

	for {
		response, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("error receiving message from DiscoverHosts: %w", err)
		}

		klog.Infof("got discovery messsage: %v", prototext.Format(response))

		for _, record := range response.Records {
			records[record.Name] = record
		}

		if response.Complete {
			if err := h.writeEtcHosts(ctx, records); err != nil {
				return err
			}
		}

	}
}

func (h *HostsWriter) writeEtcHosts(ctx context.Context, records map[string]*pb.HostRecord) error {
	etcHostsPath := "/etc/hosts"

	mutator := func(existing []string) (*hosts.HostMap, error) {
		hostMap := &hosts.HostMap{}
		badLines := hostMap.Parse(existing)
		if len(badLines) != 0 {
			klog.Warningf("ignoring unexpected lines in /etc/hosts: %v", badLines)
		}

		for _, record := range records {
			var addresses []string
			for _, addr := range record.Addresses {
				addresses = append(addresses, addr.Address)
			}
			hostMap.ReplaceRecords(record.Name, addresses)
		}

		return hostMap, nil
	}

	if err := hosts.UpdateHostsFileWithRecords(etcHostsPath, mutator); err != nil {
		return fmt.Errorf("failed to update /etc/hosts: %w", err)
	}
	return nil
}
