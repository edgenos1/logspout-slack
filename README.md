# logspout-slack

[![Build Status](https://travis-ci.org/kalisio/logspout-slack.png?branch=master)](https://travis-ci.org/kalisio/logspout-slack)
[![Code Climate](https://codeclimate.com/github/kalisio/logspout-slack/badges/gpa.svg)](https://codeclimate.com/github/kalisio/logspout-slack)
[![Test Coverage](https://codeclimate.com/github/kalisio/logspout-slack/badges/coverage.svg)](https://codeclimate.com/github/kalisio/logspout-slack/coverage)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

A minimalistic adapter for [logspout](https://github.com/gliderlabs/logspout) to send notifications to [Slack](https://slack.com/) using webhooks.

Follow the instructions to build your own [logspout image](https://github.com/gliderlabs/logspout/tree/master/custom) including this module.
In a nutshell, copy the contents of the `custom` folder and add the following import line above others in `modules.go`:
```go
import (
  _ "github.com/kalisio/logspout-slack"
  ...
)
```

If you'd like to select a particular version create the following `Dockerfile`:
```
ARG VERSION
FROM gliderlabs/logspout:$VERSION

ONBUILD COPY ./build.sh /src/build.sh
ONBUILD COPY ./modules.go /src/modules.go
```

Then build your image with: `docker build --no-cache --pull --force-rm --build-arg VERSION=v3.2.11 -f dockerfile -t logspout:v3.2.11 .`

Run the container like this:
```
docker run --name="logspout" \
	--volume=/var/run/docker.sock:/var/run/docker.sock \
	logspout:v3.2.11 \
	slack://hooks.slack.com
```

You can also deploy it in a Docker Swarm using a configuration like this:
```yml
version: '3.5'

services:
  logspout:
    image: logspout:v3.2.11
    command:
      - 'slack://hooks.slack.com?filter.sources=stdout%2Cstderr&filter.name=*aktnmap*'
    volumes:
      - /etc/hostname:/etc/host_hostname:ro
      - /var/run/docker.sock:/var/run/docker.sock
    environment:
      - SLACK_WEBHOOK_URL=${LOGSPOUT_SLACK_WEBHOOK_URL}
      - SLACK_MESSAGE_FILTER=${LOGSPOUT_SLACK_MESSAGE_FILTER}
      - BACKLOG=false
    healthcheck:
      test: ["CMD", "wget", "-q", "--tries=1", "--spider", "http://localhost:80/health"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 1m
    deploy:
      mode: global
      resources:
        limits:
          cpus: '0.20'
          memory: 256M
        reservations:
          cpus: '0.10'
          memory: 128M
      restart_policy:
        condition: on-failure
    networks:
      - network

networks:
  network:
    name: ${DOCKER_NETWORK}
    external: true
```

## Configuration options

You can use the standard logspout filters to filter container names and output types:
```
docker run --name="logspout" \
	--volume=/var/run/docker.sock:/var/run/docker.sock \
	logspout:v3.2.11 \
	slack://hooks.slack.com?filter.sources=stdout%2Cstderr&filter.name=*my_container*'
```

*Note: you must URL-encode parameter values such as the comma and the name filter is not a regex but rather a [path pattern](https://godoc.org/path#Match)*

You can set your webhook URL using the `SLACK_WEBHOOK_URL` environment variable:
```
docker run --name="logspout" \
	--volume=/var/run/docker.sock:/var/run/docker.sock \
	-e SLACK_WEBHOOK_URL="xxx" \
	logspout:v3.2.11 \
	slack://hooks.slack.com
```

You can filter the messages to be sent to Slack using a [regex](https://godoc.org/regexp#Regexp.MatchString) in the `SLACK_MESSAGE_FILTER` environment variable:
```
docker run --name="logspout" \
	--volume=/var/run/docker.sock:/var/run/docker.sock \
	-e SLACK_MESSAGE_FILTER=".*error" \
	logspout:v3.2.11 \
	slack://hooks.slack.com
```
