# GoCrack in Docker

Requirements:

1. A server with docker installed for the server container (this does not require any GPUs).
1. A server with NVIDIA GPUs w/ [NVIDIA docker](https://github.com/NVIDIA/nvidia-docker) fork installed.

## Why NVIDIA Docker? 

NVIDIA has a [nice writeup](https://github.com/NVIDIA/nvidia-docker/wiki/Motivation) explaining why they had to fork docker but it essentially boils down to abstracting away the device configuration and need for drivers in the docker container. 

If you have AMD GPUs and know how to easily create containers, submit a PR!

## Running the server container

    docker run -it \
        -e USER_ID=$(echo "$UID") \
        -p <api port>:<api port> \
        -p <rpc port>:1339 \
        -v "/var/lib/gocrack_server:/var/lib/gocrack_server" \
        gocrack/server

The `-p` parameter exposes a port from the docker container to your host so that the GoCrack API and RPC interface are accessible to remote worker. Change these values to the ports your listeners are running on as defined by the [configuration file](config.md).

The `-v` parameter maps a folder from your host machine `(local:docker_container)` into the container. This is where the server config and all the files are saved. The `docker_container` path should always be `/var/lib/gocrack`

The server's yaml configuration file should reside in the local directory you've mapped into the container as `config.yaml`.

## Running the worker container

    nvidia-docker run -it \
        -e USER_ID=$(echo "$UID") \
        -v "/var/lib/gocrack_worker:/var/lib/gocrack_worker" \
        gocrack/worker

The worker's yaml configuration file should reside in the local directory you've mapped into the container as `config.yaml` along with the directive `server.connect_to` being set to IP of the docker host and `<api port>` that you mapped while creating the server container.
