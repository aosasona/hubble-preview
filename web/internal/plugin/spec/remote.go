package spec

import (
	"errors"
	"fmt"
	"net/url"
	"path"
	"strings"
)

var (
	ErrEmptyURL           = errors.New("url is empty")
	ErrMalformedURL       = errors.New("url is malformed or invalid")
	ErrUnknownAuthMethod  = errors.New("unknown auth method")
	ErrMissingAuth        = errors.New("missing auth")
	ErrInvalidAuthPortion = errors.New("invalid auth portion")
	ErrPathNotProvided    = errors.New("path not provided")
	ErrAuthMethodMismatch = errors.New("auth method mismatch")
	ErrInvalidExtension   = errors.New("invalid extension")
)

//go:generate go tool github.com/abice/go-enum --marshal

// ENUM(ssh,basic_auth,none)
type AuthMethod string

// ENUM(http, https, git, ssh)
type Protocol string

/*
Valid Git urls for this system are:

`ssh` -> `ssh://<username>:<namespace>@<host>:<port>/<path>.git`
`basic_auth` -> `(http(s)|git)://<username>:<password>@<host>:<port>/<path>.git`
`none` -> `(http(s)|git)://<host>:<port>/<path>.git`
*/
type RemoteSource struct {
	rawURL     string
	host, path string
	protocol   Protocol
	// auth is a 2 element array, where the first element is the username and the second element is the password/token
	auth       [2]string
	authMethod AuthMethod
	valid      bool
}

func (r *RemoteSource) Host() string       { return r.host }
func (r *RemoteSource) Path() string       { return r.path }
func (r *RemoteSource) Valid() bool        { return r.valid }
func (r *RemoteSource) Protocol() Protocol { return r.protocol }

func (r *RemoteSource) Name() string {
	// The name is the last part of the path
	parts := strings.Split(r.path, "/")
	return strings.TrimSuffix(parts[len(parts)-1], ".git")
}

func (r *RemoteSource) String() string {
	if r.valid {
		return r.rawURL
	}
	return ""
}

func (r *RemoteSource) RawURL() string {
	return r.rawURL
}

func (r *RemoteSource) FormattedGitURL() string {
	switch r.protocol {
	case ProtocolHttp, ProtocolHttps, ProtocolGit:
		protocol := r.protocol
		// NOTE: `git` protocol is kind of unsupported these days, so we convert to `https` - this behaviour might have to change in the future
		if r.protocol == ProtocolGit {
			protocol = ProtocolHttps
		}

		if r.authMethod == AuthMethodBasicAuth {
			a, _ := r.BasicAuth()
			return fmt.Sprintf(
				"%s://%s:%s@%s/%s",
				protocol, a.Username, a.Password, r.host, r.path,
			)
		}

		return fmt.Sprintf("%s://%s/%s", protocol, r.host, r.path)

	case ProtocolSsh:
		a, _ := r.SshAuth()
		return fmt.Sprintf("%s@%s:%s/%s", a.Username, r.host, a.Namespace, r.path)

	default:
		return fmt.Sprintf("%s://%s/%s", r.protocol, r.host, r.path)
	}
}

type (
	BasicAuth struct{ Username, Password string }

	SshAuth struct {
		// Username is the username for the SSH host e.g. in `git@github.com:username/repo.git`, the username is `git`
		Username string

		// Namespace is the identifier of the target user e.g. `git@github.com:username/repo.git`, the namespace is `username`
		Namespace string
	}
)

func (r *RemoteSource) BasicAuth() (BasicAuth, error) {
	if r.authMethod != AuthMethodBasicAuth {
		return BasicAuth{}, ErrAuthMethodMismatch
	}

	return BasicAuth{Username: r.auth[0], Password: r.auth[1]}, nil
}

func (r *RemoteSource) SshAuth() (SshAuth, error) {
	if r.authMethod != AuthMethodSsh {
		return SshAuth{}, ErrAuthMethodMismatch
	}

	return SshAuth{Username: r.auth[0], Namespace: r.auth[1]}, nil
}

func (r *RemoteSource) UnmarshalText(text []byte) error {
	parsed, err := ParseRemoteSource(string(text))
	if err != nil {
		return err
	}

	*r = parsed
	return nil
}

func (r *RemoteSource) MarshalText() ([]byte, error) {
	if !r.valid {
		return nil, errors.New("invalid remote source")
	}

	return []byte(r.rawURL), nil
}

func ParseRemoteSource(raw string) (RemoteSource, error) {
	if raw == "" {
		return RemoteSource{}, ErrEmptyURL
	}

	if !strings.Contains(raw, "://") {
		return RemoteSource{}, ErrMalformedURL
	}

	switch strings.Split(raw, "://")[0] {
	case "ssh":
		// `ssh://<username>:<namespace>@<host>:<port>/<path>.git`
		raw = strings.TrimPrefix(raw, "ssh://")

		parts := strings.Split(raw, "@")
		if len(parts) != 2 {
			return RemoteSource{}, ErrMalformedURL
		}

		// The first part is the username and namespace
		authParts := strings.Split(parts[0], ":")
		if len(authParts) != 2 {
			return RemoteSource{}, ErrInvalidAuthPortion
		}
		username, namespace := authParts[0], authParts[1]

		// The second part is the host and path}
		slashIdx := strings.Index(parts[1], "/")
		if slashIdx == -1 {
			return RemoteSource{}, ErrPathNotProvided
		}
		host, paths := parts[1][:slashIdx], parts[1][slashIdx+1:]

		if !strings.HasSuffix(paths, ".git") {
			return RemoteSource{}, ErrPathNotProvided
		}

		urlPath := path.Join(strings.Split(paths, "/")...)

		return RemoteSource{
			rawURL:     raw,
			host:       host,
			path:       strings.TrimPrefix(urlPath, "/"),
			protocol:   ProtocolSsh,
			auth:       [2]string{username, namespace},
			authMethod: AuthMethodSsh,
			valid:      true,
		}, nil

	case "http", "https", "git":
		var username, password string
		var authMethod AuthMethod

		u, err := url.Parse(raw)
		if err != nil {
			return RemoteSource{}, ErrMalformedURL
		}

		if u.User != nil {
			username = strings.TrimSpace(u.User.Username())
			password, _ = u.User.Password()
			password = strings.TrimSpace(password)
		}

		switch {
		case username != "" && password != "":
			authMethod = AuthMethodBasicAuth

		case username == "" && password == "":
			authMethod = AuthMethodNone

		default:
			return RemoteSource{}, ErrMalformedURL
		}

		var protocol Protocol
		switch u.Scheme {
		case "https":
			protocol = ProtocolHttps
		case "http":
			protocol = ProtocolHttp
		case "git":
			protocol = ProtocolGit
		}

		if !strings.HasSuffix(u.Path, ".git") {
			return RemoteSource{}, ErrInvalidExtension
		}

		return RemoteSource{
			rawURL:     raw,
			host:       u.Host,
			path:       strings.TrimPrefix(u.Path, "/"),
			auth:       [2]string{username, password},
			protocol:   protocol,
			authMethod: authMethod,
			valid:      true,
		}, nil

	default:
		return RemoteSource{}, ErrUnknownAuthMethod
	}
}
