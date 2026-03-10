package r2

import (
	"KnowLedger/internal/model"
	"KnowLedger/internal/storage"
	"KnowLedger/pkg/utils"
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/gabriel-vasile/mimetype"
)

type R2CASStorage struct {
	client    *s3.Client
	bucket    string
	publicUrl string
}

func NewR2CASStorage(bucketName, accessKey, secretKey, endpoint, publicUrl string) (*R2CASStorage, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2CASStorage{
		client:    client,
		bucket:    bucketName,
		publicUrl: publicUrl,
	}, nil
}

func (r *R2CASStorage) Upload(ctx context.Context, data []byte) (*model.MediaItem, error) {
	contentType := mimetype.Detect(data)
	if contentType.Is("application/octet-stream") {
		return nil, errors.New("unsupported or unrecognized file type")
	}

	hash := utils.SHA256Hash(data)
	ext := contentType.Extension()
	key := utils.KeyFromHash(hash) + ext
	size := int64(len(data))
	ctType := contentType.String()

	exists, err := r.Exists(ctx, key)
	if err != nil {
		return nil, err
	}
	if exists {
		return &model.MediaItem{
			Key:         key,
			Hash:        hash,
			Size:        size,
			ContentType: ctType,
			URL:         r.GetURL(ctx, key),
		}, nil
	}

	_, err = r.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType.String()),
		Metadata: map[string]string{
			"sha256": hash,
		},
	})

	if err != nil {
		return nil, err
	}

	return &model.MediaItem{
		Key:         key,
		Hash:        hash,
		Size:        size,
		ContentType: ctType,
		URL:         r.GetURL(ctx, key),
	}, nil
}

func (r *R2CASStorage) Delete(ctx context.Context, key string) error {
	_, err := r.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	return err
}

func (r *R2CASStorage) Exists(ctx context.Context, key string) (bool, error) {
	_, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})

	if err != nil {

		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func (r *R2CASStorage) GetURL(ctx context.Context, key string) string {
	return r.publicUrl + key
}

func (r *R2CASStorage) GetDetails(ctx context.Context, key string) (*model.MediaItem, error) {
	output, err := r.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var notFound *types.NotFound
		if errors.As(err, &notFound) {
			return nil, fmt.Errorf("object %s not found", key)
		}
		return nil, err
	}

	var contentType string
	if output.ContentType != nil {
		contentType = *output.ContentType
	}

	var size int64
	if output.ContentLength != nil {
		size = *output.ContentLength
	}

	hash := output.Metadata["sha256"]

	return &model.MediaItem{
		Key:         key,
		Hash:        hash,
		Size:        size,
		ContentType: contentType,
		URL:         r.GetURL(ctx, key),
	}, nil
}

func (r *R2CASStorage) DeleteBatch(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	const maxBatchSize = 1000

	for i := 0; i < len(keys); i += maxBatchSize {
		end := i + maxBatchSize
		if end > len(keys) {
			end = len(keys)
		}

		chunk := keys[i:end]
		objectIds := make([]types.ObjectIdentifier, len(chunk))
		for j, key := range chunk {
			objectIds[j] = types.ObjectIdentifier{
				Key: aws.String(key),
			}
		}

		output, err := r.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
			Bucket: aws.String(r.bucket),
			Delete: &types.Delete{
				Objects: objectIds,
				Quiet:   aws.Bool(true), // hanya return error, bukan success list
			},
		})
		if err != nil {
			return fmt.Errorf("batch delete failed at chunk %d: %w", i/maxBatchSize, err)
		}

		if len(output.Errors) > 0 {
			var errMsgs []string
			for _, e := range output.Errors {
				errMsgs = append(errMsgs, fmt.Sprintf("key=%s code=%s message=%s",
					aws.ToString(e.Key),
					aws.ToString(e.Code),
					aws.ToString(e.Message),
				))
			}
			return fmt.Errorf("partial delete errors: %s", strings.Join(errMsgs, "; "))
		}
	}

	return nil
}

func (r *R2CASStorage) ScanAll(ctx context.Context, fn func(item storage.ScanResult) error) error {
	var continuationToken *string

	for {
		output, err := r.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
			Bucket:            aws.String(r.bucket),
			ContinuationToken: continuationToken,
		})
		if err != nil {
			return fmt.Errorf("list objects failed: %w", err)
		}

		for _, obj := range output.Contents {
			result := storage.ScanResult{
				Key:          aws.ToString(obj.Key),
				Size:         aws.ToInt64(obj.Size),
				ETag:         aws.ToString(obj.ETag),
				LastModified: aws.ToTime(obj.LastModified),
			}

			if err := fn(result); err != nil {
				return err
			}
		}

		if !aws.ToBool(output.IsTruncated) {
			break
		}
		continuationToken = output.NextContinuationToken
	}

	return nil
}
