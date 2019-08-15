# Funk-Server


Funk is an simple to use Logging Collector to get your logs to the Elasticsearch Database and visible it with Kibana.
It will be automatically collect all Data from all running Dockercontainer. 

## Motivation

ELK is the state of the art stack to collect Logmessages. 
But 90% only need 1% of the possibilities of Logstash.

The normal modern way is: 

You have an orchestration system like docker (swarm), kubernetes ...
And want to collect the messages of all your Containers. Solution: [funk-agent](https://github.com/fasibio/funk_agent)

I also need a Solution to Manage the logs of different Cluster on different host different Provider at one Database. 
So my Servercomponent have to be protected Publish. 

And last but not least: 
My Mindset ist: 
Hold everything near by your Code. 
So with this solution you can define How and whats logging directly at you docker-compose or manifest over labels tag.

## Getting Started (10min)

### Prerequisites

Take a look at the [docker-compose.yml](./example/docker-compose.yml) and try it!
It will start 5 Container.(funk-server, funk-agent, Kibana, elasticsearch, a test http container) 

After a while of starttime open the browser at http://localhost:5601.

Now go to Management at the Drawer Menu. 

At Elasticsearch ==> Index Management you can find all generate Indexes. All ends with _funk

At Kibana ==> Index Pattern ==> Create index pattern you can create an Index for example httpd_logs*. 

Set Timeinformation and save. 

Now you have an Index with all information of this Service. 

Go to Discover at Drawermenu. 

Here you can filter your logs now. 

Open a second Tab and go to http://127.0.0.1:8080 to create new Log entries

Repeat this for all Index you need Information. 



If you want to find out more about the [funk-agent](https://github.com/fasibio/funk_agent)

All other steps are normal Kibana steps. (Visualization, Dashboards, etc)


Thats it! 

## Whats Stats info ?? 
The [funk-agent](https://github.com/fasibio/funk_agent) can log stats info to give Information about Hardware Usage of each Container like CPU, Memory etc. 
Be careful with this Information! It need many Space at your Elasticsearch db. You can take it off globally at your [funk-agent](https://github.com/fasibio/funk_agent) instaltion or configure for each Container.
 
# Hardware 
Both Service are written with go. 
So it need minimum space. 
- Imagesize [funk-server](https://hub.docker.com/r/fasibio/funk_server): 14 MB

- Imagesize [funk-agent](https://hub.docker.com/r/fasibio/funk_agent/tags): 13 MB

- Mem usage funk-server: ~ 3 MB

- Mem usage funk-agent: ~ 7 MB


## Possible Configuration

 - HTTP_PORT (default: 3000) ==> port to start the server on
 - ELASTICSEARCH_URL ==> URL to Elasticsearch DB
 - CONNECTION_KEY ==> The connectionkey given to the funk_agent so he can connect
 - USE_DELETE_POLICY (default: true) ==>  it will set an [ilm](https://www.elastic.co/guide/en/elasticsearch/reference/current/getting-started-index-lifecycle-management.html) on funk indexes
 - MIN_AGE_DELETE_POLICY (default: 90d) ==> Set the Date to delete data from the funk indexes


To see possible Configurations and available Labels for the container at [funk-agent](https://github.com/fasibio/funk_agent) see there. 


# Protokoll

You can write your own Clients... for spezial problems. 

Its a Websocketconnection...

Path: /data/subscribe

Header ```funk.connection``` should be the CONNECTION_KEY

## Dataformat: 
Look add the [Message struct](./types.go)
Your Logdata will be set on data as json string

You have to send a list of Message as json to the server. 



# Other Clients

A list of all non Docker clients: 
- [funk js agent](https://github.com/fasibio/funk-js-agent) ==> An Agent to call direkt from your javascript/node application to funk server


If you have trouble or need help. Please feel free to oben an issue. 