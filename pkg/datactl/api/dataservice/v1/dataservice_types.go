package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FileInfo struct {
	metav1.ObjectMeta `json:",inline"`

	// +optional
	Id string `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	// +optional
	Size uint32 `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	// +optional
	Source string `protobuf:"bytes,4,opt,name=source,proto3" json:"source,omitempty"`
	// +optional
	SourceType string `protobuf:"bytes,5,opt,name=sourceType,proto3" json:"sourceType,omitempty"`
	// +optional
	Checksum string `protobuf:"bytes,10,opt,name=checksum,proto3" json:"checksum,omitempty"`
	// +optional
	MimeType string `protobuf:"bytes,11,opt,name=mimeType,proto3" json:"mimeType,omitempty"`
	// +optional
	CreatedAt *metav1.Time `protobuf:"bytes,15,opt,name=created_at,json=createdAt,proto3" json:"createdAt,omitempty"`
	// +optional
	UpdatedAt *metav1.Time `protobuf:"bytes,16,opt,name=updated_at,json=updatedAt,proto3" json:"updatedAt,omitempty"`
	// +optional
	DeletedAt *metav1.Time `protobuf:"bytes,17,opt,name=deleted_at,json=deletedAt,proto3,oneof" json:"deletedAt,omitempty"`
	// +optional
	Metadata map[string]string `protobuf:"bytes,20,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type ListFilesResponse struct {
	// The field name should match the noun "files" in the method name.  There
	// will be a maximum number of items returned based on the page_size field
	// in the request.
	Files []*FileInfo `protobuf:"bytes,1,rep,name=files,proto3" json:"files,omitempty"`

	// Token to retrieve the next page of results, or empty if there are no
	// more results in the list.
	NextPageToken string `protobuf:"bytes,2,opt,name=next_page_token,json=nextPageToken,proto3" json:"next_page_token,omitempty"`

	// The maximum number of items to return.
	PageSize int32 `protobuf:"varint,3,opt,name=page_size,json=pageSize,proto3" json:"page_size,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type GetFileResponse struct {
	Info *FileInfo `protobuf:"bytes,1,opt,name=info,proto3" json:"info,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FileInfoCTLAction struct {
	*FileInfo `json:",inline"`

	// +optional
	Action string `json:"action,omitempty"`

	// +optional
	UploadID string `protobuf:"-" json:"uploadID,omitempty"`

	// +optional
	Error string `protobuf:"-" json:"error,omitempty"`

	// +optional
	UploadError string `protobuf:"-" json:"uploadError,omitempty"`

	// +optional
	Pushed bool `protobuf:"-" json:"pushed,omitempty"`

	// +optional
	Committed bool `protobuf:"-" json:"committed,omitempty"`
}

func NewFileInfoCTLAction(info *FileInfo) *FileInfoCTLAction {
	if info.CreatedAt != nil {
		info.CreationTimestamp = *info.CreatedAt
	}
	info.DeletionTimestamp = info.DeletedAt

	return &FileInfoCTLAction{
		FileInfo: info,
	}
}
