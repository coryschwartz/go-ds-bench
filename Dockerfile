FROM golang:bullseye
WORKDIR /bench
VOLUME /out
RUN apt update -y && apt install -y  cmake librados-dev
COPY . .
RUN go mod download
ENTRYPOINT [ "bash", "-c" ]
CMD [ "make worker.test && mv worker.test /out/" ]
