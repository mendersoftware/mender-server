// Copyright 2023 Northern.tech AS
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//        http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package azblob

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/mendersoftware/mender-server/services/deployments/model"
	"github.com/mendersoftware/mender-server/services/deployments/storage"
	"github.com/mendersoftware/mender-server/services/deployments/utils"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/bloberror"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/blockblob"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/container"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob/sas"
)

const (
	headerBlobType = "x-ms-blob-type"

	blobTypeBlock = "BlockBlob"
)

type client struct {
	DefaultClient *container.Client
	credentials   *azblob.SharedKeyCredential
	contentType   *string
	proxyURL      *url.URL
	bufferSize    int64
}

func NewEmpty(ctx context.Context, opts ...*Options) (storage.ObjectStorage, error) {
	opt := NewOptions(opts...)
	objStore := &client{
		bufferSize:  opt.BufferSize,
		contentType: opt.ContentType,
		proxyURL:    opt.ProxyURI,
	}
	return objStore, nil
}

func New(ctx context.Context, bucket string, opts ...*Options) (storage.ObjectStorage, error) {
	var (
		err    error
		cc     *container.Client
		azCred *azblob.SharedKeyCredential
	)
	opt := NewOptions(opts...)
	objectStorage, err := NewEmpty(ctx, opt)
	if err != nil {
		return nil, err
	}
	clientOptions := &container.ClientOptions{
		ClientOptions: azcore.ClientOptions{
			Transport: &http.Client{
				Transport: &http.Transport{
					TLSClientConfig: &tls.Config{
						RootCAs: storage.GetRootCAs(),
					},
				},
			},
		},
	}
	if opt.ConnectionString != nil {
		cc, err = container.NewClientFromConnectionString(
			*opt.ConnectionString, bucket, clientOptions,
		)
		if err == nil {
			azCred, err = keyFromConnString(*opt.ConnectionString)
		}
	} else if sk := opt.SharedKey; sk != nil {
		var containerURL string
		containerURL, azCred, err = sk.azParams(bucket)
		if err == nil {
			cc, err = container.NewClientWithSharedKeyCredential(
				containerURL,
				azCred,
				clientOptions,
			)
		}
	}
	if err != nil {
		return nil, err
	}
	objectStorage.(*client).DefaultClient = cc
	objectStorage.(*client).credentials = azCred
	if err := objectStorage.HealthCheck(ctx); err != nil {
		return nil, err
	}
	return objectStorage, nil
}

func (c *client) clientFromContext(
	ctx context.Context,
) (client *container.Client, err error) {
	client = c.DefaultClient
	if settings, _ := storage.SettingsFromContext(ctx); settings != nil {
		if err = settings.Validate(); err != nil {
			return nil, err
		} else if settings.ConnectionString != nil {
			client, err = container.NewClientFromConnectionString(
				*settings.ConnectionString,
				settings.Bucket,
				&container.ClientOptions{},
			)
		} else {
			var (
				containerURL string
				azCreds      *azblob.SharedKeyCredential
			)
			creds := SharedKeyCredentials{
				AccountName: settings.Key,
				AccountKey:  settings.Secret,
			}
			if settings.Uri != "" {
				creds.URI = &settings.Uri
			}

			containerURL, azCreds, err = creds.azParams(settings.Bucket)
			if err == nil {
				client, err = container.NewClientWithSharedKeyCredential(
					containerURL,
					azCreds,
					&container.ClientOptions{},
				)
			}
		}
	}
	if client == nil {
		return nil, ErrEmptyClient
	}
	return client, err
}

func (c *client) HealthCheck(ctx context.Context) error {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		if err == ErrEmptyClient {
			return nil
		}
		return OpError{
			Op:     OpHealthCheck,
			Reason: err,
		}
	}
	_, err = azClient.GetProperties(ctx, &container.GetPropertiesOptions{})
	if err != nil {
		return OpError{
			Op:     OpHealthCheck,
			Reason: err,
		}
	}
	return nil
}

type objectReader struct {
	io.ReadCloser
	length int64
}

func (r objectReader) Length() int64 {
	return r.length
}

