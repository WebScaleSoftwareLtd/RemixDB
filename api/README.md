# RemixDB API

The RemixDB API is the API used for handling RemixDB within the application. The API is used to handle everything relating to operations within the database. The `api` package exposes a low level API wrapper with a interface to implement all the methods. Each API method will be of the signature `func(ctx RequestCtx) (<response type>, error)`. The response type should be marshalable into JSON. The main implementation lives in the `implementation` sub-package.

Note that authentication checking should be done inside the methods since the role of this server is to be low level and not to handle this.

`routes.go` is a file which contains the routes to map the HTTP routes to the API methods. To specify a URL parameter, you should use `{this_syntax}`. It will automatically be converted to both httprouter and fasthttp router compatible routes.
