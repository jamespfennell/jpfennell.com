FROM golang:1.16 AS builder

WORKDIR /hugo
RUN curl -L -o hugo.tar.gz https://github.com/gohugoio/hugo/releases/download/v0.84.0/hugo_0.84.0_Linux-64bit.tar.gz
RUN tar -zxvf hugo.tar.gz
RUN mv hugo /usr/bin

WORKDIR /website
COPY . .
RUN hugo

FROM nginx
COPY --from=builder /website/public /usr/share/nginx/html
RUN ls /usr/share/nginx/html
RUN cat /etc/nginx/conf.d/default.conf
# COPY ./nginx.conf /etc/nginx/conf.d/realtimerail.nyc-nginx.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
