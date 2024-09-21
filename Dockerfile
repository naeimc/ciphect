# syntax=docker/dockerfile:1

FROM golang:1.23-alpine3.20 AS build
RUN apk update && apk add --no-cache git
RUN git config --global url.ssh://git@github.com/.insteadOf https://github.com/
WORKDIR /go/src/
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN GOOS=linux GOARCH=amd64 go build -ldflags "-w -s" -o /go/bin/ciphect .

FROM scratch 
COPY --from=build /go/bin/ciphect /bin/ciphect
ENV CIPHECT_ADDRESS=0.0.0.0:80
EXPOSE 80
ENTRYPOINT [ "/bin/ciphect" ]