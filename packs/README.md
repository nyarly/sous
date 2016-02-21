# buildpacks

This is the plan for parsing and using buildpacks contained in github.com/opentable/sous-buildpacks

1. buildpack directory name is the name of the buildpack
2. TemporaryLinkResource("golang/detect.sh") run it in real repo dir
3. Create script template:
     #!/bin/sh
	 cd /mnt/repo
     cp $(git ls-files --ignore-standard --cached) /project/$PROJ_NAME
	 cd /project/$PROJ_NAME/$REPO_WORKDIR
	 . /buildpack/common.sh # copy the common (to all packs) script here as well
     . /buildpack/base.sh
     . /buildpack/compile.sh
4. Create Dockerfile adding:
     VOLUME $(host's repo base dir) /mnt/repo
     VOLUME $(host's artifact path) /mnt/artifact
     ADD ./buildpacks/golang /buildpack
     ADD $(script from 3.) /build-$PROJ_NAME
     RUN groupadd + useradd stuff from current buildpacks
     ENV set these from context:
         PROJ_NAME PROJ_VERSION PROJ_REVISION PROJ_DIRTY BASE_DIR
         REPO_DIR REPO_WORKDIR ARTIFACT_DIR  
	 RUN mkdir -p "$BASE_DIR/$REPO_DIR"
     WORKDIR $BASE_DIR/$REPO_DIR/$REPO_WORKDIR
	 ENTRYPOINT ["/build-$PROJNAME"]

5. Create Docker image from Dockerfile in 5. 
6. Create Docker container from Docker image in 5. This will run and
   if it is using the compilation step, produce artifacts as well.
7. Assuming we created artifacts, zip them up into a file, e.g.
     ARTIFACT_PATH=$PROJ_NAME-$PROJ-VERSION-$PROJ-REVISION.tar.gz
8. Create the app container using something like this Dockerfile:
     FROM some-baseimage # (will need to be defined in sous-state somewhere)
     ENV APP_DIR="$PROJ_NAME-$PROJ_VERSION-$PROJ_REVISION"
     ADD $ARTIFACT_PATH /srv/$APP_DIR
     WORKDIR $APP_DIR 
     ENTRYPOINT $(/buildpack/command.sh)

9. All done!
