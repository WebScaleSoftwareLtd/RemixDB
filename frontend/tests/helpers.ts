// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { Page, test } from "@playwright/test";

export const screenshotErrors = () => {
    test.afterEach(async ({ page }, testInfo) => {
        // Ignore successful tests.
        if (testInfo.status === testInfo.expectedStatus) return;

        // Get the output path.
        const screenshotPath = testInfo.outputPath("screenshot.png");

        // Push the path to attachments.
        testInfo.attachments.push({
            name: "Screenshot",
            path: screenshotPath,
            contentType: "image/png",
        });

        // Take a screenshot.
        await page.screenshot({ path: screenshotPath });
    });
};

export type TestInitConfig = {
    page: Page;
    path: string;
    permissions: string[];
    sudoPartition: boolean;
    username?: string;
    skipWait?: boolean;
};

export const initTest = async (config: TestInitConfig) => {
    // Make sure that /api/v1/partition/created returns true.
    await config.page.route("/api/v1/partition/created", route => {
        route.fulfill({ json: true });
    });

    // Go to the page.
    await config.page.goto(config.path);

    // Wait for the API key input selector.
    const input = await config.page.waitForSelector(
        "input[placeholder='Enter your API key']"
    );

    // Defines the result of /api/v1/user.
    await config.page.route("/api/v1/user", route =>
        route.fulfill({
            json: {
                username: config.username ? config.username : "test-username",
                permissions: config.permissions,
                sudo_partition: config.sudoPartition,
            },
        })
    );

    // Type in the API key.
    await input.fill("test-api-key");

    // Click the login button.
    await config.page.click("button");

    // Wait for the header to appear.
    if (!config.skipWait) await config.page.waitForSelector("header");
};
