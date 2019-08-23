// +build integration

package main

import (
	"context"
	"encoding/json"
	"os"
	"testing"
	"time"

	"github.com/bradleyjkemp/cupaloy/v2"
	"github.com/olivere/elastic"
)

func TestNewElasticDb_AddLog(t *testing.T) {
	testindex := "integrationtestunit"
	db, err := NewElasticDb(getElasticUrl(), "")
	if err != nil {
		t.Fatal("Error by connect to db" + err.Error())
	}
	type Testdata struct {
		Message string
	}
	countBefore, err := getCountOfDocumentsAtIndex(db.ctx, db.dbClient, testindex)
	if err != nil {
		countBefore = 0
	}
	db.AddLog(LogData{
		Timestamp: time.Now(),
		Type:      string(MessageType_Log),
		Logs: Testdata{
			Message: "Test",
		},
		Attributes: Attributes{
			Host: "Integrationtest",
		},
	}, testindex)

	countAfter, err := getCountOfDocumentsAtIndex(db.ctx, db.dbClient, testindex)

	if err != nil {
		t.Fatal("Error by get response" + err.Error())
	}
	if countBefore+1 != countAfter {
		t.Errorf("Added Log but counts of document is not one more want %v got %v", countBefore+1, countAfter)
	}

}

func TestNewElasticDb_AddStats(t *testing.T) {
	testindex := "integrationtestunit_stats"

	db, err := NewElasticDb(getElasticUrl(), "")
	if err != nil {
		t.Fatal("Error by connect to db" + err.Error())
	}
	type Testdata struct {
		Message string
	}
	countBefore, err := getCountOfDocumentsAtIndex(db.ctx, db.dbClient, testindex)
	if err != nil {
		countBefore = 0
	}
	db.AddStats(StatsData{
		Timestamp: time.Now(),
		Type:      string(MessageType_Log),
		Stats: Testdata{
			Message: "Test",
		},
		Attributes: Attributes{
			Host: "Integrationtest",
		},
	}, testindex)

	countAfter, err := getCountOfDocumentsAtIndex(db.ctx, db.dbClient, testindex)

	if err != nil {
		t.Fatal("Error by get response" + err.Error())
	}
	if countBefore+1 != countAfter {
		t.Errorf("Added Log but counts of document is not one more want %v got %v", countBefore+1, countAfter)
	}

}

func TestNewElasticDb_SetPolicyTemplate(t *testing.T) {
	db, err := NewElasticDb(getElasticUrl(), "")
	if err != nil {
		t.Fatal("Error by connect to db" + err.Error())
	}
	err = db.SetPolicyTemplate()
	if err != nil {
		t.Fatal("Error by SetPolicyTemplate " + err.Error())
	}
	ts := elastic.NewIndicesGetTemplateService(db.dbClient)
	resp, err := ts.Name("funk_template").Do(db.ctx)
	if err != nil {
		t.Fatal("Error by get response" + err.Error())
	}
	cupaloy.SnapshotT(t, resp)

}

func TestNewElasticDb_SetIlmPolicy(t *testing.T) {
	db, err := NewElasticDb(getElasticUrl(), "")
	if err != nil {
		t.Fatal("Error by connect to db" + err.Error())
	}
	db.SetIlmPolicy("20d")
	ls := elastic.NewXPackIlmGetLifecycleService(db.dbClient)
	res, err := ls.Policy("funk_policy").Do(db.ctx)
	if err != nil {
		t.Fatal("Error by get response" + err.Error())
	}
	phases := res["funk_policy"].Policy["phases"]
	jsonphases, err := json.Marshal(phases)
	cupaloy.SnapshotT(t, jsonphases)
	if err != nil {
		t.Fatal("Error by json.Marshal" + err.Error())
	}

}

func getCountOfDocumentsAtIndex(ctx context.Context, db *elastic.Client, index string) (int64, error) {
	is := elastic.NewIndicesStatsService(db)
	resp, err := is.Index(index).Do(ctx)

	if err != nil {
		return 0, err
	}
	return resp.All.Primaries.Indexing.IndexTotal, nil
}

func getElasticUrl() string {
	res := os.Getenv("ELASTICSEARCH")
	if res == "" {
		return "http://127.0.0.1:9200"
	}
	return res
}
