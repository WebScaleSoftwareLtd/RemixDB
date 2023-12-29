// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { test, expect, Page, ElementHandle } from "@playwright/test";
import { initTest, screenshotErrors } from "./helpers";

test.describe("navbar", () => {
    screenshotErrors();

    // Defines the default test initialization setup.
    const defaultSetup = async (
        page: Page,
        permissions: string[],
        sudo: boolean,
        hidden: boolean
    ) => {
        // Setup the test with our helper.
        await initTest({
            page,
            path: "/",
            permissions,
            sudoPartition: sudo,
        });

        // Wait for the navbar to load since it is imported async.
        await page.waitForSelector("[data-navbar]", {
            state: hidden ? "attached" : "visible",
        });
    };

    // Checks the elements match what is in the objects specified in the same order.
    const elementsMatch = async (
        links: ElementHandle<HTMLElement>[],
        expects: {
            name: string;
            path: string;
        }[]
    ) => {
        // Check the length matches.
        expect(links.length).toBe(expects.length);

        // Loop through the links.
        for (let i = 0; i < links.length; i++) {
            // Get the link.
            const link = links[i];

            // Get the expected link.
            const expected = expects[i];

            // Check the text matches.
            expect(await link.textContent()).toBe(expected.name);

            // Check the href matches.
            expect(await link.getAttribute("href")).toBe(expected.path);
        }
    };

    test("all non-sudo permissions", async ({ page }) => {
        // Load the page.
        await defaultSetup(page, ["*"], false, false);

        // Get the links from the navbar.
        const links = (await page.$$("[data-navbar] a")) as ElementHandle<HTMLElement>[];

        // Validate the links.
        await elementsMatch(links, [
            { name: "Users", path: "/users" },
            { name: "Contracts", path: "/contracts" },
            { name: "Migrations", path: "/migrations" },
            { name: "Structures", path: "/structures" },
        ]);
    });

    test("all sudo permissions", async ({ page }) => {
        // Load the page.
        await defaultSetup(page, ["*"], true, false);

        // Get the links from the navbar.
        const links = (await page.$$("[data-navbar] a")) as ElementHandle<HTMLElement>[];

        // Validate the links.
        await elementsMatch(links, [
            { name: "Users", path: "/users" },
            { name: "Contracts", path: "/contracts" },
            { name: "Migrations", path: "/migrations" },
            { name: "Structures", path: "/structures" },
            { name: "Servers", path: "/servers" },
        ]);
    });

    test("no permissions, non-sudo partition", async ({ page }) => {
        // Load the page.
        await defaultSetup(page, [], false, true);

        // Get the links from the navbar.
        const links = (await page.$$("[data-navbar] a")) as ElementHandle<HTMLElement>[];

        // Validate the links.
        await elementsMatch(links, []);
    });

    test("no permissions, sudo partition", async ({ page }) => {
        // Load the page.
        await defaultSetup(page, [], true, true);

        // Get the links from the navbar.
        const links = (await page.$$("[data-navbar] a")) as ElementHandle<HTMLElement>[];

        // Validate the links.
        await elementsMatch(links, []);
    });
});
