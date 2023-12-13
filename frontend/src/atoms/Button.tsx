// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props = {
    type? : "primary" | "warning" | "danger";
    children: React.ReactNode;
};

export default ({ type, children }: Props) => {
    // Get the classes.
    let className = "w-full text-white font-bold py-2 px-4 rounded ";
    switch (type) {
        case "primary":
            className += "bg-blue-500 hover:bg-blue-600";
            break;
        case "warning":
            className += "bg-yellow-500 hover:bg-yellow-600";
            break;
        case "danger":
            className += "bg-red-500 hover:bg-red-600";
            break;
    }

    // Return the component.
    return <button className={className}>
        {children}
    </button>;
};
