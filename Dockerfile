FROM gcr.io/distroless/static-debian10
COPY dashboard-manager /
ENTRYPOINT ["/dashboard-manager"]
