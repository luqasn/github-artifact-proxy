FROM --platform=linux/amd64 alpine:latest as certs
RUN apk --update add ca-certificates

FROM scratch
COPY --from=certs /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY github-artifact-proxy /

ENTRYPOINT ["/github-artifact-proxy"]
