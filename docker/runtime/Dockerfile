FROM scratch

WORKDIR /app

COPY --chown=10001:10001 release/bin/aeibi /app/aeibi
COPY --chown=10001:10001 internal/repository/db/migration /app/migrations

USER 10001:10001

ENTRYPOINT ["/app/aeibi"]
