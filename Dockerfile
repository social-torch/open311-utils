FROM golang

WORKDIR /storch
RUN apt-get -y update && apt-get install -y groff awscli

ENTRYPOINT "/bin/bash"
