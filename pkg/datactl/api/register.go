package api

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "datactl", Version: "__internal"}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Config{},
	)
	return nil
}

func (obj *Config) GetObjectKind() schema.ObjectKind { return obj }

func (obj *Config) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *Config) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "Config")
}