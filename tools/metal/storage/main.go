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

package main

import (
	"context"
	"encoding/xml"
	"flag"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/kubernetes/kops/tools/metal/dhcp/pkg/objectstore"
	"github.com/kubernetes/kops/tools/metal/dhcp/pkg/objectstore/fsobjectstore"
	"github.com/kubernetes/kops/tools/metal/dhcp/pkg/s3model"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/klog/v2"
)

func main() {
	err := run(context.Background())
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	log := klog.FromContext(ctx)

	httpListen := ""
	flag.StringVar(&httpListen, "http-listen", httpListen, "endpoint on which to serve HTTP requests")

	storageDir := ""
	flag.StringVar(&storageDir, "storage-dir", storageDir, "directory in which to store data")

	flag.Parse()

	if httpListen == "" {
		return fmt.Errorf("must specify http-listen flag")
	}

	if storageDir == "" {
		return fmt.Errorf("must specify storage-dir flag")
	}

	// store := testobjectstore.New()
	store := fsobjectstore.New(storageDir)

	s3Server := &S3Server{
		store: store,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if err := s3Server.ServeRequest(ctx, w, r); err != nil {
			code := status.Code(err)
			log.Error(err, "failed to serve request", "code", code)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}

	})

	httpServer := http.Server{
		Addr: httpListen,
	}

	log.Info("serving http", "endpoint", httpListen)
	if err := httpServer.ListenAndServe(); err != nil {
		return fmt.Errorf("serving http: %w", err)
	}

	return nil
}

type S3Server struct {
	store objectstore.ObjectStore
}

func (s *S3Server) ListAllMyBuckets(ctx context.Context, req *s3Request, r *ListAllMyBucketsInput) error {
	output := &s3model.ListAllMyBucketsResult{}

	buckets, err := s.store.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("listing buckets: %w", err)
	}

	for _, bucket := range buckets {
		output.Buckets = append(output.Buckets, s3model.Bucket{
			CreationDate: bucket.CreationDate.Format(s3TimeFormat),
			Name:         bucket.Name,
		})
	}

	return req.writeXML(ctx, output)
}

type ListAllMyBucketsInput struct {
}

func (s *S3Server) ServeRequest(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	log := klog.FromContext(ctx)
	log.Info("http request", "request.url", r.URL.String(), "request.method", r.Method)

	tokens := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")

	values, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		return fmt.Errorf("failed to parse query: %w", err)
	}

	req := &s3Request{
		w: w,
		r: r,
	}

	if len(tokens) == 1 && tokens[0] == "" {
		return s.ListAllMyBuckets(ctx, req, &ListAllMyBucketsInput{})
	}

	if len(tokens) == 1 {
		bucket := tokens[0]
		switch r.Method {
		case http.MethodGet:
			return s.ListObjectsV2(ctx, req, &ListObjectsV2Input{
				Bucket:    bucket,
				Delimiter: values.Get("delimiter"),
				Prefix:    values.Get("prefix"),
			})
		case http.MethodPut:
			return s.CreateBucket(ctx, req, &CreateBucketInput{
				Bucket: bucket,
			})

		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return nil
		}
	}

	if len(tokens) > 1 {
		bucket := tokens[0]
		key := strings.TrimPrefix(r.URL.Path, "/"+bucket+"/")
		switch r.Method {
		case http.MethodGet:
			if values.Has("acl") {
				return s.GetObjectACL(ctx, req, &GetObjectACLInput{
					Bucket: bucket,
					Key:    key,
				})
			}
			return s.GetObject(ctx, req, &GetObjectInput{
				Bucket: bucket,
				Key:    key,
			})
		case http.MethodHead:
			// GetObject can handle req.Method == HEAD
			return s.GetObject(ctx, req, &GetObjectInput{
				Bucket: bucket,
				Key:    key,
			})
		case http.MethodPut:
			return s.PutObject(ctx, req, &PutObjectInput{
				Bucket: bucket,
				Key:    key,
			})
		default:
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return nil
		}
	}

	return fmt.Errorf("unhandled path %q", r.URL.Path)
}

type ListObjectsV2Input struct {
	Bucket string

	Delimiter string
	Prefix    string
}

const s3TimeFormat = "2006-01-02T15:04:05.000Z"

func (s *S3Server) ListObjectsV2(ctx context.Context, req *s3Request, input *ListObjectsV2Input) error {
	log := klog.FromContext(ctx)

	bucket, _, err := s.store.GetBucket(ctx, input.Bucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket %q: %w", input.Bucket, err)
	}
	if bucket == nil {
		return req.writeError(ctx, http.StatusNotFound, &s3model.Error{
			Code:    "NoSuchBucket",
			Message: "The specified bucket does not exist",
		})
	}

	objects, err := bucket.ListObjects(ctx)
	if err != nil {
		return fmt.Errorf("failed to list objects in bucket %q: %w", input.Bucket, err)
	}

	output := &s3model.ListBucketResult{
		Name: input.Bucket,
	}

	prefixes := make(map[string]bool)
	for _, object := range objects {
		log.V(4).Info("found candidate object", "obj", object)
		if input.Prefix != "" && !strings.HasPrefix(object.Key, input.Prefix) {
			continue
		}
		if input.Delimiter != "" {
			afterPrefix := object.Key[len(input.Prefix):]

			tokens := strings.SplitN(afterPrefix, input.Delimiter, 2)
			if len(tokens) == 2 {
				prefixes[input.Prefix+tokens[0]+input.Delimiter] = true
				continue
			}
		}
		// TODO: support delimiter
		output.Contents = append(output.Contents, s3model.Object{
			Key:          object.Key,
			LastModified: object.LastModified.Format(s3TimeFormat),
			Size:         object.Size,
		})
	}

	if input.Delimiter != "" {
		for prefix := range prefixes {
			output.CommonPrefixes = append(output.CommonPrefixes, s3model.CommonPrefix{
				Prefix: prefix,
			})
		}
		output.Delimiter = input.Delimiter
	}
	output.Prefix = input.Prefix
	output.KeyCount = len(output.Contents)
	output.IsTruncated = false

	return req.writeXML(ctx, output)
}

