# podman build -t logger-app -f Containerfile .
# podman tag logger-app quay.io/kenmoini/logger-app:latest
# podman push quay.io/kenmoini/logger-app:latest

FROM registry.access.redhat.com/ubi9/go-toolset:1.19.13-4.1697647145

CMD ["/opt/app-root/src/bin/main"]

COPY . /opt/app-root/src

RUN go build -o /opt/app-root/src/bin/ main.go
