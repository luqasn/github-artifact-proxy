FROM scratch
COPY github-artifact-proxy /
ENTRYPOINT ["/github-artifact-proxy"]
