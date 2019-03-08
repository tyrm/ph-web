FROM alpine
LABEL maintainer="tyr@pettingzoo.co"

EXPOSE 8080

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

ADD ph-web /app/
ADD static /app/static
ADD templates /app/templates
ADD models/migrations /app/migrations

WORKDIR /app/
CMD ["/app/ph-web"]
