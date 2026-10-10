FROM docker.io/library/golang:1.27.1-trixie
RUN apt-get update -qq && apt-get install --no-install-recommends -y libvips-dev ffmpeg sqlite3 ruby git
COPY source /build/source
COPY upstream /build/upstream
RUN cd /build/source && CGO_ENABLED=1 go build -buildvcs=false -trimpath -tags sqlite_fts5 -o /go-final ./cmd/campfire && go build -trimpath -o /httprate ./bench/httprate && cd /build/upstream && CGO_ENABLED=1 go build -buildvcs=false -trimpath -tags sqlite_fts5 -o /go-upstream ./cmd/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/go-final","server"]
