package es

import (
	"log/slog"
	"time"

	"github.com/elastic/go-elasticsearch/v9"
)

var client *elasticsearch.Client

func Get() *elasticsearch.Client {
	return client
}

func Init() error {
	var err error
	client, err = elasticsearch.NewClient(elasticsearch.Config{
		CloudID: "<CloudID>",
		APIKey:  "<ApiKey>",
	})
	if err != nil {
		time.Sleep(3 * time.Second)
		slog.Info("Retrying to connect to Elasticsearch")
		return Init()
	}
	return nil
}
