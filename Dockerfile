FROM scratch
LABEL maintainer="tyr@pettingzoo.co"

EXPOSE 8080

ADD ph-web /app/
ADD static /app/static
ADD templates /app/templates

WORKDIR /app/
CMD ["/app/ph-web"]