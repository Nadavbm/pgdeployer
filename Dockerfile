# Build the manager binary
FROM golang:1.17-stretch as builder

# Copy the go source
COPY . /pgdeploy-operator
WORKDIR /pgdeploy-operator

# Build
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on go build -a -o pgdeploy-operator main.go


FROM alpine:latest
WORKDIR /
COPY --from=builder /pgdeploy-operator/pgdeploy-operator /pgdeploy-operator

CMD /pgdeploy-operator
