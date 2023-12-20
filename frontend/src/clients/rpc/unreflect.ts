// RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
// Author: Astrid Gealer <astrid@gealer.email>

import { unreflect } from ".";
import type { RPCReflection } from "./shared";

// Un-reflects a nested RPC value and appends it to the array.
const unreflectNested = (value: RPCReflection, arr: Uint8Array) => {
    // Switch on the type.
    switch (typeof value) {
        case "string":
            // A empty string is 0x04.
            if (value === "") {
                arr[arr.length] = 0x04;
                return;
            }

            // Create a new array with the header and the string.
            const str = new TextEncoder().encode(value);
            const a = new Uint8Array(str.length + 5);
            a[0] = 0x06;
            a.set(str, 1);
            return a;
        case "object":
            // Null should just be 0x00.
            if (value === null) {
                arr[arr.length] = 0x00;
                return;
            }

            // Handle if this is a byte array.
            if (value instanceof Uint8Array) {
                // Empty bytes is 0x03.
                if (!value.length) {
                    arr[arr.length] = 0x03;
                    return;
                }

                // Create a new array with the header and the string.
                const a = new Uint8Array(value.length + 5);
                a[0] = 0x05;
                a.set(value, 1);
                return a;
            }
    }
    const a = unreflect(value);
    arr.set(a, arr.length);
};

// Un-reflects a struct into bytes.
const unreflectStruct = ({
    structName,
    fields,
}: {
    structName: string;
    fields: Map<string, RPCReflection>;
}): Uint8Array => {
    // Make the new array with the header. 0x09 means struct.
    const a = new Uint8Array([0x09, structName.length]);

    // Add the struct name.
    a.set(new TextEncoder().encode(structName), 2);

    // Write the length as uint32 little endian.
    const view = new DataView(a.buffer);
    let i = 2 + structName.length;
    view.setUint32(i, fields.size, true);
    i += 4;

    // Write each element.
    for (const [key, val] of fields) {
        // Add the key.
        view.setUint16(i, key.length, true);
        i += 2;
        a.set(new TextEncoder().encode(key), i);

        // Add the value.
        i += key.length;
        unreflectNested(val, a);
    }

    // Return the array.
    return a;
};

// Does compression of a RPC int value.
const compressRpcValue = (
    value: bigint | number,
    negStart: number,
    negEnd: number,
    posStart: number,
    posEnd: number
) => {
    // Normalize the value.
    value = BigInt(value);

    // Get the positive range.
    const posRange = BigInt(posEnd - posStart);

    // Check if the value is in the positive range.
    if (value >= 0 && value <= posRange) {
        // Return the value plus the start as a number.
        return Number(value + BigInt(posStart));
    }

    // Get the negative range.
    const negRange = BigInt(negEnd - negStart);

    // Check if the value is in the negative range.
    if (value < 0 && value >= -negRange) {
        // Return the value plus the start as a number.
        return Number(value + BigInt(negStart));
    }

    // Return null if it is out of range.
    return null;
};

