package api

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "rhmctl", Version: "__internal"}

var (
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)
	AddToScheme   = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&Config{},
		&FileInfo{},
		&ListFilesResponse{},
		&GetFileResponse{},
	)
	return nil
}

func (obj *Config) GetObjectKind() schema.ObjectKind { return obj }

func (obj *Config) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *Config) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "Config")
}

func (obj *FileInfo) GetObjectKind() schema.ObjectKind { return obj }

func (obj *FileInfo) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *FileInfo) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "FileInfo")
}

func (obj *ListFilesResponse) GetObjectKind() schema.ObjectKind { return obj }

func (obj *ListFilesResponse) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *ListFilesResponse) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "ListFileResponse")
}

func (obj *GetFileResponse) GetObjectKind() schema.ObjectKind { return obj }

func (obj *GetFileResponse) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *GetFileResponse) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "GetFileResponse")
}
