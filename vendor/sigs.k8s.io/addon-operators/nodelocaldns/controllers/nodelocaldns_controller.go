package controllers

import (
	"context"
	"fmt"
	"net"
	"strings"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/klog"
	api "sigs.k8s.io/addon-operators/nodelocaldns/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/addon/pkg/status"
	"sigs.k8s.io/kubebuilder-declarative-pattern/pkg/patterns/declarative"
)

var _ reconcile.Reconciler = &NodeLocalDNSReconciler{}

// NodeLocalDNSReconciler reconciles a NodeLocalDNS object
type NodeLocalDNSReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	declarative.Reconciler
}

// +kubebuilder:rbac:groups=addons.k8s.io,resources=nodelocaldns,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=addons.k8s.io,resources=nodelocaldns/status,verbs=get;update;patch

func (r *NodeLocalDNSReconciler) SetupWithManager(mgr ctrl.Manager) error {
	addon.Init()

	labels := map[string]string{
		"k8s-app": "nodelocaldns",
	}

	watchLabels := declarative.SourceLabel(mgr.GetScheme())

	replacePlaceholders := func(ctx context.Context, object declarative.DeclarativeObject, s string) (string, error) {
		// TODO: Should we default and if so where?
		dnsDomain := "" // o.Spec.DNSDomain
		if dnsDomain == "" {
			dnsDomain = "cluster.local"
		}

		dnsServerIP := "" // o.Spec.DNSServerIP
		if dnsServerIP == "" {
			ip, err := findServiceClusterIP(ctx, mgr.GetClient(), "kube-system", "kube-dns")
			if err != nil {
				return "", fmt.Errorf("unable to find kube-dns IP: %v", err)
			}
			dnsServerIP = ip.String()
		}

		localDNS := "169.254.20.10"

		s = strings.Replace(s, "__PILLAR__DNS__DOMAIN__", dnsDomain, -1)
		s = strings.Replace(s, "__PILLAR__DNS__SERVER__", dnsServerIP, -1)
		s = strings.Replace(s, "__PILLAR__LOCAL__DNS__", localDNS, -1)

		return s, nil
	}

	if err := r.Reconciler.Init(mgr, &api.NodeLocalDNS{},
		declarative.WithObjectTransform(declarative.AddLabels(labels)),
		declarative.WithOwner(declarative.SourceAsOwner),
		declarative.WithLabels(watchLabels),
		declarative.WithStatus(status.NewBasic(mgr.GetClient())),
		// TODO: add an application to your manifest:  declarative.WithObjectTransform(addon.TransformApplicationFromStatus),
		// TODO: add an application to your manifest:  declarative.WithManagedApplication(watchLabels),
		declarative.WithRawManifestOperation(replacePlaceholders),
		declarative.WithObjectTransform(addon.ApplyPatches),
	); err != nil {
		return err
	}

	c, err := controller.New("nodelocaldns-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to NodeLocalDNS
	err = c.Watch(&source.Kind{Type: &api.NodeLocalDNS{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to deployed objects
	_, err = declarative.WatchAll(mgr.GetConfig(), c, r, watchLabels)
	if err != nil {
		return err
	}

	return nil
}

func findServiceClusterIP(ctx context.Context, c client.Client, namespace string, name string) (net.IP, error) {
	key := types.NamespacedName{Namespace: namespace, Name: name}

	service := &corev1.Service{}
	if err := c.Get(ctx, key, service); err != nil {
		return nil, fmt.Errorf("error getting service %s: %v", key, err)
	}

	ip := net.ParseIP(service.Spec.ClusterIP)
	if ip == nil {
		return nil, fmt.Errorf("cannot parse service %s ClusterIP %q", key, service.Spec.ClusterIP)
	}

	klog.Infof("got ClusterIP for %s: %q", key, ip)
	return ip, nil
}
