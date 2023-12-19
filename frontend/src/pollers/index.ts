// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { APIV1Client } from "@/clients/api";
import Poller from "./Poller";

// Defines the poller for metrics data from the API.
export const metricsPoller = new Poller(() => {
    return APIV1Client.metrics();
}, 2000);
