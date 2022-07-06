FROM golang:1.18-bullseye


# Install zsh
RUN sh -c "$(wget -O- https://github.com/deluan/zsh-in-docker/releases/download/v1.1.2/zsh-in-docker.sh)" -- \
-p git


# Install dev dependencies for code completion, definition jumping
RUN \
    go install github.com/cweill/gotests/gotests@latest && \
    go install github.com/fatih/gomodifytags@latest && \
    go install github.com/josharian/impl@latest && \
    go install github.com/haya14busa/goplay/cmd/goplay@latest && \
    go install honnef.co/go/tools/cmd/staticcheck@latest && \
    # --> Go language server
    go install golang.org/x/tools/gopls@latest && \
    # --> Go debugger
    go install github.com/go-delve/delve/cmd/dlv@latest

ENV APP_HOME /root/app
RUN mkdir -p "$APP_HOME"

COPY go.mod /root/app/go.mod
COPY go.sum /root/app/go.sum

WORKDIR "$APP_HOME"


CMD ["tail", "-f", "/dev/null"]