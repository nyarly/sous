# Sous quick-start

## What is sous?

Sous is a smart wrapper around the Docker CLI tool that provides some useful defaults and simplifies the build-test-publish cycle. Sous also knows how to invoke your application in a local simulated Mesos environment, and can even check your app against contracts required by the platform. This allows you to make sure all the basics are right before you even publish your app, and will hopefully lead to quicker success on the Mesos platform.

_In phase 2, we will be adding support for synchronised global deployments, including remote smoke testing and blue/green deployment pools. This should take a lot of the complexity out of your build and deploy configurations, and may encourage more people to start using these advanced build pipelines._

Sous currently only supports NodeJS projects, but we plan to add support for C#, Java, Ruby, Go, and others very soon.

Once you are happy building your project with Sous, you should remove any Dockerfile left over in your project; Sous does not, and cannot, use your custom Dockerfile, it generates standardised dockerfiles on the fly instead. If you leave your Dockerfile in place, it will simply be ignored by Sous.

## The basics

Here is a quick overview of key concepts in Sous

### Targets

Sous uses the concept of "targets" to build Docker images with specialised purposes. Here's the current list of targets...

- `app` is the main (default) target, its Docker image will serve your application when it is run with `docker run`
- `test` is a special target whose only job is to run unit tests (`docker run` on this target just runs unit tests and then exits)
- `compile` is an optional special target whose job is to gather all dependencies, and possibly perform other precompile steps. `docker run` on this target will invoke the build process with a special directory mounted at `/artifacts`; the build container's job is to perform all necessary tasks, and then place a complete representation of the application in the `/artifacts` directory. This will then be placed inside the `app` container ready for deployment.
- `smoke` is an optional special target whose job is to run remote smoke tests against individual instances of your deployed application. Smoke tests here can be as simple or as involved as you like, but should at a minimum ensure that your `/health` endpoint is reporting healthy, and that all your other essential routes seem to be working. Eventually this will become a required dependency of the deployment target (coming soon).

Once you have installed sous, following the instructions in the readme, the best place to start is

### Transparency

Sous is not magic, and doesn't pretend to be, it is designed solely to take the tedium out of doingitrite. As such, it follows a few design principles:

- It logs most of the shell commands it is executing to screen prefixed by `shell>`. This means if someehing goes wrong during a sous operation, you should be able to scroll up and see what it was trying to do, which may be helpful in fixing said issues.
- `Dockerfiles` that sous would build for a given target are always available by issuing `sous dockerfile <target>` inside your project directory. If you omit `target` then `app` is always used as the default.
- If you prefer to invoke docker yourself for some reason, you can add the `-command` flag to the `build`, `test`, `smoke`, `run` commands to see what command sous would issue, but without issuing it directly.

### How do I add support for target X?

Sous very deliberately does not require any special files (imagine a "Sousfile") or anything else specific to Sous in your project. One of its key design principles is that a project that works well with Sous will also be a model project in the language/framework in which it was built. To say it another way, the complexity of building projects as Docker images to run on the Mesos Platform is the responsibility of Sous. Your responsibility should be limited to expressing in a way idiomatic to your chosen stack all of your project's dependencies (in the generic sense, including build-time libs, the runtime, any run-time modules, etc.).

Therefore, for each target supported in your stack, there is online help available telling you how to add support for that target. You can issue a `sous detect` to check what targets your project currently supports.

