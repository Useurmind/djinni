FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    git \
    curl \
    make \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://go.dev/dl/go1.22.5.linux-amd64.tar.gz | tar -C /usr/local -xz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN useradd -m -s /bin/bash agent

USER agent

WORKDIR /home/agent

RUN bash -c 'echo "export PATH=$PATH:/home/agent/go/bin" >> /home/agent/.bashrc'

RUN go install golang.org/x/tools/cmd/deadcode@latest
RUN go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.12.2
RUN go install golang.org/x/tools/gopls@latest


RUN curl -fsSL https://opencode.ai/install | bash

CMD ["opencode"]