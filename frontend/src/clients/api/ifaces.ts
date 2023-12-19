// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

// Defines the structure of an API error.
export interface APIError {
    code: string;
    message: string;
}

// Defines the response with metric information.
export interface MetricsResponse {
    cpu_percent: number;
    ram_mb: number;
    goroutines: number;
    gcs: number;
}
