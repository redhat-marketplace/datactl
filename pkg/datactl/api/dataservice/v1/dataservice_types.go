// Copyright 2021 IBM Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package v1

import (
	"github.com/fatih/color"
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
	Action `json:"action,omitempty"`

	// +optional
	Result `json:"lastActionResult,omitempty"`

	// +optional
	Error string `protobuf:"-" json:"error,omitempty"`

	// +optional
	UploadID string `protobuf:"-" json:"uploadID,omitempty"`

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

type Result string

const (
	Ok     Result = "Ok"
	Error  Result = "Err"
	DryRun Result = "DryRun"
)

var (
	red    = color.New(color.FgRed)
	green  = color.New(color.FgGreen)
	yellow = color.New(color.FgYellow)
)

func (a Result) MarshalText() ([]byte, error) {
	switch a {
	case Ok:
		return []byte(green.Sprint(a)), nil
	case Error:
		return []byte(red.Sprint(a)), nil
	case DryRun:
		return []byte(yellow.Sprint(a)), nil
	default:
		return []byte(string(a)), nil
	}
}

func (a *Result) UnmarshalText(text []byte) error {
	*a = Result(string(text))

	switch *a {
	case Ok:
		fallthrough
	case Error:
		return nil
	default:
		return nil
	}
}

type Action string

const (
	Pull   Action = "Pull"
	Push   Action = "Push"
	Commit Action = "Commit"
)

func (a Action) MarshalText() ([]byte, error) {
	return []byte(string(a)), nil
}

func (a *Action) UnmarshalText(text []byte) error {
	*a = Action(string(text))

	switch *a {
	case Commit:
		fallthrough
	case Pull:
		fallthrough
	case Push:
		return nil
	default:
		return nil
	}
}
