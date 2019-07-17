package main

import (
	"context"
	"time"

	"github.com/fasibio/funk-server/logger"
	"github.com/olivere/elastic"
	"github.com/segmentio/ksuid"
)

type KonfigData struct {
	dbClient *elastic.Client
	ctx      context.Context
	mapping  string
}

func genID() string {
	id := ksuid.New()
	return id.String()
}
func NewElasticDb(url, esmapping string) (KonfigData, error) {
	ctx := context.Background()
	var client *elastic.Client
	for i := 0; i < 20; i++ {
		c, err := elastic.NewSimpleClient(elastic.SetURL(url))
		if err != nil {
			time.Sleep(1 * time.Second)
			if i == 19 {
				return KonfigData{}, err
			}
			logger.Get().Errorw("Error by Connect to Elastic Search:" + err.Error())
		} else {
			client = c
			break
		}
	}

	for i := 0; i < 10; i++ {
		info, code, err := client.Ping(url).Do(ctx)
		if err != nil {
			time.Sleep(2 * time.Second)
			logger.Get().Infow("Error by Ping Try again to Find Elasticsearchdb")
			if i == 9 {
				return KonfigData{}, err
			}
		} else {
			logger.Get().Infof("Elasticsearch returned with code %d and version %s\n", code, info.Version.Number)
			break
		}

	}
	esversion, err := client.ElasticsearchVersion(url)
	if err != nil {
		logger.Get().Errorw("Error by VersionQuestion, return empty Handler: " + err.Error())
		return KonfigData{}, err
	}
	logger.Get().Infof("Elasticsearch version %s\n", esversion)

	// exists, err := client.IndexExists(index).Do(ctx)
	// if err != nil {
	// 	return KonfigData{}, err
	// }
	// if exists {
	// 	log.Println("Index found Delete Index", index)
	// 	client.DeleteIndex(index_logs).Do(ctx)
	// }
	// _, err = client.CreateIndex(index_logs).BodyString(esmapping).Do(ctx)
	// if err != nil {
	// 	return KonfigData{}, err
	// }

	return KonfigData{
		ctx:      ctx,
		dbClient: client,
		mapping:  esmapping,
	}, nil
}

type StatsData struct {
	Containername string      `json:"containername,omitempty"`
	ContainerID   string      `json:"containerid,omitempty"`
	Timestamp     time.Time   `json:"timestamp,omitempty"`
	Host          string      `json:"host,omitempty"`
	Type          string      `json:"-"`
	Stats         interface{} `json:"stats,omitempty"`
}

type LogData struct {
	Containername string      `json:"containername,omitempty"`
	ContainerID   string      `json:"containerid,omitempty"`
	Timestamp     time.Time   `json:"timestamp,omitempty"`
	Host          string      `json:"host,omitempty"`
	Type          string      `json:"-"`
	Logs          interface{} `json:"logs,omitempty"`
}

func (k *KonfigData) AddStats(data StatsData, index string) {
	bulkRequest := k.dbClient.Bulk()
	tmp := elastic.NewBulkIndexRequest().Index(index).Type(data.Type).Id(genID()).Doc(data)
	bulkRequest.Add(tmp)
	bulkRequest.Do(k.ctx)
}

func (k *KonfigData) AddLog(data LogData, index string) {
	logger.Get().Infow("logData from Client for index: " + index)

	bulkRequest := k.dbClient.Bulk()
	tmp := elastic.NewBulkIndexRequest().Index(index).Type(data.Type).Id(genID()).Doc(data)
	bulkRequest.Add(tmp)
	bulkRequest.Do(k.ctx)
}
