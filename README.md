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

Take a log at the [docker-compose.yml](./example/docker-compose.yml) and try it!
It will start 5 Container.(funk-server, funk-agent, kibana, elasticsearch, a test http container) 

If you want to find out more about the [funk-agent](https://github.com/fasibio/funk_agent)


## Whats Stats info ?? 
The [funk-agent](https://github.com/fasibio/funk_agent) can log stats info to give Information about Hardware Usage of each Container like CPU, Memory etc. 
Be careful with this Information! It need many Space at your Elasticseach db. You can take it off globally at your [funk-agent](https://github.com/fasibio/funk_agent) instaltion or configure for each Container.
 