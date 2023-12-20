// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

// The types that RPC can relfect to within the protocol.
export type RPCReflection =
    | null
    | boolean
    | Uint8Array
    | string
    | RPCReflection[]
    | Map<RPCReflection, RPCReflection>
    | { structName: string; fields: Map<string, RPCReflection> }
    | { type: "int" | "uint" | "bigint"; value: string | number | bigint }
    | { type: "float"; value: number }
    | { type: "timestamp"; unixSeconds: number };
