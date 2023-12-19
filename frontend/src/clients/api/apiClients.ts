// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { MetricsResponse } from "./ifaces";
import apiFetcher from "./fetch";

export class APIV1Client {
    static metrics(): Promise<MetricsResponse> {
        return apiFetcher("/api/v1/metrics");
    }
}
