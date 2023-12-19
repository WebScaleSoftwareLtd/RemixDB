// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import React from "react";
import { useSudoPartition } from "@/authState";
import { Pointer, metricsPoller } from "@/pollers";
import Container from "@/atoms/Container";
import Center from "@/atoms/Center";
import ErrorView from "@/atoms/ErrorView";
import Title from "@/atoms/Title";
import Box from "@/atoms/Box";
import Subtitle from "@/atoms/Subtitle";
import Spinner from "@/atoms/Spinner";

const chartImport = import("@tremor/react").then(x => ({ default: x.AreaChart }));
const AreaChart = React.lazy(() => chartImport);

const GenericView = () => <Center>
    <h1 className="text-4xl mb-4">Welcome to RemixDB!</h1>
    <p>
        Use the navigation bar to select a route for the items you have permission to access.
    </p>
</Center>;

const MetricsView = () => {
    // Defines loading the metrics.
    const metrics = React.useSyncExternalStore(
        alert => {
            const subscriptionId = metricsPoller.subscribe(alert);
            return () => metricsPoller.unsubscribe(subscriptionId);
        },
        () => {
            try {
                return metricsPoller.events;
            } catch (e) {
                return e as Error;
            }
        }
    );

    // If there is an error, return it.
    if (!(metrics instanceof Pointer)) {
        return <Center>
            <ErrorView error={metrics} />
        </Center>;
    }

    // Return the metrics view.
    return <>
        <Title>Metrics</Title>
        <p>
            Here are the metrics in your cluster:
        </p>

        <Box large margin>
            <Subtitle>CPU Usage</Subtitle>
            <p>
                This is the CPU usage of the cluster. It is the sum of the CPU
                usage of all the nodes in the cluster.
            </p>

            <React.Suspense fallback={<Spinner />}>
                <AreaChart
                    className="h-72 mt-4"
                    data={metrics.value.map(x => ({
                        at: Intl.DateTimeFormat("en-GB", {
                            hour: "numeric",
                            minute: "numeric",
                            second: "numeric",
                        }).format(x.at),
                        "CPU Usage": x.event.cpu_percent,
                    }))}
                    categories={["CPU Usage"]}
                    valueFormatter={x => `${x.toFixed(2)}%`}
                    colors={["red"]}
                    index="at"
                />
            </React.Suspense>
        </Box>

        <Box large margin>
            <Subtitle>Memory Usage</Subtitle>
            <p>
                This is the memory usage of the cluster. It is the sum of the
                memory usage of all the nodes in the cluster.
            </p>

            <React.Suspense fallback={<Spinner />}>
                <AreaChart
                    className="h-72 mt-4"
                    data={metrics.value.map(x => ({
                        at: Intl.DateTimeFormat("en-GB", {
                            hour: "numeric",
                            minute: "numeric",
                            second: "numeric",
                        }).format(x.at),
                        "Memory Usage": x.event.ram_mb,
                    }))}
                    categories={["Memory Usage"]}
                    valueFormatter={x => `${x} MB`}
                    colors={["emerald"]}
                    index="at"
                />
            </React.Suspense>
        </Box>

        <Box large margin>
            <Subtitle>Number of Goroutines</Subtitle>
            <p>
                This is the number of goroutines in the cluster. Goroutines work
                a bit like threads, but are not quite as heavy since they can be
                paused and also do not need to create all thread resources.
            </p>

            <React.Suspense fallback={<Spinner />}>
                <AreaChart
                    className="h-72 mt-4"
                    data={metrics.value.map(x => ({
                        at: Intl.DateTimeFormat("en-GB", {
                            hour: "numeric",
                            minute: "numeric",
                            second: "numeric",
                        }).format(x.at),
                        "Number of Goroutines": x.event.goroutines,
                    }))}
                    categories={["Number of Goroutines"]}
                    colors={["blue"]}
                    index="at"
                />
            </React.Suspense>
        </Box>

        <Box large margin>
            <Subtitle>Number of Garbage Collections</Subtitle>
            <p>
                Defines the number of garbage collections that have happened in
                the cluster. Garbage collection is the process of freeing up
                memory that is no longer needed.
            </p>

            <React.Suspense fallback={<Spinner />}>
                <AreaChart
                    className="h-72 mt-4"
                    data={metrics.value.map(x => ({
                        at: Intl.DateTimeFormat("en-GB", {
                            hour: "numeric",
                            minute: "numeric",
                            second: "numeric",
                        }).format(x.at),
                        "Number of Garbage Collections": x.event.gcs,
                    }))}
                    categories={["Number of Garbage Collections"]}
                    colors={["orange"]}
                    index="at"
                />
            </React.Suspense>
        </Box>
    </>;
};

export const IndexRoute = () => {
    const sudoPartition = useSudoPartition();
    return <Container>{sudoPartition ? <>
        <GenericView />
        <hr className="my-8 border-gray-200" />
        <MetricsView />
    </> : <GenericView />}</Container>;
};
