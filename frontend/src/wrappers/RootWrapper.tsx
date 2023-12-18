// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import AuthenticationWrapper from "./AuthenticationWrapper";
import Navbar from "@/molecules/Navbar";

type Props = {
    element: React.FunctionComponent;
};

export default ({ element: Element }: Props) => <AuthenticationWrapper>
    <Navbar />
    <Element />
</AuthenticationWrapper>;
