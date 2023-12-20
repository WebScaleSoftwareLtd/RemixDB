// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
    return twMerge(clsx(inputs));
}

export function alwaysPreventDefault(hn: () => void) {
    return (e: React.SyntheticEvent) => {
        e.preventDefault();
        hn();
        return false;
    };
}
