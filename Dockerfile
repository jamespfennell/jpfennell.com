FROM alpine:3 AS builder

ARG HUGO_VERSION=0.167.0
ARG TARGETARCH
RUN wget -qO- https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_linux-${TARGETARCH:-amd64}.tar.gz \
    | tar -xz -C /usr/local/bin hugo

WORKDIR /website
COPY . .
RUN hugo

FROM nginx
COPY --from=builder /website/public /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
