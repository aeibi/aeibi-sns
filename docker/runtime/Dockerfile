FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

COPY release/bin/aeibi /app/aeibi
COPY internal/repository/db/migration /app/migrations

ENTRYPOINT ["/app/aeibi"]
CMD ["--config", "/app/config.yaml"]