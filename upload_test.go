package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"reflect"
	"testing"
	pb "upload-service/proto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var s3Endpoint string
var newSession = session.Must(session.NewSession())
var s3Client *s3.S3

func init() {
	s3AccessKey := os.Getenv("S3_ACCESS_KEY")
	s3SecretKey := os.Getenv("S3_SECRET_KEY")
	s3Endpoint = os.Getenv("S3_ENDPOINT")
	s3Token := ""

	// Configure to use S3 Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(s3AccessKey, s3SecretKey, s3Token),
		Endpoint:         aws.String(s3Endpoint),
		Region:           aws.String("eu-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
	}
	newSession = session.New(s3Config)
	s3Client = s3.New(newSession)
}

func TestUploadService_UploadFile(t *testing.T) {

	type fields struct {
		s3Client *s3.S3
	}
	type args struct {
		file     io.Reader
		key      *string
		bucket   *string
		metadata map[string]*string
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *string
		wantErr bool
	}{
		{
			name:   "upload text file",
			fields: fields{s3Client: s3Client},
			args: args{
				key:      aws.String("testfile.txt"),
				bucket:   aws.String("testbucket"),
				file:     bytes.NewReader([]byte("Hello, World!")),
				metadata: nil,
			},
			wantErr: false,
			want:    aws.String(fmt.Sprintf("%s/testbucket/testfile.txt", s3Endpoint)),
		},
		{
			name:   "upload text file in a folder",
			fields: fields{s3Client: s3Client},
			args: args{
				key:      aws.String("testfolder/testfile.txt"),
				bucket:   aws.String("testbucket"),
				file:     bytes.NewReader([]byte("Hello, World!")),
				metadata: nil,
			},
			wantErr: false,
			want:    aws.String(fmt.Sprintf("%s/testbucket/testfolder/testfile.txt", s3Endpoint)),
		},
		{
			name:   "upload text file with empty key",
			fields: fields{s3Client: s3Client},
			args: args{
				key:      aws.String(""),
				bucket:   aws.String("testbucket"),
				file:     bytes.NewReader([]byte("Hello, World!")),
				metadata: nil,
			},
			wantErr: true,
		},
		{
			name:   "upload text file with empty bucket",
			fields: fields{s3Client: s3Client},
			args: args{
				key:      aws.String("testfile.txt"),
				bucket:   aws.String(""),
				file:     bytes.NewReader([]byte("Hello, World!")),
				metadata: nil,
			},
			wantErr: true,
		},
		{
			name:   "upload text file with nil key",
			fields: fields{s3Client: s3Client},
			args: args{
				key:      nil,
				bucket:   aws.String("testbucket"),
				file:     bytes.NewReader([]byte("Hello, World!")),
				metadata: nil,
			},
			wantErr: true,
		},
		{
			name:   "upload text file with nil bucket",
			fields: fields{s3Client: s3Client},
			args: args{
				key:      aws.String("testfile.txt"),
				bucket:   nil,
				file:     bytes.NewReader([]byte("Hello, World!")),
				metadata: nil,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := UploadService{
				s3Client: tt.fields.s3Client,
			}

			got, err := s.UploadFile(tt.args.file, tt.args.metadata, tt.args.key, tt.args.bucket)
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadService.UploadFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != nil && *got != *tt.want {
				t.Errorf("UploadService.UploadFile() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestUploadHandler_UploadMedia(t *testing.T) {
	uploadservice := UploadService{
		s3Client: s3Client,
	}
	type fields struct {
		UploadService UploadService
	}
	type args struct {
		ctx     context.Context
		request *pb.UploadMediaRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *pb.UploadMediaResponse
		wantErr bool
	}{
		{
			name:   "UploadMedia - text file",
			fields: fields{UploadService: uploadservice},
			args: args{
				ctx: context.Background(),
				request: &pb.UploadMediaRequest{
					Key:    "testfile.txt",
					Bucket: "testbucket",
					File:   []byte("Hello, World!"),
				},
			},
			wantErr: false,
			want: &pb.UploadMediaResponse{
				Output: fmt.Sprintf("%s/testbucket/testfile.txt", s3Endpoint),
			},
		},
		{
			name:   "UploadMedia - text file - without key",
			fields: fields{UploadService: uploadservice},
			args: args{
				ctx: context.Background(),
				request: &pb.UploadMediaRequest{
					Key:    "",
					Bucket: "testbucket",
					File:   []byte("Hello, World!"),
				},
			},
			wantErr: true,
		},
		{
			name:   "UploadMedia - text file - without bucket",
			fields: fields{UploadService: uploadservice},
			args: args{
				ctx: context.Background(),
				request: &pb.UploadMediaRequest{
					Key:    "testfile.txt",
					Bucket: "",
					File:   []byte("Hello, World!"),
				},
			},
			wantErr: true,
		},
		{
			name:   "UploadMedia - text file - with nil file",
			fields: fields{UploadService: uploadservice},
			args: args{
				ctx: context.Background(),
				request: &pb.UploadMediaRequest{
					Key:    "testfile.txt",
					Bucket: "testbucket",
					File:   nil,
				},
			},
			wantErr: false,
			want: &pb.UploadMediaResponse{
				Output: fmt.Sprintf("%s/testbucket/testfile.txt", s3Endpoint),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := UploadHandler{
				UploadService: tt.fields.UploadService,
			}
			got, err := h.UploadMedia(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("UploadHandler.UploadMedia() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != nil && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("UploadHandler.UploadMedia() = %v, want %v", got, tt.want)
			}
		})
	}
}