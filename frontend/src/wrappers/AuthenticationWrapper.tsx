// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import * as authState from "../authState";
import Alert from "../atoms/Alert";
import Button from "../atoms/Button";
import Box from "../atoms/Box";
import Title from "../atoms/Title";
import Textbox from "../atoms/Textbox";

const Login = () => {
    // Defines the page state. If it is null, it is loading. If it is a string,
    // it is an error. If it is undefined, it is just rendered.
    const [pageState, setPageState] = React.useState<undefined | null | string>();

    // Defines the state of the API key input.
    const [apiKey, setApiKey] = React.useState("");

    // Handles doing a login.
    const handleLogin = React.useCallback(async () => {
        // Set the page state to loading.
        setPageState(null);

        try {
            // Attempt to log in.
            await authState.login(apiKey);
        } catch (e) {
            // Set the page state to the error.
            setPageState((e as Error).message);
        }
    }, [apiKey]);

    // Return the login page.
    return <div className="flex items-center justify-center h-screen">
        <Box>
            <Title>RemixDB Login</Title>

            {
                pageState && <div className="mb-3">
                    <Alert type="danger">
                        {pageState}
                    </Alert>
                </div>
            }

            <form className="space-y-4" onSubmit={e => {
                e.preventDefault();
                handleLogin();
                return false;
            }}>
                <Textbox
                    title="API Key" value={apiKey} onValueChange={setApiKey}
                    placeholder="Enter your API key" loading={pageState === null}
                />

                <div>
                    <Button type="primary">Sign In</Button>
                </div>
            </form>
        </Box>
    </div>;
};

// The tiny function that basically wraps the authentication state to ensure
// that the user is authenticated before they can access the page. If not,
// we invoke the login page.
export default ({ children }: React.PropsWithChildren) => {
    // Get the current authentication state from fetch information.
    const authHook = React.useSyncExternalStore(
        authState.subscribeAuthenticated,
        () => authState.authenticated,
    );

    // If the user is authenticated, render the children.
    if (authHook) return <>{children}</>;

    // Otherwise, render the login page.
    return <Login />;
};
