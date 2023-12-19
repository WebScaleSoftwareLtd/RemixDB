// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { authManagedFetch } from "@/authState";
import { APIError } from "./ifaces";

// Used for 401's where flashing an error would be annoying. Inifinitely in a un-resolved state.
const infinitePromise = new Promise(() => {});

// Does a API request. This will automatically handle the state.
export default async function doApiRequest<T>(
    path: string,
    method?: string,
    body?: any
): Promise<T> {
    // Do the actual request.
    const res = await authManagedFetch(path, {
        method: method || "GET",
        headers: {
            Accept: "application/json",
            "Content-Type": "application/json",
        },
        body: body ? JSON.stringify(body) : undefined,
    });

    if (!res.ok) {
        // If it was a 401, return a promise that will never resolve.
        if (res.status === 401) return infinitePromise as Promise<T>;

        // Get the text response of what was sent back.
        const text = await res.text();

        let err: APIError;
        try {
            // Check if it is an API error.
            err = JSON.parse(text) as APIError;
            if (typeof err.code !== "string" || typeof err.message !== "string") {
                throw 1;
            }
        } catch {
            // If it isn't an API error, throw a generic error.
            throw new Error(
                `Failed to ${method || "GET"} ${path}: ${res.status} ${
                    res.statusText
                } - ${text}`
            );
        }
        throw err;
    }

    // Parse the response and return it.
    return res.json() as Promise<T>;
}
