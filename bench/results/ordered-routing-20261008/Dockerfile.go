FROM campfire-compare-toolchain
LABEL org.opencontainers.image.revision="98c6914" org.nijaru.experiment="ordered-routing-and-canonical-captures"
COPY perf-routing/go-candidate /rails/bin/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/rails/bin/campfire", "server"]
