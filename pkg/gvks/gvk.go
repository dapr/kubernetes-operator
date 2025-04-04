package gvks

import (
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"

	apiextensions "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var CustomResourceDefinition = schema.GroupVersionKind{
	Group:   apiextensions.GroupName,
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

var Secret = schema.GroupVersionKind{
	Group:   corev1.GroupName,
	Version: "v1",
	Kind:    "Secret",
}

var Deployment = schema.GroupVersionKind{
	Group:   appsv1.GroupName,
	Version: "v1",
	Kind:    "Deployment",
}

var StatefulSet = schema.GroupVersionKind{
	Group:   appsv1.GroupName,
	Version: "v1",
	Kind:    "StatefulSet",
}

var MutatingWebhookConfiguration = schema.GroupVersionKind{
	Group:   admissionregistrationv1.GroupName,
	Version: "v1",
	Kind:    "MutatingWebhookConfiguration",
}
