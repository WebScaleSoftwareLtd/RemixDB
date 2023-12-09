// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import ReactDOM from "react-dom/client";
import {
    createBrowserRouter,
    RouterProvider,
} from "react-router-dom";
import routes from "../routes.json";
import * as possibleRoutes from "./routes";

const router = createBrowserRouter(Object.entries(routes).map(([key, val]) => {
    const Element = (possibleRoutes as Record<string, React.FunctionComponent>)[val];
    if (!Element) {
        throw new Error(`No route element for ${val}`);
    }
    return {
        path: key,
        element: <Element />,
    };
}));

ReactDOM.createRoot(document.getElementById("root")!).render(
    <React.StrictMode>
        <RouterProvider router={router} />
    </React.StrictMode>
);
