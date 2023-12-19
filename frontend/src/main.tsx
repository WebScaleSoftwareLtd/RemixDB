// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import ReactDOM from "react-dom/client";
import routes from "../routes.json";
import * as possibleRoutes from "./routes";
import RootWrapper from "./wrappers/RootWrapper";

Promise.all([
    import("react-router-dom"),
]).then(([{ createBrowserRouter, RouterProvider }]) => {
    // Create the router.
    const router = createBrowserRouter(
        Object.entries(routes).map(([key, val]) => {
            const element = (possibleRoutes as Record<string, React.FunctionComponent>)[val];
            if (!element) {
                throw new Error(`No route element for ${val}`);
            }
            return {
                path: key,
                element: <RootWrapper element={element} />,
            };
        })
    );

    // Render in the root page.
    ReactDOM.createRoot(document.getElementById("root")!).render(
        <React.StrictMode>
            <RouterProvider router={router} />
        </React.StrictMode>
    );
});
