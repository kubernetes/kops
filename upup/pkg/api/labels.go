package api

// AnnotationNameManagement is the annotation that indicates that a cluster is under external or non-standard management
const AnnotationNameManagement = "kops.kubernetes.io/management"

// AnnotationValueManagementImported is the annotation value that indicates a cluster was imported, typically as part of an upgrade
const AnnotationValueManagementImported = "imported"
