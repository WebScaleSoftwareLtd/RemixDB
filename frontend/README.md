# RemixDB Frontend

This is the frontend for RemixDB. It is a React app that uses the [RemixDB API](../api/README.md) to display data. The frontend code is generated using Vite.

Styling is done with TailwindCSS, allowing for fairly clean and light styling. With our usage of Tailwind, we prefer a set of re-usable components that we can import from many places for consistency.

Routes are stored in the routes.json file. The key/value map is used to defines a route mapping to a component name specified in the `src/routes` exports. Routes are managed by `react-router-dom`. When deployed, the `dist` folder will be served by the web server within the application. This is done with `npm run build:prod`.

In production, the route keys are used to map the index.html file within Go too. Therefore, make sure that the keys support both `react-router-dom` and `julienschmidt/httprouter`.

## Development

To start the development server, run `npm run dev`. This will start a Vite server. From here, we can use the environment variable `REMIXDB_DEV_FRONTEND_HOST` to proxy RemixDB's web server to the Vite server to allow us to develop the frontend without having to restart RemixDB. For example, if the frontend is running on `localhost:3000` and RemixDB is running on `localhost:8080`, we can use `REMIXDB_DEV_FRONTEND_HOST=localhost:3000` to proxy to the frontend from RemixDB. Any route that 404's will be proxied to the frontend.
