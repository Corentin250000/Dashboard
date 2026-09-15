# ---- Build ----
FROM golang:1.22-alpine AS builder
WORKDIR /app

# Deps first (better layer caching)
COPY go.mod ./
RUN go mod download 2>/dev/null || true

COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /server ./cmd/server

# ---- Run ----
FROM alpine:3.20
# curl + gnupg : requis par le script d'install Doppler (curl || wget, + gpg pour la signature).
RUN apk add --no-cache ca-certificates curl gnupg
# CLI Doppler officielle (sans gestionnaire de paquets).
# Si DOPPLER_TOKEN est fourni au conteneur, `doppler run` injecte les secrets ;
# sinon il laisse passer la commande telle quelle (cf. README "Gestion des secrets").
RUN (curl -Ls --tlsv1.2 --proto "=https" --retry 3 https://cli.doppler.com/install.sh || wget -t 3 -qO- https://cli.doppler.com/install.sh) | sh
WORKDIR /app
COPY --from=builder /server /server
COPY entrypoint.sh /app/entrypoint.sh
RUN chmod +x /app/entrypoint.sh
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["/app/entrypoint.sh"]
CMD ["doppler", "run", "--", "/server"]
