FROM campfire-compare-toolchain
LABEL org.opencontainers.image.revision="94c2118" org.nijaru.experiment="cookie-verification-allocation"
COPY perf-cookie/go-candidate /rails/bin/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/rails/bin/campfire", "server"]
