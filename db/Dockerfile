FROM postgres:16-alpine

RUN apk add --no-cache ca-certificates curl gnupg

RUN (curl -Ls --tlsv1.2 --proto "=https" --retry 3 https://cli.doppler.com/install.sh || wget -t 3 -qO- https://cli.doppler.com/install.sh) | sh

COPY entrypoint.sh /usr/local/bin/doppler-entrypoint.sh
RUN chmod +x /usr/local/bin/doppler-entrypoint.sh

ENTRYPOINT ["doppler-entrypoint.sh"]
CMD ["postgres"]
