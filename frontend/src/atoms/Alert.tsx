// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props = {
    type: "primary" | "secondary" | "danger";
    children: React.ReactNode;
};

export default ({ type, children }: Props) => {
    // Get the classes.
    let className = "px-4 py-2 rounded text-white";
    switch (type) {
        case "primary":
            className += " bg-blue-500";
            break;
        case "secondary":
            className += " bg-gray-500";
            break;
        case "danger":
            className += " bg-red-500";
            break;
    }

    // Return the component.
    return <div className={className}>{children}</div>;
};