type CreateBucketInput struct {
	Bucket string
}

func (s *S3Server) CreateBucket(ctx context.Context, req *s3Request, input *CreateBucketInput) error {
	log := klog.FromContext(ctx)

	bucketInfo, err := s.store.CreateBucket(ctx, input.Bucket)
	if err != nil {
		code := status.Code(err)
		log.Error(err, "failed to create bucket", "code", code)
		if status.Code(err) == codes.AlreadyExists {
			return req.writeError(ctx, http.StatusConflict, &s3model.Error{
				Code:       "BucketAlreadyExists",
				Message:    "The requested bucket name is not available. The bucket namespace is shared by all users of the system. Select a different name and try again.",
				BucketName: input.Bucket,
			})
		}
		return fmt.Errorf("failed to create bucket %q: %w", input.Bucket, err)
	}

	log.Info("bucket created", "bucket", bucketInfo)
	return req.writeEmpty200(ctx)
}

type GetObjectInput struct {
	Bucket string
	Key    string
}

func (s *S3Server) GetObject(ctx context.Context, req *s3Request, input *GetObjectInput) error {
	bucket, _, err := s.store.GetBucket(ctx, input.Bucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket %q: %w", input.Bucket, err)
	}
	if bucket == nil {
		return req.writeError(ctx, http.StatusNotFound, &s3model.Error{
			Code:    "NoSuchBucket",
			Message: "The specified bucket does not exist",
		})
	}

	object, err := bucket.GetObject(ctx, input.Key)
	if err != nil {
		return fmt.Errorf("failed to get object %q in bucket %q: %w", input.Key, input.Bucket, err)
	}

	if object == nil {
		return req.writeError(ctx, http.StatusNotFound, &s3model.Error{
			Code:    "NoSuchKey",
			Message: "The specified key does not exist.",
		})
	}

	return object.WriteTo(req.r, req.w)
}

type GetObjectACLInput struct {
	Bucket string
	Key    string
}

func (s *S3Server) GetObjectACL(ctx context.Context, req *s3Request, input *GetObjectACLInput) error {
	bucket, bucketInfo, err := s.store.GetBucket(ctx, input.Bucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket %q: %w", input.Bucket, err)
	}
	if bucket == nil {
		return req.writeError(ctx, http.StatusNotFound, &s3model.Error{
			Code:    "NoSuchBucket",
			Message: "The specified bucket does not exist",
		})
	}

	object, err := bucket.GetObject(ctx, input.Key)
	if err != nil {
		return fmt.Errorf("failed to get object %q in bucket %q: %w", input.Key, input.Bucket, err)
	}

	if object == nil {
		return req.writeError(ctx, http.StatusNotFound, &s3model.Error{
			Code:    "NoSuchKey",
			Message: "The specified key does not exist.",
		})
	}

	owner := bucketInfo.Owner

	output := &s3model.ObjectACLResult{
		Owner: &s3model.Owner{
			ID: owner,
		},
		Grants: []*s3model.Grant{
			{
				Grantee: &s3model.Grantee{
					ID:   owner,
					Type: "CanonicalUser",
				},
				Permission: "FULL_CONTROL",
			},
		},
	}

	return req.writeXML(ctx, output)
}

type PutObjectInput struct {
	Bucket string
	Key    string
}

func (s *S3Server) PutObject(ctx context.Context, req *s3Request, input *PutObjectInput) error {
	log := klog.FromContext(ctx)

	bucket, _, err := s.store.GetBucket(ctx, input.Bucket)
	if err != nil {
		return fmt.Errorf("failed to get bucket %q: %w", input.Bucket, err)
	}
	if bucket == nil {
		return req.writeError(ctx, http.StatusNotFound, &s3model.Error{
			Code:    "NoSuchBucket",
			Message: "The specified bucket does not exist",
		})
	}

	objectInfo, err := bucket.PutObject(ctx, input.Key, req.r.Body)
	if err != nil {
		return fmt.Errorf("failed to create object %q in bucket %q: %w", input.Key, input.Bucket, err)
	}
	log.Info("object created", "object", objectInfo)

	return nil
}

type s3Request struct {
	Action  string
	Version string

	w http.ResponseWriter
	r *http.Request
}

func (s *s3Request) writeXML(ctx context.Context, output any) error {
	log := klog.FromContext(ctx)

	b, err := xml.Marshal(output)
	if err != nil {
		return fmt.Errorf("failed to convert to xml: %w", err)
	}
	log.Info("writing xml response", "xml", string(b))
	s.w.Write(b)
	return nil
}

func (s *s3Request) writeEmpty200(ctx context.Context) error {
	s.w.WriteHeader(http.StatusOK)
	s.w.Write(nil)
	return nil
}

func (s *s3Request) writeError(ctx context.Context, statusCode int, error *s3model.Error) error {
	log := klog.FromContext(ctx)

	s.w.WriteHeader(statusCode)
	if error != nil {
		b, err := xml.Marshal(error)
		if err != nil {
			return fmt.Errorf("failed to convert error to xml: %w", err)
		}
		log.Info("writing xml error response", "code", statusCode, "xml", string(b))
		s.w.Write(b)
	} else {
		log.Info("writing empty error response", "code", statusCode)
	}
	return nil
}
