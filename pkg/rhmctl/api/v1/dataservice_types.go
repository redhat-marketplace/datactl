package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type FileInfo struct {
	Id         string            `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Name       string            `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	Size       uint32            `protobuf:"varint,3,opt,name=size,proto3" json:"size,omitempty"`
	Source     string            `protobuf:"bytes,4,opt,name=source,proto3" json:"source,omitempty"`
	SourceType string            `protobuf:"bytes,5,opt,name=sourceType,proto3" json:"sourceType,omitempty"`
	Checksum   string            `protobuf:"bytes,10,opt,name=checksum,proto3" json:"checksum,omitempty"`
	MimeType   string            `protobuf:"bytes,11,opt,name=mimeType,proto3" json:"mimeType,omitempty"`
	CreatedAt  *metav1.Time      `protobuf:"bytes,15,opt,name=created_at,json=createdAt,proto3" json:"created_at,omitempty"`
	UpdatedAt  *metav1.Time      `protobuf:"bytes,16,opt,name=updated_at,json=updatedAt,proto3" json:"updated_at,omitempty"`
	DeletedAt  *metav1.Time      `protobuf:"bytes,17,opt,name=deleted_at,json=deletedAt,proto3,oneof" json:"deleted_at,omitempty"`
	Metadata   map[string]string `protobuf:"bytes,20,rep,name=metadata,proto3" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
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
