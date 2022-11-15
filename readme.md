# Traefik Hub Agent for Traefik

<p align="center">
    <picture>
      <source media="(prefers-color-scheme: dark)" srcset="./traefik-hub-horizontal-dark-mode@3x.png">
      <source media="(prefers-color-scheme: light)" srcset="./traefik-hub-horizontal-light-mode@3x.png">
      <img alt="Traefik Hub Logo" src="./traefik-hub-horizontal-light-mode@3x.png">
    </picture>
</p>

## Usage

```
NAME:
   Traefik Hub agent for Traefik - Manages a Traefik Hub agent installation

USAGE:
   agent [global options] command [command options] [arguments...]

COMMANDS:
   run      Runs the Hub Agent
   version  Shows the Traefik Hub agent for Traefik version information
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

```
NAME:
   agent run - Runs the Hub Agent

USAGE:
   agent run [command options] [arguments...]

OPTIONS:
   --log.level value                   Log level to use (debug, info, warn, error or fatal) (default: "info") [$LOG_LEVEL]
   --log.format value                  Log format to use (json or console) (default: "json") [$LOG_FORMAT]
   --traefik.host value                Host to advertise for Traefik to reach the Agent authentication server. Required when the automatic discovery fails [$TRAEFIK_HOST]
   --traefik.api-port value            Port of the Traefik entrypoint for API communication with Traefik (default: "9900") [$TRAEFIK_API_PORT]
   --traefik.tunnel-port value         Port of the Traefik entrypoint for tunnel communication (default: "9901") [$TRAEFIK_TUNNEL_PORT]
   --hub.token value                   The token to use for Hub platform API calls [$HUB_TOKEN]
   --auth-server.listen-addr value     Address on which the auth server listens for auth requests (default: "0.0.0.0:80") [$AUTH_SERVER_LISTEN_ADDR]
   --auth-server.advertise-addr value  Address on which Traefik can reach the Agent auth server. Required when the automatic IP discovery fails [$AUTH_SERVER_ADVERTISE_ADDR]
   --traefik.tls.ca value              Path to the certificate authority which signed TLS credentials [$TRAEFIK_TLS_CA]
   --traefik.tls.cert agent.traefik    Path to the certificate (must have agent.traefik domain name) used to communicate with Traefik Proxy [$TRAEFIK_TLS_CERT]
   --traefik.tls.key value             Path to the key used to communicate with Traefik Proxy [$TRAEFIK_TLS_KEY]
   --traefik.tls.insecure              Activate insecure TLS (default: false) [$TRAEFIK_TLS_INSECURE]
   --traefik.docker.swarm-mode         Activate Traefik Docker Swarm Mode (default: false) [$TRAEFIK_DOCKER_SWARM_MODE]
   --help, -h                          show help (default: false)
```


## mods
```bash

make
# add missing provider
add consulcatalog provider files from https://github.com/traefik/traefik/
# get missing dependancies 
go mod tidy
# run lint/build
make aka : default: clean lint test build


# changes
traefik.consulCatalog.namespace
traefik.consulCatalog.exposedByDefault
traefik.consulCatalog.cache
traefik.consulCatalog.watch
traefik.consulCatalog.endpoint.address
traefik.consulCatalog.endpoint.scheme
traefik.consulCatalog.endpoint.token
traefik.consulCatalog.endpoint.datacenter

#
consulcatalog.ProviderBuilder 
# test
export HUB_TOKEN=mytoken
export TRAEFIK_CONSULCATALOG_ENDPOINT_TOKEN=mytoken
#export TRAEFIK_CONSULCATALOG_NAMESPACE="default"
export TRAEFIK_CONSULCATALOG_WATCH=true
export TRAEFIK_CONSULCATALOG_CACHE=true
export TRAEFIK_CONSULCATALOG_ENDPOINT_DATACENTER="dc1"
export TRAEFIK_CONSULCATALOG_ENDPOINT_ADDRESS="http://localhost:8500"
export TRAEFIK_CONSULCATALOG_ENDPOINT_SCHEME="http"


# consul
docker run \
-d \
--name=consul \
-p 8500:8500 \
-p 8600:8600/udp \
consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
docker run \
-d \
--network traefik-hub \
--name=consul-hub \
-p 8501:8500 \
-p 8601:8600/udp \
consul agent -server -ui -node=server-1 -bootstrap-expect=1 -client=0.0.0.0
# test
curl localhost:8500/v1/catalog/nodes | jq .
# traefik
docker run \
-d \
--name traefik \
-p 80:80 \
-p 9900:9900 \
-p 9901:9901 \
-p 8080:8080 \
traefik:v2.7 \
--experimental.hub=true \
--hub.tls.insecure=true \
--metrics.prometheus.addrouterslabels=true \
--api.dashboard=true

# agent
hub-agent-traefik run \
--auth-server.advertise-url=http://localhost:8001 \
--auth-server.listen-addr=0.0.0.0:8001 \
--traefik.host localhost \
--traefik.tls.insecure true 


# --traefik.consulCatalog.namespace "default" \
# --traefik.consulcatalog.cache true \
# --traefik.consulcatalog.watch true \
# --traefik.consulCatalog.endpoint.datacenter "terraform-consul-dev" \
# --traefik.consulCatalog.endpoint.address "https://myconsul" \
# --traefik.consulCatalog.endpoint.scheme "https" \
# --traefik.consulcatalog.endpoint.token "1234"


```
