FROM debian:stable-slim

RUN apt update && apt install -y gcc libfreetype-dev ffmpeg ca-certificates

RUN mkdir -p /opt/audio-search-bot/var/metadata && chown -R nobody /opt/audio-search-bot

RUN addgroup nobody

ARG USER=nobody
USER nobody

WORKDIR /opt/audio-search-bot

COPY --chown=nobody bin/audio-search-bot .

RUN chmod +x audio-search-bot

CMD ["/opt/audio-search-bot/audio-search-bot", "bot"]
