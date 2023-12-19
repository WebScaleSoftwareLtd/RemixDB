// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

export default ({ children }: React.PropsWithChildren) => {
    return <div className="flex flex-col items-center justify-center">{children}</div>;
};
