FROM localhost/campfire-post-fix-desktop-before-wrapper:latest
COPY current /build/current-trace
COPY upstream /build/upstream-trace
RUN rm -rf /build/upstream-trace/assets/generated && cp -a /build/upstream/assets/generated /build/upstream-trace/assets/generated
RUN cd /build/current-trace && CGO_ENABLED=1 go test -race ./bench/httprate && CGO_ENABLED=1 go vet -tags sqlite_fts5 ./cmd/campfire ./internal/front ./internal/database && CGO_ENABLED=1 go build -buildvcs=false -trimpath -tags sqlite_fts5 -o /go-current-trace ./cmd/campfire && go build -trimpath -o /httprate-trace ./bench/httprate && cd /build/upstream-trace && CGO_ENABLED=1 go vet -tags sqlite_fts5 ./cmd/campfire ./internal/front ./internal/database && CGO_ENABLED=1 go build -buildvcs=false -trimpath -tags sqlite_fts5 -o /go-upstream-trace ./cmd/campfire
WORKDIR /rails
ENV CAMPFIRE_STORAGE_PATH=/rails/storage
CMD ["/go-current-trace","server"]
