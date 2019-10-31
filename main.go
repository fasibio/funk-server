package main

import (
	"errors"
	"fmt"
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
	HTTP_PORT              = "port"
	ELASTICSEARCH_URL      = "elasticSearchUrl"
	CONNECTION_KEY         = "connectionkey"
	USE_ILM_POLICY         = "useIlmPolicy"
	ELASTICSEARCH_USERNAME = "elasticsearchUsername"
	ELASTICSEARCH_PASSWORD = "elasticsearchPassword"
	DATA_ROLLOVER_PATTERN  = "datarolloverpattern"
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
			Name:   USE_ILM_POLICY,
			EnvVar: "USE_ILM_POLICY",
			Usage:  "Default is enabled it will set an ilm on funk indexes",
		},
		cli.StringFlag{
			Name:   ELASTICSEARCH_USERNAME,
			EnvVar: "ELASTICSEARCH_USERNAME",
			Value:  "",
			Usage:  "Username for elasticsearch connection",
		},
		cli.StringFlag{
			Name:   ELASTICSEARCH_PASSWORD,
			EnvVar: "ELASTICSEARCH_PASSWORD",
			Value:  "",
			Usage:  "Password for elasticsearch connection",
		},
		cli.StringFlag{
			Name:   DATA_ROLLOVER_PATTERN,
			EnvVar: "DATA_ROLLOVER_PATTERN",
			Value:  string(Weekly),
			Usage:  fmt.Sprintf("Timeintervall to rollover data possible values: %s, %s, %s", Daily, Monthly, Weekly),
		},
	}

	if err := app.Run(os.Args); err != nil {
		logger.Get().Errorw(err.Error())
	}
}

func setIlmPolicy(db ElsticConnection, indextype DataRolloverPattern) error {
	if err := db.SetIlmPolicy(indextype); err != nil {
		return errors.New("error create ilm policy: " + err.Error())
	}
	if err := db.SetPolicyTemplate(); err != nil {
		return errors.New("error set policy template: " + err.Error())
	}
	return nil
}

func run(c *cli.Context) error {
	logger.Initialize("info")
	logger.Get().Infow("elasticSearchUrl:" + c.String(ELASTICSEARCH_URL))
	db, err := NewElasticDb(c.String(ELASTICSEARCH_URL), c.String(ELASTICSEARCH_USERNAME), c.String(ELASTICSEARCH_PASSWORD), "")
	port := c.String(HTTP_PORT)
	if err != nil {
		logger.Get().Fatal(err)
	}
	dataRollover := DataRolloverPattern(c.String(DATA_ROLLOVER_PATTERN))

	handler := Handler{
		connectionkey: c.String(CONNECTION_KEY),
		dataserviceHandler: &DataServiceWebSocket{
			Db:                &db,
			ClientConnections: make(map[string]*websocket.Conn),
			genUID:            genUID,
			rollOverPattern:   dataRollover,
		},
	}
	if c.BoolT(USE_ILM_POLICY) {
		err := setIlmPolicy(&db, dataRollover)
		if err != nil {
			logger.Get().Fatalw("setIlmPolicy " + err.Error())
		}
	}
	err = db.SetFunkLogsDynamicTemplate()
	if err != nil {
		logger.Get().Fatalw("FunkLogsDynamicTemplate " + err.Error())
	}

	handler.dataserviceHandler.ConnectionAllowed = handler.ConnectionAllowed
	router := registerHandler(handler.dataserviceHandler)
	logger.Get().Fatal(http.ListenAndServe(":"+port, router))
	return nil

}

func registerHandler(handler Resolver) http.Handler {
	router := chi.NewMux()
	router.Get("/", handler.Root)
	router.HandleFunc("/data/subscribe", handler.Subscribe)
	return router
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
