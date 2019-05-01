package db

import (
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	"context"
	"github.com/autom8ter/fire/api"
	"google.golang.org/api/option"
	"io"
)

type Client struct {
	proj  string
	store *firestore.Client
	blob  *storage.Client
}

func NewClient(ctx context.Context, projectID string, opts ...option.ClientOption) (*Client, error) {
	client, err := firestore.NewClient(ctx, projectID, opts...)
	if err != nil {
		return nil, err
	}
	strg, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return &Client{
		proj:  projectID,
		store: client,
		blob:  strg,
	}, nil
}

func (c *Client) DocSnapshot(ctx context.Context, group api.Grouping) (*firestore.DocumentSnapshot, error) {
	return c.store.Collection(group.Category()).Doc(group.Identifier()).Get(ctx)
}

func (c *Client) DocRef(ctx context.Context, group api.Grouping) *firestore.DocumentRef {
	return c.store.Collection(group.Category()).Doc(group.Identifier())
}

func (c *Client) MarshalDocTo(ctx context.Context, group api.Grouping, obj interface{}) error {
	snap, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Get(ctx)
	if err != nil {
		return err
	}
	return snap.DataTo(obj)
}

func (c *Client) DocDataAt(ctx context.Context, group api.Grouping, key string) (interface{}, error) {
	snap, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Get(ctx)
	if err != nil {
		return nil, err
	}
	return snap.DataAt(key)
}

func (c *Client) DocData(ctx context.Context, group api.Grouping) (map[string]interface{}, error) {
	snap, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Get(ctx)
	if err != nil {
		return nil, err
	}
	return snap.Data(), nil
}

func (c *Client) UpdateDocField(ctx context.Context, group api.Grouping, key string, value string) error {
	_, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Update(ctx, []firestore.Update{
		{
			Path:  key,
			Value: value,
		},
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CreateDoc(ctx context.Context, group api.Grouping, data map[string]interface{}) error {
	_, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Create(ctx, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteDoc(ctx context.Context, group api.Grouping) error {
	_, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Delete(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) SetDocData(ctx context.Context, group api.Grouping, data map[string]interface{}, merge bool) error {
	if merge {
		_, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Set(ctx, data, firestore.MergeAll)
		if err != nil {
			return err
		}
		return nil
	} else {
		_, err := c.store.Collection(group.Category()).Doc(group.Identifier()).Set(ctx, data)
		if err != nil {
			return err
		}
		return nil
	}
}

func (c *Client) Object(ctx context.Context, group api.Grouping) *storage.ObjectHandle {
	return c.blob.Bucket(group.Category()).Object(group.Identifier())
}

func (c *Client) CopyFromObject(ctx context.Context, from api.Grouping, to api.Grouping) *storage.Copier {
	return c.blob.Bucket(to.Category()).Object(to.Identifier()).CopierFrom(c.Object(ctx, from))
}

func (c *Client) DeleteObject(ctx context.Context, group api.Grouping) error {
	return c.Object(ctx, group).Delete(ctx)
}

func (c *Client) UpdateObjectMetadata(ctx context.Context, metagroup api.MetaGrouping) (*storage.ObjectAttrs, error) {
	return c.Object(ctx, metagroup).Update(ctx, storage.ObjectAttrsToUpdate{
		Metadata: metagroup.Meta(),
	})
}

func (c *Client) ObjectAttributes(ctx context.Context, metagroup api.MetaGrouping) (*storage.ObjectAttrs, error) {
	return c.Object(ctx, metagroup).Attrs(ctx)
}

func (c *Client) GetObjectMetadata(ctx context.Context, metagroup api.MetaGrouping) (map[string]string, error) {
	attrs, err := c.Object(ctx, metagroup).Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return attrs.Metadata, nil
}

func (c *Client) Bucket(ctx context.Context, cat api.Categorizer) *storage.BucketHandle {
	return c.blob.Bucket(cat.Category())
}

func (c *Client) CreateBucket(ctx context.Context, cat api.Categorizer) error {
	return c.blob.Bucket(cat.Category()).Create(ctx, c.proj, nil)
}

func (c *Client) ObjectsBucketName(ctx context.Context, grp api.Grouping) string {
	return c.Object(ctx, grp).BucketName()
}

func (c *Client) ObjectWriter(ctx context.Context, grp api.Grouping) *storage.Writer {
	return c.Object(ctx, grp).NewWriter(ctx)
}

func (c *Client) ObjectReader(ctx context.Context, grp api.Grouping) (*storage.Reader, error) {
	return c.Object(ctx, grp).NewReader(ctx)
}

func (c *Client) CopyObjectTo(ctx context.Context, dst io.Writer, grp api.Grouping) error {
	r, err := c.ObjectReader(ctx, grp)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, r)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) CopyToObjectFrom(ctx context.Context, from io.Reader, grp api.Grouping) error {
	_, err := io.Copy(c.ObjectWriter(ctx, grp), from)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteBucket(ctx context.Context, cat api.Categorizer) error {
	return c.Bucket(ctx, cat).Delete(ctx)
}

func (c *Client) UpdateBucket(ctx context.Context, cat api.Categorizer, attr storage.BucketAttrsToUpdate) (*storage.BucketAttrs, error) {
	return c.Bucket(ctx, cat).Update(ctx, attr)
}

func (c *Client) BucketOsbject(ctx context.Context, cat api.Categorizer) *storage.ObjectIterator {
	return c.Bucket(ctx, cat).Objects(ctx, nil)
}

func (c *Client) Buckets(ctx context.Context) *storage.BucketIterator {
	return c.blob.Buckets(ctx, c.proj)
}
