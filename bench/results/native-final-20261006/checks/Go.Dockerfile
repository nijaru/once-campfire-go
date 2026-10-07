FROM golang:1.27.1
RUN apt-get update && apt-get install -y --no-install-recommends libvips-dev libzstd-dev zstd ffmpeg python3 util-linux && rm -rf /var/lib/apt/lists/*
WORKDIR /src
