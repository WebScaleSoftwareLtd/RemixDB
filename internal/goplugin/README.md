# Go Plugin Compiler

This module contains the Go plugin compiler. The way this works is it installs a copy of Go to the users system and then compiles any code specified to it (assuming it isn't cached) to a Go plugin and then running it. If the version RemixDB is compiled with changes, the cache/locally downloaded Go version is automatically thrown out.

This allows the performance of the database to be fast because the actual database code can be pre-loaded into Go.

Seeing as we are using one of the more controversial libraries within the stdlib (`plugin`), it doesn't get as much love as some other parts of Go and there are some caveats:

- Windows support is non-existent. This is very annoying because it is the sole reason the database will not compile on Windows.
- You cannot import anything from outside of the standard library. This is because path hashing creates problems for this. To circumvent this, in the compiler we pass out a really reflected value that expects a `pluginFriendlyRpc` type from `rpc/requesthandler`. This type works to add things as needed and make types `any` where they cannot be ported over.
