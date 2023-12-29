// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { test, expect, Page } from "@playwright/test";
import { initTest, screenshotErrors } from "./helpers";

test.describe("authentication wrapper", () => {
    screenshotErrors();

    test.describe("partition generation", () => {
        // Run tests within this group in serial.
        test.describe.configure({ mode: "serial" });

        // Create the shared page.
        let page: Page;
        test.beforeAll(async ({ browser }) => {
            page = await browser.newPage();
        });
        test.afterAll(async () => {
            await page.close();
        });

        // Make sure the form and page is as expected.
        test("loads partition creation page", async () => {
            // Make sure that /api/v1/partition/created returns false.
            await page.route("/api/v1/partition/created", route => {
                route.fulfill({ json: false });
            });

            // Go to the root page.
            await page.goto("/");

            // Wait a few milliseconds.
            await page.waitForSelector("h2");

            // Expect there to be 1 h2 element.
            expect(await page.$$eval("h2", els => els.length)).toBe(1);

            // Expect the one h2 element to have the text "Setup Partition".
            expect(await page.$eval("h2", el => el.textContent)).toBe("Setup Partition");

            // Find the input with the placeholder "Enter your sudo API key".
            let input = await page.$("input[placeholder='Enter your sudo API key']");
            expect(input).toBeTruthy();

            // Type in the API key.
            await input?.fill("test-api-key");

            // Find the input with the placeholder "Enter your username".
            input = await page.$("input[placeholder='Enter your username']");
            expect(input).toBeTruthy();

            // Type in the username.
            await input?.fill("test-username");

            // Make sure there is only one button.
            expect(await page.$$eval("button", els => els.length)).toBe(1);
        });

        // TODO: a11y test here

        // Handle API errors.
        test("handles api errors", async () => {
            // Make /api/v1/partition/create return a 400.
            await page.route("/api/v1/partition/create", route => {
                route.fulfill({
                    json: {
                        code: "this_is_a_test_error_code",
                        message: "test error message",
                    },
                    status: 400,
                });
            });

            // Click the button a couple times.
            await page.click("button");
            await page.click("button");

            // Wait for the selector.
            const div = await page.waitForSelector("div.bg-red-500");

            // Expect there to be a single div with the class "bg-red-500".
            const divs = await page.$$("div.bg-red-500");
            expect(divs.length).toBe(1);

            // Expect the div to have the text "test error message".
            expect(await div.textContent()).toBe("test error message");
        });

        // Handle making sure the sudo partition checkbox works.
        test("validates sudo partition checkbox", async () => {
            // Defines the request handler that validates this.
            const validateSudoState = (sudo: boolean) =>
                page.route("/api/v1/partition/create", route => {
                    // Get the request body.
                    const body = route.request().postDataJSON();

                    // Expect the sudo partition to be the value of sudo.
                    expect(body.sudo_partition).toBe(sudo);

                    // Return a error.
                    route.fulfill({
                        json: {
                            code: "this_is_a_test_error_code",
                            message: "test error message",
                        },
                        status: 400,
                    });
                });

            // Make sure it is by default false.
            validateSudoState(false);
            await page.click("button");
            await page.waitForTimeout(20);

            // Check the checkbox.
            await page.check("input[type='checkbox']");

            // Make sure it is now true.
            validateSudoState(true);
            await page.click("button");
        });

        // Handle successful partition creation.
        test("handles successful partition creation", async () => {
            // Setup the page creation route.
            await page.route("/api/v1/partition/create", route => {
                // Handle making sure both api_key and username are correct.
                const body = route.request().postDataJSON();
                expect(body.sudo_api_key).toBe("test-api-key");
                expect(body.username).toBe("test-username");

                // Return the key.
                route.fulfill({ json: "my_awesome_api_key" });
            });

            // Click the button.
            await page.click("button");

            // Wait for the selector.
            await page.waitForSelector("code");

            // Expect it to go to a new page listing
            expect(await page.$$eval("h2", els => els.length)).toBe(1);
            expect(await page.$eval("h2", el => el.textContent)).toBe(
                "Partition Created"
            );

            // TODO: a11y test here

            // Look for a code block with the API key.
            const code = await page.$("code");
            expect(code).toBeTruthy();

            // Expect the code to have the text "my_awesome_api_key".
            expect(await code?.textContent()).toBe("my_awesome_api_key");

            // Expect there to be a single button.
            expect(await page.$$eval("button", els => els.length)).toBe(1);

            // Click the button.
            await page.click("button");

            // Wait for the selector for the data-login-form attribute on the form.
            await page.waitForSelector("form[data-login-form]");

            // Expect it to go to the login screen.
            expect(await page.$$eval("h2", els => els.length)).toBe(1);
            expect(await page.$eval("h2", el => el.textContent)).toBe("RemixDB Login");
        });

        // Handle the parition_already_exists error.
        test("handles partition_already_exists error", async ({ browser }) => {
            // Create a new page.
            const page = await browser.newPage();

            // Make sure that /api/v1/partition/created returns false.
            await page.route("/api/v1/partition/created", route => {
                route.fulfill({ json: false });
            });

            // Go to the root page.
            await page.goto("/");

            // Wait for the selector.
            await page.waitForSelector("h2");

            // Make /api/v1/partition/create return a 400.
            await page.route("/api/v1/partition/create", route => {
                route.fulfill({
                    json: {
                        code: "partition_already_exists",
                        message: "test error message",
                    },
                    status: 400,
                });
            });

            // Click the button.
            await page.click("button");

            // Wait for the selector for the login form.
            await page.waitForSelector("form[data-login-form]");

            // Expect it to go to the login screen.
            expect(await page.$$eval("h2", els => els.length)).toBe(1);
            expect(await page.$eval("h2", el => el.textContent)).toBe("RemixDB Login");
        });
    });

    test.describe("login", () => {
        // Run tests within this group in serial.
        test.describe.configure({ mode: "serial" });

        // Create the shared page.
        let page: Page;
        test.beforeAll(async ({ browser }) => {
            page = await browser.newPage();
        });
        test.afterAll(async () => {
            await page.close();
        });

        // Make sure the form and page is as expected.
        test("loads login page", async () => {
            // Make sure that /api/v1/partition/created returns true.
            await page.route("/api/v1/partition/created", route => {
                route.fulfill({ json: true });
            });

            // Go to the root page.
            await page.goto("/");

            // Wait for a h2.
            await page.waitForSelector("h2");

            // Expect there to be 1 h2 element.
            expect(await page.$$eval("h2", els => els.length)).toBe(1);

            // Expect the one h2 element to have the text "Setup Partition".
            expect(await page.$eval("h2", el => el.textContent)).toBe("RemixDB Login");

            // Look for the input with the placeholder "Enter your API key".
            const input = await page.$("input[placeholder='Enter your API key']");
            expect(input).toBeTruthy();

            // Input the API key.
            await input?.fill("test-api-key");

            // Make sure there is only one button.
            expect(await page.$$eval("button", els => els.length)).toBe(1);
        });

        // Handle API errors.
        test("handles api errors", async () => {
            // Make /api/v1/user return a 400.
            await page.route("/api/v1/user", route => {
                route.fulfill({
                    json: {
                        code: "this_is_a_test_error_code",
                        message: "test error message",
                    },
                    status: 400,
                });
            });

            // Click the button a couple times.
            await page.click("button");
            await page.click("button");

            // Wait for a div.
            const div = await page.waitForSelector("div.bg-red-500");

            // Expect there to be a single div with the class "bg-red-500".
            const divs = await page.$$("div.bg-red-500");
            expect(divs.length).toBe(1);

            // Expect the div to have the text "test error message".
            expect(await div.textContent()).toBe("test error message");
        });

        // Handle successful login.
        test("handles successful login", async () => {
            // Setup the page creation route.
            await page.route("/api/v1/user", route => {
                // Handle making sure the API key is correct.
                const headers = route.request().headers();
                expect(headers.authorization).toBe("Bearer test-api-key");

                // Return the key.
                route.fulfill({
                    json: {
                        username: "test-username",
                        sudo_partition: false,
                        permissions: [],
                    },
                });
            });

            // Click the button.
            await page.click("button");

            // Wait a couple hundred milliseconds.
            await page.waitForTimeout(200);

            // Check if test-username is in the page.
            expect(await page.textContent("body")).toContain("test-username");
        });
    });

    test("update logged in state when authentication error", async ({ page }) => {
        // Go ahead and define the error that metrics will get.
        await page.route("/api/v1/metrics", route => {
            route.fulfill({
                json: {
                    code: "test_error_code",
                    message: "test error message",
                },
                status: 401,
            });
        });

        // Setup the page.
        await initTest({
            page,
            path: "/",
            permissions: ["*"],
            sudoPartition: true,
            skipWait: true,
        });

        // Wait for the login form.
        await page.waitForSelector("form[data-login-form]");

        // Expect us to be back on the login page.
        expect(await page.$$eval("h2", els => els.length)).toBe(1);
        expect(await page.$eval("h2", el => el.textContent)).toBe("RemixDB Login");
    });
});
