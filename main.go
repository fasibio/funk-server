package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/sony/sonyflake"
	"github.com/urfave/cli"
)

type Handler struct {
	dataserviceHandler *DataServiceWebSocket
}

func main() {

	app := cli.NewApp()
	app.Name = "funk Server"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "port, p",
			EnvVar: "HTTP_PORT",
			Value:  "3000",
			Usage:  "`HTTP_PORT` to start the server on",
		},
		cli.StringFlag{
			Name:   "elasticSearchUrl",
			EnvVar: "ELASTICSEARCH_URL",
			Value:  "http://127.0.0.1:9200",
			Usage:  "Elasticsearch url",
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}

}

func run(c *cli.Context) error {
	log.Println("elasticSearchUrl", c.String("elasticSearchUrl"))
	db, err := NewElasticDb(c.String("elasticSearchUrl"), "")
	port := c.String("port")
	if err != nil {
		panic(err)
	}
	handler := Handler{
		dataserviceHandler: &DataServiceWebSocket{
			Db:                &db,
			ClientConnections: make(map[string]*websocket.Conn),
			genUID:            genUID,
		},
	}

	router := chi.NewMux()
	router.Get("/", handler.dataserviceHandler.Root)
	router.HandleFunc("/data/subscribe", handler.dataserviceHandler.Subscribe)
	log.Println("Starting at port ", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
	return nil

}

func genUID() (string, error) {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		log.Print("flake.NextID() failed with", err)
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}
