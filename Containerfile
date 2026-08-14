FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    git \
    curl \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz | tar -C /usr/local -xz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN curl -fsSL https://opencode.ai/install | bash

RUN useradd -m -s /bin/bash agent

USER agent

WORKDIR /home/agent

CMD ["opencode"]