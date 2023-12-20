// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";

type Props = {
    title: string;
    description: string;
    checked: boolean;
    setChecked: (checked: boolean) => void;
    loading?: boolean;
};

export default (props: Props) => {
    const id = React.useId();
    return <div className="flex items-center space-x-2">
        <input
            id={id}
            type="checkbox"
            checked={props.checked}
            onChange={e => props.setChecked(e.target.checked)}
            disabled={props.loading}
        />
        <label htmlFor={id} className="flex flex-col">
            <span className="text-sm font-medium text-gray-900">{props.title}</span>
            <span className="text-sm text-gray-500">{props.description}</span>
        </label>
    </div>;
};
