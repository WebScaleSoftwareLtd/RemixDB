// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props = {
    type?: string;
    loading?: boolean;
    onValueChange: (value: string) => void;
    value: string;
    title: string;
    placeholder: string;
};

export default (props: Props) => {
    const id = React.useId();
    return <div>
        <label htmlFor={id} className="block text-gray-800 font-bold mb-2">
            {props.title}:
        </label>
        <input
            type={props.type || "text"}
            id={id}
            value={props.value}
            onChange={e => props.onValueChange(e.target.value)}
            className="w-full border border-gray-300 p-2 rounded"
            placeholder={props.placeholder}
            disabled={props.loading}
        />
    </div>;
};
