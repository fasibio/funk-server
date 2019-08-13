# Funk-Server


Funk is an simple to use Logging Collector to get your logs to the Elasticsearch Database and visible it with Kibana.
It will be automatically collect all Data from all running Dockercontainer. 

## Motivation

ELK is the state of the art stack to collect Logmessages. 
But 90% off all only need 1% of the possibilities of Logstash.

The normal modern way is: 

You have an orchestration system like docker (swarm), kubernetes ...
And want to collect the messages of all your Containers. [funk-agent](https://github.com/fasibio/funk_agent)

I also need a Solution to Manage the logs of different Cluster on different host different Provider at one Database. 
So my Servercomponent have to be protected Publish. 

And last but not least: 
My Mindset ist: 
Hold everything near by your Code. 
So with this solution you can define How and whats logging directly at you docker-compose or manifest over labels tag.

## Getting Started (2min)

### Prerequisites

Take a look at the [docker-compose.yml](./example/docker-compose.yml) and try it!
It will start 5 Container.(funk-server, funk-agent, Kibana, elasticsearch, a test http container) 

If you want to find out more about the [funk-agent](https://github.com/fasibio/funk_agent)


## Whats Stats info ?? 
The [funk-agent](https://github.com/fasibio/funk_agent) can log stats info to give Information about Hardware Usage of each Container like CPU, Memory etc. 
Be careful with this Information! It need many Space at your Elasticsearch db. You can take it off globally at your [funk-agent](https://github.com/fasibio/funk_agent) instaltion or configure for each Container.
 

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

Dataformat: 
Look add the [Message struct](./types.go)
Your Logdata will be set on data as json string

You have to send a list of Message as json to the server. 



# Other Clients

A list of all non Docker clients: 
- [funk js agent](https://github.com/fasibio/funk-js-agent) ==> An Agent to call direkt from your javascript/node application to funk server (WIP)


If you have trouble or need help. Please feel free to oben an issue. 