func (c *client) GetObject(
	ctx context.Context,
	objectPath string,
) (io.ReadCloser, error) {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, OpError{
			Op:     OpGetObject,
			Reason: err,
		}
	}
	bc := azClient.NewBlockBlobClient(objectPath)
	out, err := bc.DownloadStream(ctx, &blob.DownloadStreamOptions{})
	if bloberror.HasCode(err,
		bloberror.BlobNotFound,
		bloberror.ContainerNotFound,
		bloberror.ResourceNotFound) {
		err = storage.ErrObjectNotFound
	}
	if err != nil {
		return nil, OpError{
			Op:     OpGetObject,
			Reason: err,
		}
	}
	if out.ContentLength != nil {
		return objectReader{
			ReadCloser: out.Body,
			length:     *out.ContentLength,
		}, nil
	}
	return out.Body, nil
}

func (c *client) PutObject(
	ctx context.Context,
	objectPath string,
	src io.Reader,
) error {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return OpError{
			Op:     OpPutObject,
			Reason: err,
		}
	}
	bc := azClient.NewBlockBlobClient(objectPath)
	var blobOpts = &blockblob.UploadStreamOptions{
		HTTPHeaders: &blob.HTTPHeaders{
			BlobContentType: c.contentType,
		},
	}
	blobOpts.BlockSize = c.bufferSize
	_, err = bc.UploadStream(ctx, src, blobOpts)
	if err != nil {
		return OpError{
			Op:      OpPutObject,
			Message: "failed to upload object to blob",
			Reason:  err,
		}
	}
	return err
}

func (c *client) DeleteObject(
	ctx context.Context,
	path string,
) error {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return OpError{
			Op:     OpDeleteObject,
			Reason: err,
		}
	}
	bc := azClient.NewBlockBlobClient(path)
	_, err = bc.Delete(ctx, &blob.DeleteOptions{
		DeleteSnapshots: to.Ptr(azblob.DeleteSnapshotsOptionTypeInclude),
	})
	if bloberror.HasCode(err,
		bloberror.BlobNotFound,
		bloberror.ContainerNotFound,
		bloberror.ResourceNotFound) {
		err = storage.ErrObjectNotFound
	}
	if err != nil {
		return OpError{
			Op:      OpDeleteObject,
			Message: "failed to delete object",
			Reason:  err,
		}
	}
	return nil
}

func (c *client) StatObject(
	ctx context.Context,
	path string,
) (*storage.ObjectInfo, error) {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, OpError{
			Op:     OpStatObject,
			Reason: err,
		}
	}
	bc := azClient.NewBlockBlobClient(path)
	if err != nil {
		return nil, OpError{
			Op:      OpStatObject,
			Message: "failed to initialize blob client",
			Reason:  err,
		}
	}
	rsp, err := bc.GetProperties(ctx, &blob.GetPropertiesOptions{})
	if bloberror.HasCode(err,
		bloberror.BlobNotFound,
		bloberror.ContainerNotFound,
		bloberror.ResourceNotFound,
	) {
		err = storage.ErrObjectNotFound
	}
	if err != nil {
		return nil, OpError{
			Op:      OpStatObject,
			Message: "failed to retrieve object properties",
			Reason:  err,
		}
	}
	return &storage.ObjectInfo{
		Path:         path,
		LastModified: rsp.LastModified,
		Size:         rsp.ContentLength,
	}, nil
}

