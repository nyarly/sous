# sous

Sous is your personal sous-chef for building and deploying projects
on the OpenTable Mesos Platform.

## Features

- build projects as Docker images
- test projects as Docker containers
- run projects locally
- push projects as Docker images
- deploy projects globally via a single command (coming soon)
- ensure projects fulfil contracts of the system (coming soon)

## Ethos

Sous is designed to work with existing projects, using data already contained
to determine how to properly build Docker images. It is designed to ease migrating
existing projects onto The Mesos Platform, using sensible defaults and stack-centric
conventions. It is also designed with operations in mind, tagging Docker images
with lots of metadata to ease discovery and cleaunup of images.

Sous works on your local dev machine, and on CI servers like TeamCity in the same
way, so you can be sure whatever works locally will also work in CI.

## Installation

Sous is written in Go. If you already have Go 1.5 set up on your machine, and have
your GOPATH set up correctly, you can install it by typing

    $ go install github.com/opentable/sous

Alternatively, you can install the latest version on your Mac using homebrew:
(coming soon)

    $ brew tap opentable/osx-tools
	$ brew update
	$ brew install sous

## Requirements

Sous needs to shell out to your system to interact with Git and Docker, so
you will need:

- Git >=2.2.0
- Docker >=1.8.2

On Mac, I would recommend installing Docker by installing docker-machine via the
Docker Toolbox available at https://www.docker.com/toolbox

## Basic Usage

Sous works by interrogating your Git repo to sniff out what kind of project it is
and some other info like its name, version, what runtime version it needs etc.
Using this data, it is able to create sensible Dockerfiles to perform various tasks
like building and testing your project. It also applies labels to the Dockerfile
which propagate through to the image, and finally the running containers, with data
such as which Git commit was built, what stack is running inside it, which user and
host it was built on, and a load more.

This approach is inspired by Heroku's buildpacks, which are not quite suitable for
use inside OpenTable, so we have written some of our own, which in turn can make
some assumptions specific to us, and this keep things simple.

### Commands

Note: When using these commands locally, your username will be prefixed to all
Docker images built, e.g. If your name was Algernon Scroggins, your docker tags
would all be of the form:

    docker.registry/ascroggins/product-name:v0.0.0-commitSHA-host-N

Where `v0.0.0` is the application version, `commitSHA` is the first 8 characters of
your currently checked-out commit, `host` is your machine's hostname, and `N` is
the autoincrementing build number for this commit built by this user on this
machine.

Images built on local dev boxes are not something we should be deploying anywhere,
however it can be useful to share images with others for debugging purposes, so you
can certainly push these images.

- `sous detect` has a look around your repo and tells you what can be done
- `sous build` creates a Dockerfile on the fly, and builds and tags it
- `sous run` first does a build, and then runs your container as if it were on
  Mesos\*
- `sous push` pushes your latest successfully built Docker image to the registry
- `sous contracts` (coming soon) runs your container as if it were on Mesos\*, and
  pokes it with a stick a few times to see that it conforms to certain essential
  contracts of the platform, such as discovery announcements, the /health endpoint,
  start-up and shut-down speed, and graceful shutdown.
- `sous deploy` (coming soon) deploys and configures your app globally according to
  a single manifest definition

\* Sort of... More details to be determined soon

**That's it for now, folks, this is a work in progress...**

