FROM campfire-compare-toolchain
LABEL org.opencontainers.image.revision="134f4dd" org.nijaru.experiment="completed-private-response-emission"
COPY perf-emission/go-candidate /rails/bin/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/rails/bin/campfire", "server"]
