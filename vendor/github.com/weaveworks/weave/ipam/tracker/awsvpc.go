package tracker

// The AWSVPC tracker tracks the IPAM ring changes and accordingly updates
// the AWS VPC route table and, if the announced change is local
// (denoted by local=true) to the host, the host route table.
//
// During the initialization, the tracker detects AWS VPC route table id which
// is associated with the host's subnet on the AWS network. If such a table does
// not exist, the default VPC route table is used.
//
// When a host A donates a range to a host B, the necessary route table
// updates (removal) happen on the host A first, and afterwards on
// the host B (installation).
//
// NB: there is a hard limit for 50 routes within any VPC route table
// (practically, it is 48, because one route is used by the AWS Internet GW and
// one by the AWS host subnet), therefore it is suggested to avoid
// an excessive fragmentation within the IPAM ring which might happen due to
// the claim operations or uneven distribution of containers across the hosts.

import (
	"fmt"
	"net"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/ec2metadata"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/vishvananda/netlink"

	"github.com/weaveworks/weave/common"
	wnet "github.com/weaveworks/weave/net"
	"github.com/weaveworks/weave/net/address"
)

type AWSVPCTracker struct {
	ec2          *ec2.EC2
	instanceID   string // EC2 Instance ID
	routeTableID string // VPC Route Table ID
	linkIndex    int    // The weave bridge link index
}

// NewAWSVPCTracker creates and initialises AWS VPC based tracker.
func NewAWSVPCTracker() (*AWSVPCTracker, error) {
	var (
		err     error
		session = session.New()
		t       = &AWSVPCTracker{}
	)

	// Detect region and instance id
	meta := ec2metadata.New(session)
	t.instanceID, err = meta.GetMetadata("instance-id")
	if err != nil {
		return nil, fmt.Errorf("cannot detect instance-id: %s", err)
	}
	region, err := meta.Region()
	if err != nil {
		return nil, fmt.Errorf("cannot detect region: %s", err)
	}

	t.ec2 = ec2.New(session, aws.NewConfig().WithRegion(region))

	routeTableID, err := t.detectRouteTableID()
	if err != nil {
		return nil, err
	}
	t.routeTableID = *routeTableID

	// Detect Weave bridge link index
	link, err := netlink.LinkByName(wnet.WeaveBridgeName)
	if err != nil {
		return nil, fmt.Errorf("cannot find \"%s\" interface: %s", wnet.WeaveBridgeName, err)
	}
	t.linkIndex = link.Attrs().Index

	t.infof("AWSVPC has been initialized on %s instance for %s route table at %s region",
		t.instanceID, t.routeTableID, region)

	return t, nil
}

// HandleUpdate method updates the AWS VPC and the host route tables.
func (t *AWSVPCTracker) HandleUpdate(prevRanges, currRanges []address.Range, local bool) error {
	t.debugf("replacing %q by %q; local(%t)", prevRanges, currRanges, local)

	prev, curr := removeCommon(address.NewCIDRs(merge(prevRanges)), address.NewCIDRs(merge(currRanges)))

	// It might make sense to do the removal first and then add entries
	// because of the 50 routes limit. However, in such case a container might
	// not be reachable for a short period of time which is not a desired behavior.

	// Add new entries
	for _, cidr := range curr {
		cidrStr := cidr.String()
		t.debugf("adding route %s to %s", cidrStr, t.instanceID)
		if _, err := t.createVPCRoute(cidrStr); err != nil {
			return fmt.Errorf("createVPCRoutes failed: %s", err)
		}
		if local {
			if err := t.createHostRoute(cidrStr); err != nil {
				if errno, ok := err.(syscall.Errno); !(ok && errno == syscall.EEXIST) {
					return fmt.Errorf("createHostRoute failed: %s", err)
				}
			}
		}
	}

	// Remove obsolete entries
	for _, cidr := range prev {
		cidrStr := cidr.String()
		t.debugf("removing %s route", cidrStr)
		if _, err := t.deleteVPCRoute(cidrStr); err != nil {
			return fmt.Errorf("deleteVPCRoute failed: %s", err)
		}
		if local {
			if err := t.deleteHostRoute(cidrStr); err != nil {
				return fmt.Errorf("deleteHostRoute failed: %s", err)
			}
		}
	}

	return nil
}

func (t *AWSVPCTracker) createVPCRoute(cidr string) (*ec2.CreateRouteOutput, error) {
	route := &ec2.CreateRouteInput{
		RouteTableId:         &t.routeTableID,
		InstanceId:           &t.instanceID,
		DestinationCidrBlock: &cidr,
	}
	return t.ec2.CreateRoute(route)
}

