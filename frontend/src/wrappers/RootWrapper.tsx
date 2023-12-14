// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import AuthenticationWrapper from "./AuthenticationWrapper";

type Props = {
    element: React.FunctionComponent;
};

export default ({ element: Element }: Props)  => <AuthenticationWrapper>
    <Element />
</AuthenticationWrapper>;
