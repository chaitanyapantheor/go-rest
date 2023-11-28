FROM golang:1.20-alpine

WORKDIR .

# Install task locally
RUN go install github.com/go-task/task/v3/cmd/task@latest

# RUN adduser -S -D -H -h /app appuser
# USER appuser

ENV APP_ENV "docker"
ENV GOCACHE /tmp/

# ENTRYPOINT ["tail", "-f", "/dev/null"]
CMD ["task", "run:local-app"]
