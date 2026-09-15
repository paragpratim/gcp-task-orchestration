# The binary must be built before running docker build.
FROM gcr.io/distroless/static-debian12:nonroot

COPY build/orchestrator-service /orchestrator-service

CMD ["/orchestrator-service"]
