// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props = {
    large?: boolean;
    margin?: boolean;
}

export default ({ children, large, margin }: React.PropsWithChildren<Props>) => {
    let classes = "bg-gray-50 p-8 rounded shadow-lg";
    if (!large) classes += " max-w-md";
    if (margin) classes += " mt-4";
    return <div className={classes}>{children}</div>;
};
