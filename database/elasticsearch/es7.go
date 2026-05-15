package elasticsearch

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/olivere/elastic/v7"
)

// ES7Connector implements Connector using olivere/elastic v7.
type ES7Connector struct {
	client *elastic.Client
}

// NewES7Connector creates an ES7Connector from Config.
func NewES7Connector(cfg Config) (*ES7Connector, error) {
	opts := []elastic.ClientOptionFunc{
		elastic.SetURL(cfg.Address...),
		elastic.SetSniff(cfg.Sniff),
		elastic.SetHealthcheck(cfg.Healthcheck),
	}
	if cfg.TransportMaxIdle > 0 || cfg.Timeout > 0 {
		transport := &http.Transport{
			MaxIdleConnsPerHost: cfg.TransportMaxIdle,
		}
		opts = append(opts, elastic.SetHttpClient(&http.Client{
			Transport: transport,
			Timeout:   cfg.Timeout,
		}))
	}
	client, err := elastic.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("elastic v7: %w", err)
	}
	return &ES7Connector{client: client}, nil
}

func (c *ES7Connector) Index(ctx context.Context, index, id, typeName string, body any) error {
	_, err := c.client.Index().
		Index(index).
		Type(typeName).
		Id(id).
		BodyJson(body).
		Do(ctx)
	return err
}

func (c *ES7Connector) Update(ctx context.Context, index, id string, doc any) error {
	_, err := c.client.Update().
		Index(index).
		Id(id).
		Doc(doc).
		Do(ctx)
	return err
}

func (c *ES7Connector) Delete(ctx context.Context, index, id string) error {
	_, err := c.client.Delete().
		Index(index).
		Id(id).
		Do(ctx)
	return err
}

func (c *ES7Connector) Search(ctx context.Context, index string, query any) ([]byte, error) {
	// query is expected to be a *elastic.SearchService or raw JSON map.
	var svc *elastic.SearchService
	switch q := query.(type) {
	case *elastic.SearchService:
		svc = q
	default:
		return nil, fmt.Errorf("es: unsupported query type %T", query)
	}
	result, err := svc.Index(index).Do(ctx)
	if err != nil {
		return nil, err
	}
	return json.Marshal(result)
}

func (c *ES7Connector) BulkIndex(ctx context.Context, index string, docs map[string]any) error {
	bulk := c.client.Bulk().Index(index)
	for id, doc := range docs {
		bulk.Add(elastic.NewBulkIndexRequest().Id(id).Doc(doc))
	}
	_, err := bulk.Do(ctx)
	return err
}

func (c *ES7Connector) Close() error {
	c.client.Stop()
	return nil
}
