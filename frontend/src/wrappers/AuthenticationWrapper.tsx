// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import * as authState from "@/authState";
import Alert from "@/atoms/Alert";
import Button from "@/atoms/Button";
import Box from "@/atoms/Box";
import Title from "@/atoms/Title";
import Textbox from "@/atoms/Textbox";
import Center from "@/atoms/Center";
import Spinner from "@/atoms/Spinner";
import ErrorView from "@/atoms/ErrorView";
import Checkbox from "@/atoms/Checkbox";
import { alwaysPreventDefault } from "@/utils";

// Handles the standard authentication flow for users who already have a partition.
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

            {pageState && <div className="mb-3">
                <Alert type="danger">{pageState}</Alert>
            </div>}

            <form
                className="space-y-4"
                onSubmit={alwaysPreventDefault(handleLogin)}
                data-login-form
            >
                <Textbox
                    title="API Key"
                    value={apiKey}
                    onValueChange={setApiKey}
                    placeholder="Enter your API key"
                    loading={pageState === null}
                />

                <div>
                    <Button type="primary">Sign In</Button>
                </div>
            </form>
        </Box>
    </div>;
};

// Handles the creation of a partition.
const CreatePartition = (props: { setState: (v: true) => void }) => {
    // Defines the page state. If it is null, it is loading. If it is undefined, it is just
    // rendered. If it is a string, it is the key. If it is a Error, it is an error.
    const [pageState, setPageState] = React.useState<undefined | null | Error | string>();

    // Defines the sudo API key.
    const [apiKey, setApiKey] = React.useState("");

    // Defines the username.
    const [username, setUsername] = React.useState("");

    // Defines if this is a sudo partition.
    const [sudoPartition, setSudoPartition] = React.useState(false);

    // Handle the setup.
    const doSetup = React.useCallback(async () => {
        // Set the page state to loading.
        setPageState(null);

        try {
            // Attempt to setup the partition.
            const res = await fetch("/api/v1/partition/create", {
                method: "POST",
                headers: {
                    "Content-Type": "application/json",
                },
                body: JSON.stringify({
                    sudo_api_key: apiKey,
                    username,
                    sudo_partition: sudoPartition,
                }),
            });

            // Handle any errors.
            if (!res.ok) {
                // Attempt to parse it into JSON.
                const text = await res.text();
                let val: {
                    message: string;
                    code: string;
                };
                try {
                    // Attempt to parse as an error JSON.
                    val = JSON.parse(text);
                    if (typeof val.message !== "string" || typeof val.code !== "string") {
                        throw 1;
                    }
                } catch {
                    // Error with the status code and text.
                    throw new Error(`Received ${res.status}: ${text}`);
                }

                // If the code is partition_already_exists, we can just skip this step.
                if (val.code === "partition_already_exists") {
                    return props.setState(true);
                }

                // Throw the error.
                throw new Error(val.message);
            }

            // Set the page state to the string.
            setPageState(await res.json());
        } catch (e) {
            // Set the page state to the error.
            setPageState(e as Error);
        }
    }, [apiKey, username, sudoPartition, props.setState]);

    // If the page state is a string, show a prompt with the key.
    if (typeof pageState === "string") {
        return <div className="flex items-center justify-center h-screen">
            <Box>
                <Title>Partition Created</Title>
                <p>
                    The partition has been created. You can now log in to a user with all
                    permissions using the following API key:
                </p>

                <p className="my-3">
                    <code className="bg-gray-100 dark:bg-gray-800 p-2 rounded-md">
                        {pageState}
                    </code>
                </p>

                <form onSubmit={alwaysPreventDefault(() => props.setState(true))}>
                    <Button type="primary">Continue to Login</Button>
                </form>
            </Box>
        </div>;
    }

    // Return the creation page.
    return <div className="flex items-center justify-center h-screen">
        <Box>
            <Title>Setup Partition</Title>
            <p className="mb-3">
                RemixDB is not configured for this partition. If you are a systems
                administrator for this cluster, you can create a partition here.
            </p>

            {pageState && <div className="mb-3">
                <Alert type="danger">{pageState.message}</Alert>
            </div>}

            <form className="space-y-4" onSubmit={alwaysPreventDefault(doSetup)}>
                <Textbox
                    title="Sudo API Key"
                    value={apiKey}
                    onValueChange={setApiKey}
                    placeholder="Enter your sudo API key"
                    loading={pageState === null}
                />

                <Textbox
                    title="Username"
                    value={username}
                    onValueChange={setUsername}
                    placeholder="Enter your username"
                    loading={pageState === null}
                />

                <Checkbox
                    title="Sudo Partition"
                    description="If this is a sudo partition, you will have access to cluster management from this partition."
                    checked={sudoPartition}
                    setChecked={setSudoPartition}
                    loading={pageState === null}
                />

                <div>
                    <Button type="primary">Create Partition</Button>
                </div>
            </form>
        </Box>
    </div>;
};

// Switches between the view to create a partition and the view to login.
const CreateOrLogin = () => {
    // Defines the state.
    const [state, setState] = React.useState<undefined | boolean | Error>();

    // Handles getting the result.
    React.useEffect(() => {
        fetch("/api/v1/partition/created")
            .then(async x => {
                if (!x.ok) throw new Error(`Received ${x.status} ${x.statusText}`);
                setState(await x.json());
            })
            .catch(setState);
    }, []);

    // If the state is undefined, we are loading.
    if (state === undefined) {
        return <Center>
            <Spinner />
        </Center>;
    }

    // If the state is true, we show the login page.
    if (state === true) return <Login />;

    // If the state is false, we show the create page.
    if (state === false) return <CreatePartition setState={setState} />;

    // Otherwise, we show the error.
    return <Center>
        <ErrorView error={state} />
    </Center>;
};

// The tiny function that basically wraps the authentication state to ensure that
// the user is authenticated before they can access the page. If not, we invoke
// the login page.
export default ({ children }: React.PropsWithChildren) => {
    // Get the current authentication state from fetch information.
    const authHook = React.useSyncExternalStore(
        authState.subscribeAuthenticated,
        () => authState.authenticated
    );

    // If the user is authenticated, render the children.
    if (authHook) return <>{children}</>;

    // Otherwise, render the create or login view.
    return <CreateOrLogin />;
};
