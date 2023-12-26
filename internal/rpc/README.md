# RemixDB RPC

The goal of this package is to implement client generation and the server to implement logic on top of the RemixDB RPC protocol. The server-side supports both net/http and fasthttp since they both have different benefits for different configurations within the context of this application.

The RPC server within this package is relatively low level. The `requesthandler` package inside this one contains all of the glue required to authenticate/route users. You can attach that handler to the low-level RPC server.

The reasons for implementing a custom RPC protocol are as follows:

- **We get more flexibility to customize the clients:** The clients that some RPC libraries generate are known to be fairly machine-like for a human to use. With this approach, we can generate clients that feel more like a user generated client library. For example, in the Go client libraries, we can generate cursors that use generics rather than generating many cursor types. Much cleaner!
- **We don't need to rely on any third party libraries for data types:** Since we know everything we'll need right off the bat, for data types, we do not need third party libraries.
- **We can make it only contain what we need:** This is useful because it means we can get the data type down to 1 byte.

Test case generation/verification within this package uses [jimeh's go-golden library](https://github.com/jimeh/go-golden). To update the test cases, you can use `make golden-update` in the root of the project. Please review any differences and ensure they make sense before you make a pull request!

Here is how RPC should look/behave from a client and server perspective:

## RemixDB RPC Byte Protocol

If you want to represent void, you should send no bytes here. For everything else, here is all of the supported values. Everything is listed with the first byte first. Any subsequent parts will be listed next to it:

- `0x00`: Null
- `0x01`: False (boolean)
- `0x02`: True (boolean)
- `0x03`: Empty bytes
- `0x04`: Empty string
- `0x05`: Bytes: If this is a root/struct value, the remainder of the message are the bytes. If not, the next 4 bytes should be a uint32 little endian containing the length, then the value should be the length of that integer.
- `0x06`: String: If this is a root/struct value, the remainder of the message is the string. If not, the next 4 bytes should be a uint32 little endian containing the length, then the value should be the length of that integer.
- `0x07`: Array: The next 4 bytes should be a uint32 little endian value containing the length. For each item in this array, you should refer back to this list to figure out what it is.
- `0x08`: Map: The next 4 bytes should be a uint32 little endian value containing the length. For each item in this map, the key will be first, then the value will be after. You should refer to this list for both for information on how each should be parsed.
- `0x09`: Struct: The data following this should be in the following order:
  - 1 byte: Length of struct name
  - N bytes: Struct name (length specified above)
  - 2 bytes (uint16 little endian): Length of struct items
  - For each key in the struct:
    - 2 bytes (uint16 little endian): Length of struct item key
    - N bytes: Struct item key (length specified above)
    - 4 bytes (uint32 little endian): Length of struct item value
    - N bytes (see above): Consult the list to figure out how to parse the value
- `0x0a`: Int: 64-bit little endian integer (8 bytes after this)
- `0x0b`: Float: 64-bit little endian float64 value (8 bytes after this)
- `0x0c`: Timestamp: 64-bit little endian unsigned integer (8 bits after this) representing a unix timestamp in milliseconds
- `0x0d`: Bigint: Sent the same as a string. See above for information on the layout.
- `0x0e`: Uint: 64-bit little endian unsigned integer (8 bytes after this)
- `0x10` - `0x1f`: Used to define 0 to 16 as a integer. This allows us to avoid sending extra bytes for a lot of cases. To get the integer value, simply subtract `0x10`.
- `0x20` - `0x2f`: Used to define -1 to -17 as a integer. This allows us to avoid sending extra bytes for a lot of negative cases. To get the integer value, subtract `0x20` from this byte and then subtract that from -1.
- `0x30`-`0x3f`: Used to define 0 to 16 as a unsigned integer. This allows us to avoid sending extra bytes for a lot of cases. To get the integer value, simply subtract `0x30`.
- `0x40`-`0x4f`: Used to define 0 to 16 as a bigint. This allows us to avoid sending extra bytes for a lot of cases. To get the bigint value, simply subtract `0x40`.
- `0x50` - `0x5f`: Used to define -1 to -17 as a bigint. This allows us to avoid sending extra bytes for a lot of negative cases. To get the bigint value, subtract `0x50` from this byte and then subtract that from -1.
- `0x60`-`0x6f`: Used to define 0 to 16 as a float. This allows us to avoid sending extra bytes for a lot of cases. To get the float value, simply subtract `0x60`.
- `0x70` - `0x7f`: Used to define -1 to -17 as a float. This allows us to avoid sending extra bytes for a lot of negative cases. To get the float value, subtract `0x70` from this byte and then subtract that from -1.

The length of the total packet is assumed to be known with this protocol.

## Schema Method Hash

The schema hash should be a base64 URL encoded version of the below:

- Argument: See the [type hash](#type-hash) documentation below for how to decode.
- Expected Output: See the [type hash](#type-hash) documentation below for how to decode.

### Type Hash

All type hashes start with a int32 in little endian form (4 bytes). If it is negative, it is a built-in type which can be one of the following:

- `-1`: Void
- `-2`: Nullable Boolean
- `-3`: Boolean
- `-4`: Nullable Bytes
- `-5`: Bytes
- `-6`: Nullable String
- `-7`: String
- `-8`: Nullable Array (in this case, the next 4 bytes will be another type hash for the underlying type)
- `-9`: Array (in this case, the next 4 bytes will be another type hash for the underlying type)
- `-10`: Nullable Map (in this case, 2 type hashes will follow for the key and then value)
- `-11`: Nullable Integer
- `-12`: Integer
- `-13`: Nullable Float
- `-14`: Float
- `-15`: Nullable Timestamp
- `-16`: Timestamp
- `-17`: Nullable Bigint
- `-18`: Bigint
- `-19`: Nullable Uint
- `-20`: Uint

If it is positive, it is a revision of a structure. Follow the [struct representation](#struct-representation) documentation below with the bytes after this.

#### Struct Representation

The next 2 bytes are a uint16 little endian repersentation of the struct name length. From there, for the length specified, the sturct name will be present.

TODO

## Non-Cursor HTTP Request

For non-cursors, we use a simple POST request to handle the RPC request. To do this, the client should make a POST request to `/rpc/:method` (where `method` is the RPC method) with the header `X-RemixDB-Schema-Hash` set to the [schema method hash](#schema-method-hash) and `Content-Type` set to `application/x-remixdb-rpc-mixed`. The HTTP body should be [as documented below](#http-request-body).

For the response, the header `X-Is-RemixDB` should always be set to `true`. If this is not the case, you can assume anything came from a application in the middle and throw a error since this is unexpected.

If the response status is either 200 or 204, the request was successful and the response body should be [RemixDB RPC bytes](#remixdb-rpc-byte-protocol) in the shape of the expected output.

If this is not the case, this is an error. The client should check the `X-RemixDB-Exception` header. If this is blank, the body will be a RemixDB server error encoded in JSON. These are in the following format:

- `code` (string): The error code.
- `message` (string): The error message.

If not, this is a custom exception and the header value will be the name of the struct representing this exception within your schema. The response body will be a JSON encoded version of the contents of this struct.

### HTTP Request Body

For the request body, it should be split at the first new line. Everything before the first new line should be a JSON object containing user values for all of the authentication keys that were specified in the RPC structure. This is used for authentication.

Everything after that new line should be [RemixDB RPC bytes](#remixdb-rpc-byte-protocol) in the shape of the expected input.

## Cursor WebSocket Request

Since cursors need to be accessed from many languages that do not have good support for long running HTTP connections, a fairly simple byte protocol is used in place of HTTP for methods that require cursors.

To start the connection, the client will make a WebSocket connection to `/rpc`. To prevent any language limitations, no custom headers are used in this request. From here, when the connection is established, the client should immediately send the following data in the order specified as a binary message:

- 2 bytes (uint16 little endian): Length of the RPC method
- N bytes (specified by the length above): RPC method
- 2 bytes (uint16 little endian): Length of the schema method hash
- N bytes (specified by the length above): [Schema method hash](#schema-method-hash)
- Remainder of the message: Should be treated the same as a [HTTP request body](#http-request-body)

After this is sent, the client should wait for the message back. If the first byte is `0x00` or `0x01`, the message should be treated as a [cursor exception](#cursor-exception). If the first byte is `0x02`, the cursor is ready to be iterated.

To iterate a cursor, the client should send a binary message with the single byte of `0x01`. It should then wait for the server response. A response with the first byte of `0x00` or `0x01` should be treated as a [cursor exception](#cursor-exception). If the response has the first byte of `0x02`, this byte should be sliced off the start of the message and then it should be treated as [RemixDB RPC bytes](#remixdb-rpc-byte-protocol) in the shape of a output of the type that the cursor is emitting. If the response has the first byte of `0x03`, it has hit the end of the items in the cursor. The connection will disconnect after this is sent.

The client should just close the connection when it is done.

### Cursor Exception

How the exception should be parsed depends on the first byte of the message. The first byte should be stored then sliced off. If this byte was `0x00`, it was a RemixDB server error. If it was `0x01`, it was a custom exception.

From here, the layout is the following:

- 2 bytes (uint16 little endian): Length of the error code/exception name (more information below)
- N bytes (specified by the length above): The error code if it was a RemixDB custom exception or the struct name that is the exception if it was not
- Remainder of the message: The JSON body of the struct if it was a custom exception or the error message if it was a server error
