#!/bin/bash -x

rm -Rf ph-web

go get github.com/antonlindstrom/pgstore
go get github.com/gobuffalo/genny
go get github.com/gobuffalo/packr/v2/...
go get github.com/gobuffalo/packr/v2/packr2
go get github.com/juju/loggo
go get github.com/patrickmn/go-cache
go get github.com/tdewolff/minify
go get golang.org/x/crypto/bcrypt

packr2 clean
packr2 install
packr2 build -v -o ph-web .
packr2 clean