FROM alpine:3 AS builder

ARG HUGO_VERSION=0.167.0
ARG TARGETARCH
RUN wget -qO- https://github.com/gohugoio/hugo/releases/download/v${HUGO_VERSION}/hugo_${HUGO_VERSION}_linux-${TARGETARCH:-amd64}.tar.gz \
    | tar -xz -C /usr/local/bin hugo

WORKDIR /website
COPY . .
RUN hugo
# The node template is only used by the node image, not served on the website.
RUN mv public/node-template/index.html /node-template.html && rmdir public/node-template

FROM golang:1.25 AS node-builder
WORKDIR /build
COPY node/go.mod node/go.sum ./
RUN go mod download
COPY node/*.go ./
RUN CGO_ENABLED=0 go build -o /node .

# Status page for each of my servers; see node/README.md.
# Build with --target node.
FROM gcr.io/distroless/static AS node
COPY --from=node-builder /node /node/node
COPY --from=builder /node-template.html /node/template.html
COPY --from=builder /website/public/normalize.css /website/public/stylesheet.css /node/static/
WORKDIR /node
ENTRYPOINT ["/node/node"]

# The website. This is the last stage so that it's the default target.
FROM nginx AS website
COPY --from=builder /website/public /usr/share/nginx/html
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
