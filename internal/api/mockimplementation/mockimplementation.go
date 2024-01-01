// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

package mockimplementation

import (
	"encoding/json"
	"os"
	"regexp"
	"runtime"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"remixdb.io/internal/api"
)

type impl struct {
	userCount int
	users     map[string][]string
	usersLock sync.Mutex

	partitionSetup uintptr
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
	// Handle is perms is nil.
	if perms == nil {
		perms = []string{}
	}

	// Handle if the partition is not setup.
	if os.Getenv("MOCK_PARTITION_STATE") == "setup_required" && atomic.LoadUintptr(&i.partitionSetup) == 0 {
		return "", nil, api.APIError{
			StatusCode: 400,
			Code:       "partition_not_setup",
			Message:    "The partition is not setup.",
		}
	}

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
		SudoPartition: !i.isNonSudo(),
		Username:      username,
		Permissions:   permissions,
	}, nil
}

func (i *impl) isNonSudo() bool {
	loaded := atomic.LoadUintptr(&i.partitionSetup)
	if loaded != 0 {
		return loaded == 1
	}

	return os.Getenv("MOCK_PARTITION_STATE") == "nonsudo"
}

func (i *impl) GetMetricsV1(ctx api.RequestCtx) (api.MetricsV1, error) {
	_, _, err := i.validateUser(ctx)
	if err != nil {
		return api.MetricsV1{}, err
	}

	if i.isNonSudo() {
		return api.MetricsV1{}, api.APIError{
			StatusCode: 400,
			Code:       "sudo_required",
			Message:    "The sudo_partition permission is required to access this endpoint.",
		}
	}

	var gcStats debug.GCStats
	debug.ReadGCStats(&gcStats)

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	return api.MetricsV1{
		RAMMegabytes: memStats.Alloc / 1024 / 1024,
		Goroutines:   runtime.NumGoroutine(),
		GCS:          int(gcStats.NumGC),
	}, nil
}

func (i *impl) GetPartitionCreatedStateV1(ctx api.RequestCtx) (bool, error) {
	if os.Getenv("MOCK_PARTITION_STATE") == "setup_required" &&
		atomic.LoadUintptr(&i.partitionSetup) == 0 {
		return false, nil
	}

	return true, nil
}

func (i *impl) CreatePartitionV1(ctx api.RequestCtx) (string, error) {
	// Get the body.
	var body api.CreatePartitionV1Body
	if err := json.Unmarshal(ctx.GetRequestBody(), &body); err != nil {
		return "", api.APIError{
			StatusCode: 400,
			Code:       "invalid_body",
			Message:    "The body is invalid.",
		}
	}

	// Handle if we are pretending to be a partition that is already setup.
	if atomic.LoadUintptr(&i.partitionSetup) != 0 {
		return "", api.APIError{
			StatusCode: 400,
			Code:       "partition_already_exists",
			Message:    "The partition already exists.",
		}
	}

	// Handle if the sudo key is invalid.
	if body.SudoAPIKey != "sudo" {
		return "", api.APIError{
			StatusCode: 400,
			Code:       "invalid_sudo_key",
			Message:    "The sudo key is invalid.",
		}
	}

	// Handle if the username is invalid.
	if body.Username != "username" {
		return "", api.APIError{
			StatusCode: 400,
			Code:       "invalid_username",
			Message:    "The username is invalid.",
		}
	}

	// Do the atomic operation.
	val := uintptr(1)
	if body.SudoPartition {
		val = 2
	}
	atomic.CompareAndSwapUintptr(&i.partitionSetup, 0, val)

	// Return no errors.
	return "*", nil
}

// New returns a new mock implementation.
func New() api.APIImplementation {
	return &impl{
		users: map[string][]string{},
	}
}
