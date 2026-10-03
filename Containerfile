FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    git \
    curl \
    make \
    file \
    build-essential \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

RUN curl -fsSL https://go.dev/dl/go1.26.6.linux-amd64.tar.gz | tar -C /usr/local -xz

ENV PATH="/usr/local/go/bin:${PATH}"

RUN useradd -m -s /bin/bash agent
RUN chown agent:agent -R /home/agent/


USER agent

WORKDIR /home/agent

# sometimes the agent tries to put its plan here
RUN mkdir -p /home/agent/.opencode/plan

RUN bash -c 'echo "export PATH=$PATH:/home/agent/go/bin" >> /home/agent/.bashrc'

RUN curl -LsSf https://raw.githubusercontent.com/ast2llm/ast2llm-go/main/install.sh | sh
RUN curl -sSfL https://golangci-lint.run/install.sh | sh -s -- -b $(go env GOPATH)/bin v2.14.0

RUN go install github.com/fpt/go-dev-mcp/godevmcp@latest
RUN go install golang.org/x/tools/cmd/deadcode@latest
RUN go install golang.org/x/tools/gopls@latest

RUN go clean -cache -modcache

RUN curl -fsSL https://opencode.ai/v2/install | bash

CMD ["opencode"]