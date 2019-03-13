#!/bin/bash -x

rm -Rf ph-web

go get github.com/antonlindstrom/pgstore
go get github.com/bamzi/jobrunner
go get github.com/dustin/go-humanize
go get github.com/eefret/gravatar
go get github.com/go-telegram-bot-api/telegram-bot-api
go get github.com/gobuffalo/genny
go get github.com/gobuffalo/packr/v2/...
go get github.com/gobuffalo/packr/v2/packr2
go get github.com/gorilla/mux
go get github.com/juju/loggo
go get github.com/minio/minio-go
go get github.com/patrickmn/go-cache
go get github.com/rubenv/sql-migrate
go get golang.org/x/crypto/bcrypt
go get gopkg.in/alexcesaro/statsd.v2

#packr2 clean
#packr2 install
#packr2 build -o ph-web .

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ph-web .
