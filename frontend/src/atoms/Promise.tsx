// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props<T> = {
    children: React.ReactNode;
    loading: React.ReactNode;
    error: React.FunctionComponent<{ error: Error }>;
    setValue: (value: T) => void;
    promise: Promise<T>;
};

export default <T,>(props: Props<T>) => {
    // Keep information about the state of the promise.
    const [state, setState] = React.useState<undefined | Error | true>();
    React.useEffect(() => {
        props.promise.then(
            value => {
                props.setValue(value);
                setState(true);
            },
            error => {
                setState(error);
            }
        );
    }, [props.promise, props.setValue]);

    // If it is undefined, return the loading component.
    if (state === undefined) return props.loading;

    // If it is true, return the children.
    if (state === true) return props.children;

    // Otherwise, return the error component.
    return <props.error error={state} />;
};
