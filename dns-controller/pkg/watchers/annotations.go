package watchers

// AnnotationNameDnsExternal is used to set up a DNS name for accessing the resource from outside the cluster
// For a service of Type=LoadBalancer, it would map to the external LB hostname or IP
const AnnotationNameDnsExternal = "dns.alpha.kubernetes.io/external"

// AnnotationNameDnsInternal is used to set up a DNS name for accessing the resource from inside the cluster
// This is only supported on Pods currently, and maps to the Internal address
const AnnotationNameDnsInternal = "dns.alpha.kubernetes.io/internal"
