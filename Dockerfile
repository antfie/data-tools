FROM alpine AS build
RUN addgroup -g 10001 app && adduser --disabled-password -u 10001 -G app -h /app app -s /bin/data-tools
RUN apk --no-cache add ca-certificates build-base go
ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=arm64
ARG BUILD_FLAGS=""
WORKDIR /app
ADD . /app
RUN go build -ldflags="$BUILD_FLAGS" -buildvcs=false -trimpath -o "/app/data-tools" /app

FROM scratch
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /app/data-tools /bin/data-tools
USER app
WORKDIR /app
ENTRYPOINT ["/bin/data-tools"]