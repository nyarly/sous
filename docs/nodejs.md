# Souss NodeJS Pack

Please note that all builds performed by sous are based on Dockerfiles. At any point, you can use the command

    $ sous dockerfile

inside your project to see what dockerfile sous would build if you performed a `sous build`. Likewise, if you want to see the Dockerfile sous would use for testing `sous test`, you can use

    $ sous dockerfile test

In fact, `sous dockerfile` is just shorthand for `sous dockerfile build`.


## Detect
The NodeJS pack is enabled if you have a package.json file in the root of your git repo.

## Build
NodeJS builds your project using one of the available base images for nodeJS. You can see what base images are available by typing `sous pack-info nodejs`

The typical production build will perform an `npm install --production` inside the docker container on build. You can override this npm install step if you want by providing an additional script in your `package.json` called "installProduction", for example:

```json
	{
		...
		"scripts": {
			"installProduction": "npm install"
		}
		...
	}
```

## Test
NodeJS runs your unit tests inside the container. It typically does this by first performing a simple `npm install` and then directly invoking `npm test` inside the container.

You cannot configure the `npm test` part, so the only way to get tests to work with sous NodeJS is to add a `test` script to your package.json. If necessary though, you can override the standard `npm install` used for running the unit tests by providing a separate script named `installTest`, for example:


```json
	{
		...
		"scripts": {
			"installTest": "npm install --dev"
		}
		...
	}
```

## Run
NodeJS runs your project by first building it, according to [Build](#build), above. It then invokes the container, passing in some environment variables, simulating the environment your app will be running in in Mesos.

Presently, `sous run` invokes the following command:

```shell
docker run --net=host -e TASK_HOST=$TASK_HOST -e PORT0=$PORT0 -t "latest-built-image-tag"
```

Where `$TASK_HOST` is automatically set to your docker hostname, and `$PORT0` is automatically set to a free port.

Your application should listen for HTTP traffic at `http://$TASK_HOST:$PORT0`
