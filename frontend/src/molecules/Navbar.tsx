// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import Button from "@/atoms/Button";
import { useUsername, logout, usePermissions, useSudoPartition } from "@/authState";
import { alwaysPreventDefault } from "@/utils";
import { Link } from "react-router-dom";

// Defines the navigation menu import.
const navMenu = import("@/shadcn/ui/navigation-menu");
const NavigationMenuLink = React.lazy(() =>
    navMenu.then(m => ({ default: m.NavigationMenuLink }))
);
const NavigationMenuList = React.lazy(() =>
    navMenu.then(m => ({ default: m.NavigationMenuList }))
);
const NavigationMenu = React.lazy(() =>
    navMenu.then(m => ({ default: m.NavigationMenu }))
);

type Route = {
    to: string;
    name: string;
    icon: React.ReactNode;
    hasPermission: (permissions: string[], sudoPartition: boolean) => boolean;
};

const possibleRoutes: Route[] = [
    {
        to: "/users",
        name: "Users",
        icon: <UsersIcon className="h-4 w-4 mr-2" />,
        hasPermission: (permissions: string[]) =>
            permissions.find(x => {
                return x === "*" || x.startsWith("users:");
            }) !== undefined,
    },
    {
        to: "/contracts",
        name: "Contracts",
        icon: <FunctionSquareIcon className="h-4 w-4 mr-2" />,
        hasPermission: (permissions: string[]) =>
            permissions.find(x => {
                return x === "*" || x.startsWith("contracts:");
            }) !== undefined,
    },
    {
        to: "/migrations",
        name: "Migrations",
        icon: <MoveIcon className="h-4 w-4 mr-2" />,
        hasPermission: (permissions: string[]) =>
            permissions.find(x => {
                return x === "*" || x.startsWith("migrations:");
            }) !== undefined,
    },
    {
        to: "/structures",
        name: "Structures",
        icon: <BuildingIcon className="h-4 w-4 mr-2" />,
        hasPermission: (permissions: string[]) =>
            permissions.find(x => {
                return x === "*" || x.startsWith("structs:");
            }) !== undefined,
    },
    {
        to: "/servers",
        name: "Servers",
        icon: <ServerIcon className="h-4 w-4 mr-2" />,
        hasPermission: (permissions: string[], sudoPartition: boolean) =>
            sudoPartition &&
            permissions.find(x => {
                return x === "*" || x.startsWith("servers:");
            }) !== undefined,
    },
];

export default () => {
    // Defines the username that is currently logged in.
    const username = useUsername();

    // Defines the users current permissions.
    const permissions = usePermissions();

    // Check if this is a sudo partition.
    const sudoPartition = useSudoPartition();

    // Defines the navigation menu items.
    const menuItems: React.ReactNode[] = [];

    // Go through each of the menu items.
    for (let i = 0; i < possibleRoutes.length; i++) {
        const route = possibleRoutes[i];
        if (route.hasPermission(permissions, sudoPartition)) {
            menuItems.push(
                <NavigationMenuLink asChild key={i}>
                    <Link
                        className="group inline-flex h-9 w-max items-center justify-center rounded-md bg-white px-4 py-2 text-sm font-medium transition-colors hover:bg-gray-100 hover:text-gray-900 focus:bg-gray-100 focus:text-gray-900 focus:outline-none disabled:pointer-events-none disabled:opacity-50 data-[active]:bg-gray-100/50 data-[state=open]:bg-gray-100/50 dark:bg-gray-950 dark:hover:bg-gray-800 dark:hover:text-gray-50 dark:focus:bg-gray-800 dark:focus:text-gray-50 dark:data-[active]:bg-gray-800/50 dark:data-[state=open]:bg-gray-800/50"
                        to={route.to}
                    >
                        {route.icon}
                        {route.name}
                    </Link>
                </NavigationMenuLink>
            );
        }
    }

    // Return all of the items.
    return <header className="flex h-20 w-full items-center px-4 md:px-6">
        <Link className="mr-6 hidden lg:flex items-center" to="/">
            <span className="font-bold text-lg">RemixDB</span>
        </Link>
        <React.Suspense>
            <NavigationMenu className="hidden lg:flex">
                <NavigationMenuList>{menuItems}</NavigationMenuList>
            </NavigationMenu>
        </React.Suspense>

        <div className="ml-auto flex items-center gap-2">
            <span className="text-gray-600 dark:text-gray-400 mr-2">{username}</span>
            <form onSubmit={alwaysPreventDefault(logout)}>
                <Button type="outline">Logout</Button>
            </form>
        </div>
    </header>;
};

function BuildingIcon(props: React.SVGProps<SVGSVGElement>) {
    return <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <rect width="16" height="20" x="4" y="2" rx="2" ry="2" />
        <path d="M9 22v-4h6v4" />
        <path d="M8 6h.01" />
        <path d="M16 6h.01" />
        <path d="M12 6h.01" />
        <path d="M12 10h.01" />
        <path d="M12 14h.01" />
        <path d="M16 10h.01" />
        <path d="M16 14h.01" />
        <path d="M8 10h.01" />
        <path d="M8 14h.01" />
    </svg>;
}

function FunctionSquareIcon(props: React.SVGProps<SVGSVGElement>) {
    return <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <rect width="18" height="18" x="3" y="3" rx="2" ry="2" />
        <path d="M9 17c2 0 2.8-1 2.8-2.8V10c0-2 1-3.3 3.2-3" />
        <path d="M9 11.2h5.7" />
    </svg>;
}

function MoveIcon(props: React.SVGProps<SVGSVGElement>) {
    return <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <polyline points="5 9 2 12 5 15" />
        <polyline points="9 5 12 2 15 5" />
        <polyline points="15 19 12 22 9 19" />
        <polyline points="19 9 22 12 19 15" />
        <line x1="2" x2="22" y1="12" y2="12" />
        <line x1="12" x2="12" y1="2" y2="22" />
    </svg>;
}

function ServerIcon(props: React.SVGProps<SVGSVGElement>) {
    return <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <rect width="20" height="8" x="2" y="2" rx="2" ry="2" />
        <rect width="20" height="8" x="2" y="14" rx="2" ry="2" />
        <line x1="6" x2="6.01" y1="6" y2="6" />
        <line x1="6" x2="6.01" y1="18" y2="18" />
    </svg>;
}

function UsersIcon(props: React.SVGProps<SVGSVGElement>) {
    return <svg
        {...props}
        xmlns="http://www.w3.org/2000/svg"
        width="24"
        height="24"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
    >
        <path d="M16 21v-2a4 4 0 0 0-4-4H6a4 4 0 0 0-4 4v2" />
        <circle cx="9" cy="7" r="4" />
        <path d="M22 21v-2a4 4 0 0 0-3-3.87" />
        <path d="M16 3.13a4 4 0 0 1 0 7.75" />
    </svg>;
}
