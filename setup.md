###Project setup###

go get github.com/go-martini/martini

go get github.com/martini-contrib/binding

go get github.com/jinzhu/gorm

go get github.com/lib/pq

go get github.com/gosimple/conf

go get code.google.com/p/go-uuid/uuid

go get code.google.com/p/go.crypto/bcrypt

go get github.com/martini-contrib/render

go get github.com/stretchr/graceful


####To run the server####

(For initializing db)

go run *.go -initdb=true

####Test dependencies####

go get github.com/axw/gocov/gocov

go get gopkg.in/matm/v1/gocov-html

go get gopkg.in/check.v1

To generate coverage report run "gocov test | gocov-html > coverage.html"


