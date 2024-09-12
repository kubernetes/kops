/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package s3model

type Error struct {
	Code       string `xml:"Code"`
	Message    string `xml:"Message"`
	BucketName string `xml:"BucketName"`
	RequestId  string `xml:"RequestId"`
	HostId     string `xml:"HostId"`
}

type ListAllMyBucketsResult struct {
	Buckets []Bucket `xml:"Buckets>Bucket"`
}

type Bucket struct {
	Name         string `xml:"Name"`
	CreationDate string `xml:"CreationDate"`
}

type ListBucketResult struct {
	IsTruncated           bool           `xml:"IsTruncated"`
	Contents              []Object       `xml:"Contents"`
	Name                  string         `xml:"Name"`
	Prefix                string         `xml:"Prefix"`
	Delimiter             string         `xml:"Delimiter"`
	MaxKeys               int            `xml:"MaxKeys"`
	CommonPrefixes        []CommonPrefix `xml:"CommonPrefixes"`
	EncodingType          string         `xml:"EncodingType"`
	KeyCount              int            `xml:"KeyCount"`
	ContinuationToken     string         `xml:"ContinuationToken"`
	NextContinuationToken string         `xml:"NextContinuationToken"`
	StartAfter            string         `xml:"StartAfter"`
}

type CommonPrefix struct {
	Prefix string `xml:"Prefix"`
}

type Object struct {
	ChecksumAlgorithm string         `xml:"ChecksumAlgorithm"`
	ETag              string         `xml:"ETag"`
	Key               string         `xml:"Key"`
	LastModified      string         `xml:"LastModified"`
	Owner             *Owner         `xml:"Owner"`
	RestoreStatus     *RestoreStatus `xml:"RestoreStatus"`
	Size              int64          `xml:"Size"`
	StorageClass      string         `xml:"StorageClass"`
}
type Owner struct {
	DisplayName string `xml:"DisplayName"`
	ID          string `xml:"ID"`
}

type RestoreStatus struct {
	IsRestoreInProgress bool    `xml:"IsRestoreInProgress"`
	RestoreExpiryDate   *string `xml:"RestoreExpiryDate"`
}

type ObjectACLResult struct {
	Owner  *Owner   `xml:"Owner"`
	Grants []*Grant `xml:"Grant"`
}

type Grant struct {
	Grantee    *Grantee `xml:"Grantee"`
	Permission string   `xml:"Permission"`
}

type Grantee struct {
	ID   string `xml:"ID"`
	Type string `xml:"Type"`
}
