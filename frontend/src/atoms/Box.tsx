// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

export default ({ children }: React.PropsWithChildren) => {
    return <div className="bg-white max-w-md p-8 rounded shadow-lg">
        {children}
    </div>;
};
