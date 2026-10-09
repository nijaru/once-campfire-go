FROM campfire-compare-toolchain
LABEL org.opencontainers.image.revision="47cf073" org.nijaru.experiment="length-framed-response-identities"
COPY perf-keys/go-candidate /rails/bin/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/rails/bin/campfire", "server"]
