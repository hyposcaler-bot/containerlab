// Copyright 2020 Nokia
// Licensed under the BSD 3-Clause License.
// SPDX-License-Identifier: BSD-3-Clause

package utils

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFilenameForURL(t *testing.T) {
	type args struct {
		rawUrl string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "regular filename url",
			args: args{
				rawUrl: "http://myserver.foo/download/raw/node1.cfg",
			},
			want: "node1.cfg",
		},
		{
			name: "folder URL",
			args: args{
				rawUrl: "http://myserver.foo/download/raw/",
			},
			want: "raw",
		},
		{
			name: "with get parameters",
			args: args{
				rawUrl: "http://myserver.foo/download/raw/node1.cfg?foo=bar&bar=foo",
			},
			want: "node1.cfg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := FilenameForURL(tt.args.rawUrl); got != tt.want {
				t.Errorf("got: %v, want: %v", got, tt.want)
			}
		})
	}
}

func TestFileLines(t *testing.T) {
	type args struct {
		path       string
		commentStr string
	}
	tests := []struct {
		name    string
		args    args
		want    []string
		wantErr bool
	}{
		{
			name: "regular file",
			args: args{
				path:       "test_data/keys1.txt",
				commentStr: "#",
			},
			want:    []string{"valid line", "another valid line"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileLines(tt.args.path, tt.args.commentStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileLines() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if d := cmp.Diff(got, tt.want); d != "" {
				t.Errorf("FileLines() diff = %s", d)
			}
		})
	}
}

func TestIsHttpURL(t *testing.T) {
	tests := []struct {
		name            string
		url             string
		allowSchemaless bool
		want            bool
	}{
		{
			name:            "Valid HTTP URL",
			url:             "http://example.com",
			allowSchemaless: false,
			want:            true,
		},
		{
			name:            "Valid HTTPS URL",
			url:             "https://example.com",
			allowSchemaless: false,
			want:            true,
		},
		{
			name:            "Valid URL without schema",
			url:             "srlinux.dev/clab-srl",
			allowSchemaless: true,
			want:            true,
		},
		{
			name:            "Valid URL without schema and schemaless not allowed",
			url:             "srlinux.dev/clab-srl",
			allowSchemaless: false,
			want:            false,
		},
		{
			name:            "Invalid URL",
			url:             "/foo/bar",
			allowSchemaless: false,
			want:            false,
		},
		{
			name:            "stdin symbol '-'",
			url:             "-",
			allowSchemaless: false,
			want:            false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsHttpURL(tt.url, tt.allowSchemaless); got != tt.want {
				t.Errorf("IsHttpUri() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsS3URL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Valid S3 URL",
			url:  "s3://bucket/key/to/file.yaml",
			want: true,
		},
		{
			name: "Valid S3 URL with subdirectories",
			url:  "s3://my-bucket/path/to/deep/file.cfg",
			want: true,
		},
		{
			name: "HTTP URL should not match",
			url:  "https://example.com/file.yaml",
			want: false,
		},
		{
			name: "Local file path should not match",
			url:  "/path/to/file.yaml",
			want: false,
		},
		{
			name: "Empty string should not match",
			url:  "",
			want: false,
		},
		{
			name: "S3 without bucket/key should match",
			url:  "s3://",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsS3URL(tt.url); got != tt.want {
				t.Errorf("IsS3URL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseS3URL(t *testing.T) {
	tests := []struct {
		name       string
		s3URL      string
		wantBucket string
		wantKey    string
		wantErr    bool
	}{
		{
			name:       "Valid S3 URL",
			s3URL:      "s3://my-bucket/path/to/file.yaml",
			wantBucket: "my-bucket",
			wantKey:    "path/to/file.yaml",
			wantErr:    false,
		},
		{
			name:       "Valid S3 URL with single file",
			s3URL:      "s3://bucket/file.cfg",
			wantBucket: "bucket",
			wantKey:    "file.cfg",
			wantErr:    false,
		},
		{
			name:       "Invalid - not an S3 URL",
			s3URL:      "https://example.com/file",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing bucket",
			s3URL:      "s3:///file.yaml",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing key",
			s3URL:      "s3://bucket/",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing both bucket and key",
			s3URL:      "s3://",
			wantBucket: "",
			wantKey:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBucket, gotKey, err := ParseS3URL(tt.s3URL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseS3URL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBucket != tt.wantBucket {
				t.Errorf("ParseS3URL() gotBucket = %v, want %v", gotBucket, tt.wantBucket)
			}
			if gotKey != tt.wantKey {
				t.Errorf("ParseS3URL() gotKey = %v, want %v", gotKey, tt.wantKey)
			}
		})
	}
}

func TestIsGCSURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Valid GCS URL",
			url:  "gs://my-bucket/path/to/file.txt",
			want: true,
		},
		{
			name: "Valid GCS URL with subdirectories",
			url:  "gs://bucket/dir1/dir2/file.conf",
			want: true,
		},
		{
			name: "HTTP URL should not match",
			url:  "http://example.com/file.txt",
			want: false,
		},
		{
			name: "S3 URL should not match",
			url:  "s3://bucket/file.txt",
			want: false,
		},
		{
			name: "Local file path should not match",
			url:  "/home/user/file.txt",
			want: false,
		},
		{
			name: "Empty string should not match",
			url:  "",
			want: false,
		},
		{
			name: "GS without bucket/object should match",
			url:  "gs://",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsGCSURL(tt.url); got != tt.want {
				t.Errorf("IsGCSURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsAzureBlobURL(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "Valid Azure Blob URL",
			url:  "azblob://my-container/path/to/file.txt",
			want: true,
		},
		{
			name: "Valid Azure Blob URL with subdirectories",
			url:  "azblob://container/dir1/dir2/file.conf",
			want: true,
		},
		{
			name: "HTTP URL should not match",
			url:  "http://example.com/file.txt",
			want: false,
		},
		{
			name: "S3 URL should not match",
			url:  "s3://bucket/file.txt",
			want: false,
		},
		{
			name: "GCS URL should not match",
			url:  "gs://bucket/file.txt",
			want: false,
		},
		{
			name: "Local file path should not match",
			url:  "/home/user/file.txt",
			want: false,
		},
		{
			name: "Empty string should not match",
			url:  "",
			want: false,
		},
		{
			name: "Azure without container/blob should match",
			url:  "azblob://",
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsAzureBlobURL(tt.url); got != tt.want {
				t.Errorf("IsAzureBlobURL() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseGCSURL(t *testing.T) {
	tests := []struct {
		name       string
		gcsURL     string
		wantBucket string
		wantObject string
		wantErr    bool
	}{
		{
			name:       "Valid GCS URL",
			gcsURL:     "gs://my-bucket/path/to/file.txt",
			wantBucket: "my-bucket",
			wantObject: "path/to/file.txt",
			wantErr:    false,
		},
		{
			name:       "Valid GCS URL with single file",
			gcsURL:     "gs://bucket/file.conf",
			wantBucket: "bucket",
			wantObject: "file.conf",
			wantErr:    false,
		},
		{
			name:       "Invalid - not a GCS URL",
			gcsURL:     "http://example.com/file.txt",
			wantBucket: "",
			wantObject: "",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing bucket",
			gcsURL:     "gs:///path/to/file.txt",
			wantBucket: "",
			wantObject: "",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing object",
			gcsURL:     "gs://bucket/",
			wantBucket: "",
			wantObject: "",
			wantErr:    true,
		},
		{
			name:       "Invalid - missing both bucket and object",
			gcsURL:     "gs://",
			wantBucket: "",
			wantObject: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotBucket, gotObject, err := ParseGCSURL(tt.gcsURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseGCSURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotBucket != tt.wantBucket {
				t.Errorf("ParseGCSURL() gotBucket = %v, want %v", gotBucket, tt.wantBucket)
			}
			if gotObject != tt.wantObject {
				t.Errorf("ParseGCSURL() gotObject = %v, want %v", gotObject, tt.wantObject)
			}
		})
	}
}

func TestParseAzureBlobURL(t *testing.T) {
	tests := []struct {
		name          string
		azureURL      string
		wantContainer string
		wantBlob      string
		wantErr       bool
	}{
		{
			name:          "Valid Azure Blob URL",
			azureURL:      "azblob://my-container/path/to/file.txt",
			wantContainer: "my-container",
			wantBlob:      "path/to/file.txt",
			wantErr:       false,
		},
		{
			name:          "Valid Azure Blob URL with single file",
			azureURL:      "azblob://container/file.conf",
			wantContainer: "container",
			wantBlob:      "file.conf",
			wantErr:       false,
		},
		{
			name:          "Invalid - not an Azure Blob URL",
			azureURL:      "http://example.com/file.txt",
			wantContainer: "",
			wantBlob:      "",
			wantErr:       true,
		},
		{
			name:          "Invalid - missing container",
			azureURL:      "azblob:///path/to/file.txt",
			wantContainer: "",
			wantBlob:      "",
			wantErr:       true,
		},
		{
			name:          "Invalid - missing blob",
			azureURL:      "azblob://container/",
			wantContainer: "",
			wantBlob:      "",
			wantErr:       true,
		},
		{
			name:          "Invalid - missing both container and blob",
			azureURL:      "azblob://",
			wantContainer: "",
			wantBlob:      "",
			wantErr:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotContainer, gotBlob, err := ParseAzureBlobURL(tt.azureURL)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseAzureBlobURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotContainer != tt.wantContainer {
				t.Errorf("ParseAzureBlobURL() gotContainer = %v, want %v", gotContainer, tt.wantContainer)
			}
			if gotBlob != tt.wantBlob {
				t.Errorf("ParseAzureBlobURL() gotBlob = %v, want %v", gotBlob, tt.wantBlob)
			}
		})
	}
}
