FROM campfire-compare-toolchain
LABEL org.opencontainers.image.revision="99d1cf8" org.nijaru.experiment="constant-space-negotiation"
COPY perf-policy/go-candidate /rails/bin/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/rails/bin/campfire", "server"]