func (c *client) buildSignedURL(
	ctx context.Context,
	method string,
	blobURL string,
	expire time.Duration,
	filename string,
	public bool,
) (*model.Link, error) {
	var permissions sas.BlobPermissions
	switch method {
	case http.MethodGet:
		permissions = sas.BlobPermissions{Read: true}
	case http.MethodDelete:
		permissions = sas.BlobPermissions{Delete: true}
	case http.MethodPut:
		permissions = sas.BlobPermissions{Create: true, Write: true}
	default:
		return nil, fmt.Errorf("invalid HTTP method %q", method)
	}
	now := time.Now().UTC()
	exp := now.Add(expire)
	// HACK: We cannot use BlockBlobClient.GetSASToken because the API does
	// not expose the required parameters.
	urlParts, _ := blob.ParseURL(blobURL)
	sk, proxyURL, err := c.signParamsFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve credentials: %w", err)
	}
	var contentDisposition string
	if filename != "" {
		contentDisposition = fmt.Sprintf(
			`attachment; filename="%s"`, filename,
		)
	}

	qParams, err := sas.BlobSignatureValues{
		ContainerName: urlParts.ContainerName,
		BlobName:      urlParts.BlobName,

		Permissions:        permissions.String(),
		ContentDisposition: contentDisposition,

		StartTime:  now.UTC(),
		ExpiryTime: exp.UTC(),
	}.SignWithSharedKey(sk)
	if err != nil {
		return nil, err
	}
	baseURL, err := url.Parse(blobURL)
	if err != nil {
		return nil, err
	}
	qSAS, err := url.ParseQuery(qParams.Encode())
	if err != nil {
		return nil, err
	}
	q := baseURL.Query()
	for key, values := range qSAS {
		for _, value := range values {
			q.Add(key, value)
		}
	}
	baseURL.RawQuery = q.Encode()
	if public {
		baseURL, err = utils.RewriteProxyURL(baseURL, proxyURL)
		if err != nil {
			return nil, err
		}
	}
	return &model.Link{
		Expire: exp,
		Method: method,
		Uri:    baseURL.String(),
	}, nil
}

func (c *client) GetRequest(
	ctx context.Context,
	objectPath string,
	filename string,
	duration time.Duration,
	public bool,
) (*model.Link, error) {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, OpError{
			Op:     OpGetRequest,
			Reason: err,
		}
	}
	// Check if object exists
	bc := azClient.NewBlockBlobClient(objectPath)
	if err != nil {
		return nil, OpError{
			Op:      OpGetRequest,
			Message: "failed to initialize blob client",
			Reason:  err,
		}
	}
	_, err = bc.GetProperties(ctx, &blob.GetPropertiesOptions{})
	if bloberror.HasCode(err,
		bloberror.BlobNotFound,
		bloberror.ContainerNotFound,
		bloberror.ResourceNotFound,
	) {
		err = storage.ErrObjectNotFound
	}
	if err != nil {
		return nil, OpError{
			Op:      OpGetRequest,
			Message: "failed to check preconditions",
			Reason:  err,
		}
	}
	link, err := c.buildSignedURL(
		ctx,
		http.MethodGet,
		bc.URL(),
		duration,
		filename,
		public,
	)
	if err != nil {
		return nil, OpError{
			Op:      OpGetRequest,
			Message: "failed to create pre-signed URL",
			Reason:  err,
		}
	}
	return link, nil
}

func (c *client) DeleteRequest(
	ctx context.Context,
	path string,
	duration time.Duration,
	public bool,
) (*model.Link, error) {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, OpError{
			Op:     OpGetRequest,
			Reason: err,
		}
	}
	bc := azClient.NewBlobClient(path)
	if err != nil {
		return nil, OpError{
			Op:      OpDeleteRequest,
			Message: "failed to initialize blob client",
			Reason:  err,
		}
	}
	link, err := c.buildSignedURL(ctx, http.MethodDelete, bc.URL(), duration, "", public)
	if err != nil {
		return nil, OpError{
			Op:      OpDeleteRequest,
			Message: "failed to generate signed URL",
			Reason:  err,
		}
	}

	return link, nil
}

func (c *client) PutRequest(
	ctx context.Context,
	objectPath string,
	duration time.Duration,
	public bool,
) (*model.Link, error) {
	azClient, err := c.clientFromContext(ctx)
	if err != nil {
		return nil, OpError{
			Op:     OpGetRequest,
			Reason: err,
		}
	}
	bc := azClient.NewBlobClient(objectPath)
	if err != nil {
		return nil, OpError{
			Op:      OpPutRequest,
			Message: "failed to initialize blob client",
			Reason:  err,
		}
	}
	link, err := c.buildSignedURL(ctx, http.MethodPut, bc.URL(), duration, "", public)
	if err != nil {
		return nil, OpError{
			Op:      OpPutRequest,
			Message: "failed to generate signed URL",
			Reason:  err,
		}
	}
	link.Header = map[string]string{
		headerBlobType: blobTypeBlock,
	}
	return link, nil
}
