// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props = {
    type? : "outline" | "primary" | "warning" | "danger";
    children: React.ReactNode;
};

export default ({ type, children }: Props) => {
    // Get the classes.
    let className = "w-full font-bold py-2 px-4 rounded ";
    switch (type) {
    case "outline":
        className += "bg-transparent hover:bg-black hover:text-white border border-black";
        break;
    case "primary":
        className += "text-white bg-blue-500 hover:bg-blue-600";
        break;
    case "warning":
        className += "text-white bg-yellow-500 hover:bg-yellow-600";
        break;
    case "danger":
        className += "text-white bg-red-500 hover:bg-red-600";
        break;
    }

    // Return the component.
    return <button className={className}>
        {children}
    </button>;
};
