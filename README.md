requests
========

A request unmarshaller for web servers using the Go language.

## Goals

1. Use Content-Type to load request bodies.
2. Provide methods for attempting to make all data loaded from a body
   function roughly the same (e.g. if only one value was provided for
   an input, store just the value instead of []string{value}, to make
   `application/x-www-urlencoded` and `application/json` work a little
   more similarly).
3. Keep track of all input errors and return them in an easy to parse
   way, for providing details to the user about what went wrong.
4. Provide functionality similar to json's Unmarshal process for
   generic web requests, regardless of Content-Type.

## State

The actual *logic* within this repository has been in production use
with a few projects for a while, now.  However, this project is a
massive refactor of that logic, and I haven't got any test coverage
for it yet.  You may want to play around with it, but I would suggest
waiting until there is at least 80% test coverage before using it
actively in production, yourself.

## Summary

You will likely want to read the
[package documentation](http://godoc.org/github.com/go-requests/requests)
to get the most out of this package.  However, the absolute basics
are:

```go
import "github.com/go-requests/requests"

type User struct {
	Name string `request:"username,required"`
    Pass password `request:"password,required"`
    Blurb string `request:"about_me"`
}

func HandleRequest(httpRequest *http.Request) error {
	target := new(User)
	return requests.New(httpRequest).Unmarshal(target)
}
```

The request body will be loaded into a map of parameters and then
values from the request will be applied to fields on the target user.
If either username or password are missing from the request, an error
will be returned.

## Installing

For projects with solid unit testing, or projects intending to follow
continuous integration, I recommend getting directly from github:

```
go get github.com/go-requests/requests
```

However, for projects that need more assurances that *nothing* will
*ever* change, this project does support versioning through gopkg.in:

```
go get gopkg.in/requests.v0
```

Check the project tags for version updates.  v1 will be released when
test coverage exceeds 80%.
