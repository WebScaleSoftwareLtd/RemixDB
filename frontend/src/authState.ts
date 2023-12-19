// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { useSyncExternalStore } from "react";

// Defines if the user is authenticated.
export let authenticated = false;

// Defines a map of subscribers to the authenticated state. This is used to
// notify all subscribers when the authenticated state changes.
const authenticatedSubscribers = new Map<number, (subbed: boolean) => void>();

// Defines the next ID.
let nextId = 0;

// Allows you to subscribe to the authenticated state. The callback will be run
// whenever the authenticated state changes. The returned function can be used
// to unsubscribe.
export function subscribeAuthenticated(callback: (subbed: boolean) => void) {
    const id = nextId++;
    authenticatedSubscribers.set(id, callback);
    return () => authenticatedSubscribers.delete(id);
}

// Defines the API key used within the application.
let apiKey = "";

// Defines the users login state.
let loginState = {
    username: "",
    permissions: [] as string[],
    sudo_partition: false,
};
const blankLoginState = { ...loginState };

// Logs out of the current session. Note we do not destroy the API key because
// are too manual. Note we do not broadcast the login state because the wrapper
// will do that for us.
export function logout() {
    authenticated = false;
    apiKey = "";
    loginState = blankLoginState;
    authenticatedSubscribers.forEach(sub => sub(false));
}

// Defines a map of subscribers to the login state. This is used to notify all
// subscribers when the login state changes.
const loginStateSubscribers = new Map<number, (state: typeof loginState) => void>();

// Allows you to subscribe to the login state. The callback will be run
// whenever the login state changes. The returned function can be used
// to unsubscribe.
function subscribeLoginState(callback: (state: typeof loginState) => void) {
    const id = nextId++;
    loginStateSubscribers.set(id, callback);
    return () => loginStateSubscribers.delete(id);
}

// Do a login request. This will automatically handle the state if successful.
export async function login(attemptedApiKey: string) {
    // Do the login request.
    const res = await fetch("/api/v1/user", {
        method: "GET",
        headers: {
            Accept: "application/json",
            Authorization: `Bearer ${attemptedApiKey}`,
        },
    });

    // If it is 200, setup the state. We do not inform auth subscribers because
    // the wrapper doesn't use that.
    if (res.status === 200) {
        const data = await res.json();
        loginState = data;
        apiKey = attemptedApiKey;
        authenticated = true;
        authenticatedSubscribers.forEach(sub => sub(true));
        return;
    }

    // Handle any errors.
    const text = await res.text();
    let message: any = null;
    try {
        const data = JSON.parse(text);
        if (data.message) {
            message = data.message;
        }
    } catch {
        // Ignore since this will throw if it isn't a valid API error.
    }
    if (message === null) message = `Status ${res.status}: ${text}`;
    throw new Error(`${message}`);
}

// Defines the fetch client. This automatically manages a lot of the state for us.
export function authManagedFetch(input: RequestInfo, init?: RequestInit) {
    // Setup fetch to use the API key.
    let initCpy: RequestInit = {};
    if (init) {
        initCpy = { ...init };
    }
    initCpy.headers = {
        ...initCpy.headers,
        Authorization: `Bearer ${apiKey}`,
    };
    const res = fetch(input, initCpy);

    // Handle any errors that impact the authenticated state.
    res.then(res => {
        if (res.status === 401) {
            // We are no longer authenticated.
            apiKey = "";
            loginState = blankLoginState;
            authenticated = false;
            authenticatedSubscribers.forEach(sub => sub(false));
        } else if (res.status === 403) {
            // We are authenticated but don't have permission. Check the error
            // X-RemixDB-Permissions header to get the updated permissions.
            const perms = res.headers.get("X-RemixDB-Permissions");
            if (perms) {
                loginState.permissions = perms.split(",");
                loginStateSubscribers.forEach(sub => sub(loginState));
            }
        }
    });

    // Return the response.
    return res;
}

// Defines a helper hook that allows a component to use the permissions.
export function usePermissions() {
    useSyncExternalStore(subscribeLoginState, () => loginState.permissions);
    return loginState.permissions;
}

// Defines a helper hook that allows a component to use the username.
export function useUsername() {
    useSyncExternalStore(subscribeLoginState, () => loginState.username);
    return loginState.username;
}

// Defines a helper hook that allows a component to use the sudo partition.
export function useSudoPartition() {
    useSyncExternalStore(subscribeLoginState, () => loginState.sudo_partition);
    return loginState.sudo_partition;
}
