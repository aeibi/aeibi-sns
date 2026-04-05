FROM postgres:18-bookworm

RUN set -eux; \
    apt-get update; \
    apt-get install -y --no-install-recommends ca-certificates lsb-release wget; \
    wget -O /tmp/groonga-apt-source.deb \
      "https://packages.groonga.org/debian/groonga-apt-source-latest-$(. /etc/os-release && echo "$VERSION_CODENAME").deb"; \
    apt-get install -y --no-install-recommends /tmp/groonga-apt-source.deb; \
    apt-get update; \
    apt-get install -y --no-install-recommends postgresql-18-pgdg-pgroonga; \
    rm -f /tmp/groonga-apt-source.deb; \
    rm -rf /var/lib/apt/lists/*