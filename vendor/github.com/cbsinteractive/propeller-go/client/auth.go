package client

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// Auth holds credentials for authenticating with
// the Propeller API
type Auth struct {
	User string
	Pass string
	Host string
}

// Apply attaches basic authentication to req if and only if
// the User is set and a.Host matches the request URL
// hostname.
func (a *Auth) Apply(req *http.Request) bool {
	if a == nil {
		return false
	}
	if len(a.User) == 0 {
		return false
	}
	if req.URL.Hostname() != a.Host {
		return false
	}
	req.SetBasicAuth(a.User, a.Pass)
	return true
}

func (a Auth) String() string {
	if a == (Auth{}) {
		return ""
	}
	if a.User == "" {
		return a.Host
	}
	return fmt.Sprintf("%s:%s@%s", a.User, a.Pass, a.Host)
}

// NewAuth reads the credentials and url, returning an
// Auth that attaches those credentails to an http request
// matching the host given in url_ using Basic authentication.
//
// credentials should be a colon-seperated username and
// password, url_ should be a url or hostname.
func NewAuth(credentials, url_ string) (Auth, error) {
	a := [2]string{}
	copy(a[:], strings.Split(credentials, ":"))

	u, err := url.Parse(url_)
	if err != nil {
		return Auth{}, err
	}

	return Auth{
		User: a[0],
		Pass: a[1],
		Host: u.Hostname(),
	}, nil
}
