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

type ElsticConnection interface {
	AddStats(data StatsData, index string)
	AddLog(data LogData, index string)
	SetIlmPolicy(minDeleteAge string) error
	SetPolicyTemplate() error
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
			time.Sleep(5 * time.Second)
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
	Timestamp     time.Time   `json:"timestamp,omitempty"`
	Type          string      `json:"message_type"`
	Stats         interface{} `json:"stats,omitempty"`
	Attributes    Attributes  `json:"attr,omitempty"`
	StaticContent interface{} `json:"static_content,omitempty"`
}

type LogData struct {
	Timestamp     time.Time   `json:"timestamp,omitempty"`
	Type          string      `json:"message_type,omitempty"`
	Logs          interface{} `json:"logs,omitempty"`
	Attributes    Attributes  `json:"attr,omitempty"`
	StaticContent interface{} `json:"static_content,omitempty"`
}

func getIlmPolicyBody(minDeleteAge string) string {
	return `
	{
		"policy": {                       
			"phases": {
				"delete": {
					"min_age": "` + minDeleteAge + `",           
					"actions": {
						"delete": {}              
					}
				}
			}
		}
	}
	
	`
}

func (k *KonfigData) SetIlmPolicy(minDeleteAge string) error {

	ilmservice := elastic.NewXPackIlmPutLifecycleService(k.dbClient)
	ilmservice.Policy("funk_policy")
	ilmservice.BodyString(getIlmPolicyBody(minDeleteAge))
	res, err := ilmservice.Do(k.ctx)
	if err != nil {
		return err
	}
	logger.Get().Infow("IlmPolicy Created ", "Acknowledged", res.Acknowledged)

	return nil
}

func getFunkLogsDynamicTemplateBody() string {
	return `
	{
		"index_patterns": ["*logs_funk*"],
		"mappings": {
			"dynamic_templates": [
				{
					"integers": {
						"path_match": "logs.funkgeoip.location",
						"mapping": {
							"type": "geo_point"
						}
					}
				}
			]
		}
	}`
}

func (k *KonfigData) SetFunkLogsDynamicTemplate() error {
	_, err := k.dbClient.IndexPutTemplate("funklog_dynamic_template").BodyString(getFunkLogsDynamicTemplateBody()).Do(k.ctx)
	return err
}
func getPolicyTemplateBody() string {
	return `
	{
		"index_patterns": ["*_funk*"],                 
		"settings": {
			"number_of_shards": 1,
			"number_of_replicas": 1,
			"index.lifecycle.name": "funk_policy",      
			"index.lifecycle.rollover_alias": "funk"    
		}
	}
	`
}

func (k *KonfigData) SetPolicyTemplate() error {
	template := elastic.NewIndicesPutTemplateService(k.dbClient)
	template.Name("funk_template")
	template.BodyString(getPolicyTemplateBody())

	res, err := template.Do(k.ctx)
	if err != nil {
		return err
	}
	logger.Get().Infow("PolicyTemplate Created ", "Acknowledged", res.Acknowledged, "Index", res.Index)
	return nil
}

func (k *KonfigData) AddStats(data StatsData, index string) {
	logger.Get().Debugw("statsData from Client for index: " + index)
	bulkRequest := k.dbClient.Bulk()
	tmp := elastic.NewBulkIndexRequest().Index(index).Type("_doc").Id(genID()).Doc(data)
	bulkRequest.Add(tmp)
	_, err := bulkRequest.Do(k.ctx)
	if err != nil {
		logger.Get().Warn("Error by create Document", err)
	}

}

func (k *KonfigData) AddLog(data LogData, index string) {
	logger.Get().Debugw("logData from Client for index: " + index)

	bulkRequest := k.dbClient.Bulk()
	tmp := elastic.NewBulkIndexRequest().Index(index).Type("_doc").Id(genID()).Doc(data)

	bulkRequest.Add(tmp)
	_, err := bulkRequest.Do(k.ctx)
	if err != nil {
		logger.Get().Warn("Error by create Document", err)
	}

}
