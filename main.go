package main

import (
	"net/http"
	"os"
	"strconv"

	"github.com/fasibio/funk-server/logger"

	"github.com/go-chi/chi"
	"github.com/gorilla/websocket"
	"github.com/sony/sonyflake"
	"github.com/urfave/cli"
)

type Handler struct {
	dataserviceHandler *DataServiceWebSocket
	connectionkey      string
}

const (
	HTTP_PORT             = "port"
	ELASTICSEARCH_URL     = "elasticSearchUrl"
	CONNECTION_KEY        = "connectionkey"
	USE_DELETE_POLICY     = "usedeletePolicy"
	MIN_AGE_DELETE_POLICY = "minagedeletepolicy"
)

func main() {

	app := cli.NewApp()
	app.Name = "funk Server"
	app.Action = run
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   HTTP_PORT + ", p",
			EnvVar: "HTTP_PORT",
			Value:  "3000",
			Usage:  "`HTTP_PORT` to start the server on",
		},
		cli.StringFlag{
			Name:   ELASTICSEARCH_URL,
			EnvVar: "ELASTICSEARCH_URL",
			Value:  "http://127.0.0.1:9200",
			Usage:  "Elasticsearch url",
		},
		cli.StringFlag{
			Name:   CONNECTION_KEY,
			EnvVar: "CONNECTION_KEY",
			Value:  "changeMe04cf242924f6b5f96",
			Usage:  "The connectionkey given to the funk_agent so he can connect",
		},
		cli.BoolTFlag{
			Name:   USE_DELETE_POLICY,
			EnvVar: "USE_DELETE_POLICY",
			Usage:  "Default is enabled it will set an ilm on funk indexes",
		},
		cli.StringFlag{
			Name:   MIN_AGE_DELETE_POLICY,
			EnvVar: "MIN_AGE_DELETE_POLICY",
			Value:  "90d",
			Usage:  "Set the Date to delete data from the funk indexes",
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Get().Errorw(err.Error())
	}

}

func run(c *cli.Context) error {
	logger.Initialize("info")
	logger.Get().Infow("elasticSearchUrl:" + c.String(ELASTICSEARCH_URL))
	db, err := NewElasticDb(c.String(ELASTICSEARCH_URL), "")
	port := c.String(HTTP_PORT)
	if err != nil {
		logger.Get().Fatal(err)
	}
	handler := Handler{
		connectionkey: c.String(CONNECTION_KEY),
		dataserviceHandler: &DataServiceWebSocket{
			Db:                &db,
			ClientConnections: make(map[string]*websocket.Conn),
			genUID:            genUID,
		},
	}
	if c.BoolT(USE_DELETE_POLICY) {
		if err := db.setIlmPolicy(c.String(MIN_AGE_DELETE_POLICY)); err != nil {
			logger.Get().Fatalw("error create ilm policy: " + err.Error())
		}
		if err := db.setPolicyTemplate(); err != nil {
			logger.Get().Fatalw("error set policy template: " + err.Error())
		}
	}

	handler.dataserviceHandler.ConnectionAllowed = handler.ConnectionAllowed
	router := chi.NewMux()
	router.Get("/", handler.dataserviceHandler.Root)
	router.HandleFunc("/data/subscribe", handler.dataserviceHandler.Subscribe)
	logger.Get().Infow("Starting at port " + port)
	logger.Get().Fatal(http.ListenAndServe(":"+port, router))
	return nil

}

func (h *Handler) ConnectionAllowed(r *http.Request) bool {
	key := r.Header.Get("funk.connection")
	return key == h.connectionkey
}

func genUID() (string, error) {
	flake := sonyflake.NewSonyflake(sonyflake.Settings{})
	id, err := flake.NextID()
	if err != nil {
		logger.Get().Errorw("flake.NextID() failed with" + err.Error())
		return "", err
	}
	return strconv.FormatUint(id, 10), nil
}
