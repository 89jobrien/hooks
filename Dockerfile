FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

RUN apk add --no-cache bash make

COPY . .
RUN make all

FROM alpine:3.20

RUN addgroup -S hooks && adduser -S hooks -G hooks
COPY --from=build /src/bin /opt/hooks/bin
COPY --from=build /src/config.yaml /opt/hooks/config.yaml
COPY --from=build /src/scripts /opt/hooks/scripts

ENV PATH="/opt/hooks/bin:${PATH}"
USER hooks
WORKDIR /opt/hooks

CMD ["audit"]
