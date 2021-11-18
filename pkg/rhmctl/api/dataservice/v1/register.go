package v1

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// SchemeGroupVersion is group version used to register these objects
var SchemeGroupVersion = schema.GroupVersion{Group: "rhmctl", Version: "v1"}

var (
	// TODO: move SchemeBuilder with zz_generated.deepcopy.go to k8s.io/api.
	// localSchemeBuilder and AddToScheme will stay in k8s.io/kubernetes.
	SchemeBuilder      runtime.SchemeBuilder
	localSchemeBuilder = &SchemeBuilder
	AddToScheme        = localSchemeBuilder.AddToScheme
)

func init() {
	// We only register manually written functions here. The registration of the
	// generated functions takes place in the generated files. The separation
	// makes the code compile even when the generated files are missing.
	localSchemeBuilder.Register(addKnownTypes)
}

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(SchemeGroupVersion,
		&ListFilesResponse{},
		&GetFileResponse{},
		&FileInfo{},
		&FileInfoCTLAction{},
	)
	return nil
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
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "ListFilesResponse")
}

func (obj *GetFileResponse) GetObjectKind() schema.ObjectKind { return obj }

func (obj *GetFileResponse) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *GetFileResponse) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "GetFileResponse")
}

func (obj *FileInfoCTLAction) GetObjectKind() schema.ObjectKind { return obj }

func (obj *FileInfoCTLAction) SetGroupVersionKind(gvk schema.GroupVersionKind) {
}

func (obj *FileInfoCTLAction) GroupVersionKind() schema.GroupVersionKind {
	return schema.FromAPIVersionAndKind(SchemeGroupVersion.Group, "CTLAction")
}