// Un-reflects a RPC value back to bytes.
export default (value: RPCReflection): Uint8Array => {
    // Switch on the type.
    switch (typeof value) {
        case "boolean":
            // Booleans are 0x01 for false, 0x02 for true.
            return new Uint8Array([value ? 0x02 : 0x01]);
        case "string":
            // A empty string is 0x04.
            if (value === "") return new Uint8Array([0x04]);

            // Create a new array with the header and the string.
            const str = new TextEncoder().encode(value);
            const arr = new Uint8Array(str.length + 1);
            arr[0] = 0x06;
            arr.set(str, 1);
            return arr;
        case "object":
            // Null should just be 0x00.
            if (value === null) return new Uint8Array([0x00]);

            // Handle if this is a byte array.
            if (value instanceof Uint8Array) {
                // Empty bytes is 0x03.
                if (!value.length) return new Uint8Array([0x03]);

                // Create a new array with the header and the string.
                const arr = new Uint8Array(value.length + 1);
                arr[0] = 0x05;
                arr.set(value, 1);
                return arr;
            }

            // Handle if this is a array.
            if (Array.isArray(value)) {
                // Make a new array with the header.
                const arr = new Uint8Array(5);
                arr[0] = 0x07;

                // Write the length as uint32 little endian.
                const view = new DataView(arr.buffer);
                view.setUint32(1, value.length, true);

                // Write each element.
                for (let i = 0; i < value.length; i++) {
                    // Use the nested unreflect.
                    unreflectNested(value[i], arr);
                }

                // Return the array.
                return arr;
            }

            // Handle structs.
            if ("structName" in value) return unreflectStruct(value);

            // Handle if this is a map.
            if (value instanceof Map) {
                // Make a new array with the header.
                const arr = new Uint8Array(5);
                arr[0] = 0x08;

                // Write the length as uint32 little endian.
                const view = new DataView(arr.buffer);
                view.setUint32(1, value.size, true);

                // Write each element.
                for (const [key, val] of value) {
                    // Use the nested unreflect for both key and value.
                    unreflectNested(key, arr);
                    unreflectNested(val, arr);
                }

                // Return the array.
                return arr;
            }

            // Handle if this is a date.
            if ("type" in value) {
                switch (value.type) {
                    case "timestamp":
                        // Make a new array with the header.
                        const a = new Uint8Array(9);
                        a[0] = 0x0c;

                        // Write the length as uint32 little endian.
                        const view = new DataView(a.buffer);
                        view.setBigUint64(1, BigInt(value.unixSeconds), true);

                        // Return the array.
                        return a;
                    case "bigint":
                        // Try to compress the bigint.
                        const big =
                            typeof value.value === "bigint"
                                ? value.value
                                : BigInt(value.value);
                        const v = compressRpcValue(big, 0x5c, 0x5f, 0x40, 0x4f);
                        if (v !== null) return new Uint8Array([v]);

                        // Create a new array with the header and the string.
                        const str = new TextEncoder().encode(big.toString(10));
                        const arr = new Uint8Array(str.length + 1);
                        arr[0] = 0x0d;
                        arr.set(str, 1);
                        return arr;
                    case "int":
                        // Try to compress the int.
                        const int =
                            typeof value.value === "bigint"
                                ? value.value
                                : BigInt(value.value);
                        const v2 = compressRpcValue(int, 0x10, 0x1f, 0x20, 0x2f);
                        if (v2 !== null) return new Uint8Array([v2]);

                        // Create a new array with the header and the string.
                        const arr2 = new Uint8Array(9);
                        arr2[0] = 0x0a;
                        const view2 = new DataView(arr2.buffer);
                        view2.setBigUint64(1, int, true);
                        return arr2;
                    case "float":
                        // Try to compress the float.
                        const v3 = compressRpcValue(value.value, 0x60, 0x6f, 0x70, 0x7f);
                        if (v3 !== null) return new Uint8Array([v3]);
                        const arr3 = new Uint8Array(9);
                        arr3[0] = 0x0b;
                        const view3 = new DataView(arr3.buffer);
                        view3.setFloat64(1, value.value, true);
                        return arr3;
                    case "uint":
                        // Try to compress the uint.
                        const v4 = compressRpcValue(
                            BigInt(value.value),
                            0x30,
                            0x3f,
                            0x30,
                            0x30
                        );
                        if (v4 !== null) return new Uint8Array([v4]);

                        // Create a new array with the header and the string.
                        const arr4 = new Uint8Array(9);
                        arr4[0] = 0x0e;
                        const view4 = new DataView(arr4.buffer);
                        view4.setBigUint64(1, BigInt(value.value), true);
                        return arr4;
                }
            }
    }
    throw new Error(`Unknown RPC reflection type: ${typeof value}`);
};