func (t *AWSVPCTracker) createHostRoute(cidr string) error {
	dst, err := parseCIDR(cidr)
	if err != nil {
		return err
	}
	route := &netlink.Route{
		LinkIndex: t.linkIndex,
		Dst:       dst,
		Scope:     netlink.SCOPE_LINK,
	}
	return netlink.RouteAdd(route)
}

func (t *AWSVPCTracker) deleteVPCRoute(cidr string) (*ec2.DeleteRouteOutput, error) {
	route := &ec2.DeleteRouteInput{
		RouteTableId:         &t.routeTableID,
		DestinationCidrBlock: &cidr,
	}
	return t.ec2.DeleteRoute(route)
}

func (t *AWSVPCTracker) deleteHostRoute(cidr string) error {
	dst, err := parseCIDR(cidr)
	if err != nil {
		return err
	}
	route := &netlink.Route{
		LinkIndex: t.linkIndex,
		Dst:       dst,
		Scope:     netlink.SCOPE_LINK,
	}
	return netlink.RouteDel(route)
}

// detectRouteTableID detects AWS VPC Route Table ID of the given tracker instance.
func (t *AWSVPCTracker) detectRouteTableID() (*string, error) {
	instancesParams := &ec2.DescribeInstancesInput{
		InstanceIds: []*string{aws.String(t.instanceID)},
	}
	instancesResp, err := t.ec2.DescribeInstances(instancesParams)
	if err != nil {
		return nil, fmt.Errorf("DescribeInstances failed: %s", err)
	}
	if len(instancesResp.Reservations) == 0 ||
		len(instancesResp.Reservations[0].Instances) == 0 {
		return nil, fmt.Errorf("cannot find %s instance within reservations", t.instanceID)
	}
	vpcID := instancesResp.Reservations[0].Instances[0].VpcId
	subnetID := instancesResp.Reservations[0].Instances[0].SubnetId

	// First try to find a routing table associated with the subnet of the instance
	tablesParams := &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("association.subnet-id"),
				Values: []*string{subnetID},
			},
		},
	}
	tablesResp, err := t.ec2.DescribeRouteTables(tablesParams)
	if err != nil {
		return nil, fmt.Errorf("DescribeRouteTables failed: %s", err)
	}
	if len(tablesResp.RouteTables) != 0 {
		return tablesResp.RouteTables[0].RouteTableId, nil
	}
	// Fallback to the default routing table
	tablesParams = &ec2.DescribeRouteTablesInput{
		Filters: []*ec2.Filter{
			{
				Name:   aws.String("association.main"),
				Values: []*string{aws.String("true")},
			},
			{
				Name:   aws.String("vpc-id"),
				Values: []*string{vpcID},
			},
		},
	}
	tablesResp, err = t.ec2.DescribeRouteTables(tablesParams)
	if err != nil {
		return nil, fmt.Errorf("DescribeRouteTables failed: %s", err)
	}
	if len(tablesResp.RouteTables) != 0 {
		return tablesResp.RouteTables[0].RouteTableId, nil
	}

	return nil, fmt.Errorf("cannot find routetable for %s instance", t.instanceID)
}

func (t *AWSVPCTracker) debugf(fmt string, args ...interface{}) {
	common.Log.Debugf("[tracker] "+fmt, args...)
}

func (t *AWSVPCTracker) infof(fmt string, args ...interface{}) {
	common.Log.Infof("[tracker] "+fmt, args...)
}

// Helpers

// merge merges adjacent range entries.
// The given slice has to be sorted in increasing order.
func merge(r []address.Range) []address.Range {
	var merged []address.Range

	for i := range r {
		if prev := len(merged) - 1; prev >= 0 && merged[prev].End == r[i].Start {
			merged[prev].End = r[i].End
		} else {
			merged = append(merged, r[i])
		}
	}

	return merged
}

// removeCommon filters out CIDR ranges which are contained in both a and b slices.
// Both slices have to be sorted in increasing order.
func removeCommon(a, b []address.CIDR) (newA, newB []address.CIDR) {
	i, j := 0, 0

	for i < len(a) && j < len(b) {
		switch {
		case a[i].Start() < b[j].Start() || a[i].End() < b[j].End():
			newA = append(newA, a[i])
			i++
		case a[i].Start() > b[j].Start() || a[i].End() > b[j].End():
			newB = append(newB, b[j])
			j++
		default:
			i++
			j++
		}

	}
	newA = append(newA, a[i:]...)
	newB = append(newB, b[j:]...)

	return
}

func parseCIDR(cidr string) (*net.IPNet, error) {
	ip, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return nil, err
	}
	ipnet.IP = ip

	return ipnet, nil
}
