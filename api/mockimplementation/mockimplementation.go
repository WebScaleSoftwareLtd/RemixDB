// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package mockimplementation

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"

	"remixdb.io/api"
)

type impl struct {
	userCount int
	users     map[string][]string
	usersLock sync.Mutex
}

var validIam = regexp.MustCompile(`^[a-zA-Z0-9_\-]+(:[a-zA-Z0-9_\-*]+)+$`)

func strArrayEquals(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range b {
		if a[i] != v {
			return false
		}
	}

	return true
}

func (i *impl) validateUser(ctx api.RequestCtx, perms ...string) (username string, permissions []string, err error) {
	// Defines the unauthorized error.
	unauthorized := api.APIError{
		StatusCode: 401,
		Code:       "unauthorized",
		Message:    "The IAM permissions used to authenticate this request are not valid.",
	}

	// Get the Authorization header.
	authHeader := ctx.GetRequestHeader("Authorization")
	if authHeader == nil {
		return "", nil, unauthorized
	}

	// Split by the first space.
	authHeaderSplit := strings.SplitN(string(authHeader), " ", 2)
	if len(authHeaderSplit) != 2 {
		return "", nil, unauthorized
	}

	// Check the first part is Bearer.
	if strings.ToLower(authHeaderSplit[0]) != "bearer" {
		return "", nil, unauthorized
	}

	// Split the header by comma.
	authHeaderSplit = strings.Split(authHeaderSplit[1], ",")

	// Go through each part.
	for _, part := range authHeaderSplit {
		// Trim the part.
		part = strings.TrimSpace(part)

		// Make sure it is valid.
		if part != "*" && !validIam.MatchString(part) {
			return "", nil, unauthorized
		}
	}

	// Get the user.
	i.usersLock.Lock()
	defer i.usersLock.Unlock()
	for userId, userPerms := range i.users {
		if strArrayEquals(userPerms, perms) {
			return userId, userPerms, nil
		}
	}

	// Create a new user.
	userId := "u" + strconv.Itoa(i.userCount)
	i.userCount++
	i.users[userId] = perms
	return userId, authHeaderSplit, nil
}

func (i *impl) GetServerInfoV1(ctx api.RequestCtx) (api.ServerInfoV1, error) {
	if _, _, err := i.validateUser(ctx); err != nil {
		return api.ServerInfoV1{}, err
	}

	hostname, err := os.Hostname()
	if err != nil {
		return api.ServerInfoV1{}, err
	}

	return api.ServerInfoV1{
		Version:  "v1.2.3",
		Hostname: hostname,
		HostID:   "h0",
		Uptime:   123,
	}, nil
}

func (i *impl) GetSelfUserV1(ctx api.RequestCtx) (api.User, error) {
	username, permissions, err := i.validateUser(ctx)
	if err != nil {
		return api.User{}, err
	}

	return api.User{
		Username:    username,
		Permissions: permissions,
	}, nil
}

// New returns a new mock implementation.
func New() api.APIImplementation {
	return &impl{
		users: map[string][]string{},
	}
}
