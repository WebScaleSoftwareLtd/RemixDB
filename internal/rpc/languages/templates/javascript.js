// This file is automatically generated by RemixDB. Do not edit.

"use strict";
// AUTO-GENERATION MARKER: imports

// Defines all of the auto-generated structures.
const _autoGeneratedStructures = {};

// Defines the magical any object to mean this can be any type.
const _any = {};

// Defines the magical object to mean uint or float.
const _uint = {};
const _float = {};

// Validates a type. typeAnnotation can be the _any object, a array
// (either of 1 type to mean array type or multiple meaning a union),
// a object with _key set to an array of [keyType, valueType], or a
// String/Number/Boolean/BigInt/Date/null/struct class type.
function _validateType(value, typeAnnotation) {
  if (typeAnnotation === _any) return;

  // Handle the possibilities for arrays.
  if (Array.isArray(typeAnnotation)) {
    // Handle when it is just an array case.
    if (typeAnnotation.length === 1) {
      if (!Array.isArray(value)) throw new Error("Expected array");
      for (const v of value) _validateType(v, typeAnnotation[0]);
      return;
    }

    // Handle when it is a union case.
    for (const t of typeAnnotation) {
      try {
        _validateType(value, t);
        return;
      } catch (_) {
        // Ignore.
      }
    }
    throw new Error(`Expected one of ${typeAnnotation} to match ${value}`);
  }

  // Handle null.
  if (typeAnnotation === null) {
    if (value !== null) throw new Error("Expected null");
    return;
  }

  // Handle the magic uint/float case.
  if (typeAnnotation === _uint) {
    if (typeof value !== "number") throw new Error("Expected number");
    if (value < 0) throw new Error("Expected positive number");
    if (value !== Math.floor(value)) throw new Error("Expected integer");
    return;
  }
  if (typeAnnotation === _float) {
    if (typeof value !== "number") throw new Error("Expected number");
    return;
  }

  // Handle the object cases.
  let kv;
  try {
    kv = typeAnnotation._key;
  } catch (_) {
    // Ignore.
  }
  if (kv) {
    if (typeof value !== "object") throw new Error("Expected object");
    for (const k in value) {
      _validateType(k, kv[0]);
      _validateType(value[k], kv[1]);
    }
    return;
  }

  // Handle the class case.
  if (typeof typeAnnotation === "function") {
    if (!(value instanceof typeAnnotation)) {
      throw new Error(`Expected ${typeAnnotation.name}, got ${value}`);
    }
    return;
  }
}

function _readUint32Le(bytes, offset) {
  return (
    (bytes[offset] << 0) |
    (bytes[offset + 1] << 8) |
    (bytes[offset + 2] << 16) |
    (bytes[offset + 3] << 24)
  );
}

function _readUint16Le(bytes, offset) {
  return (bytes[offset] << 0) | (bytes[offset + 1] << 8);
}

function _readInt64Le(bytes, offset) {
  const high = _readUint32Le(bytes, offset + 4);
  const low = _readUint32Le(bytes, offset);
  return (BigInt(high) << BigInt(32)) | BigInt(low);
}

function _readUint64Le(bytes, offset) {
  const high = _readUint32Le(bytes, offset + 4);
  const low = _readUint32Le(bytes, offset);
  return high | low;
}

function _readFloat64Le(bytes, offset) {
  const buffer = new ArrayBuffer(8);
  const view = new DataView(buffer);
  for (let i = 0; i < 8; i++) view.setUint8(i, bytes[offset + i]);
  return view.getFloat64(0, true);
}

function _parseBytes(bytes, root) {
  function readBytes() {
    // If it is the root, return the bytes.
    if (root) return bytes.slice(1);

    // Read the length after the type.
    const length = _readUint32Le(bytes, 1);
    return bytes.slice(5, 5 + length);
  }

  // Get the type.
  const type = bytes[0];
  switch (type) {
    case 0x00:
      return [null, 1];
    case 0x01:
      return [false, 1];
    case 0x02:
      return [true, 1];
    case 0x03:
      return [new Uint8Array(), 1];
    case 0x04:
      return ["", 1];
    case 0x05: {
      const b = readBytes();
      return [b, (root ? 1 : 5) + b.length];
    }
    case 0x06: {
      const b = readBytes();
      return [new TextDecoder().decode(b), (root ? 1 : 5) + b.length];
    }
    case 0x07: {
      const length = _readUint32Le(bytes, 1);
      bytes = bytes.slice(5);
      const result = [];
      let x = 0;
      for (let i = 0; i < length; i++) {
        const [value, length] = _parseBytes(bytes, false);
        result.push(value);
        bytes = bytes.slice(length);
        x += length;
      }
      return [result, x + 5];
    }
    case 0x08: {
      const length = _readUint32Le(bytes, 1);
      bytes = bytes.slice(5);
      const result = {};
      let x = 0;
      for (let i = 0; i < length; i++) {
        const [key, keyLength] = _parseBytes(bytes, false);
        bytes = bytes.slice(keyLength);
        x += keyLength;
        const [value, valueLength] = _parseBytes(bytes, false);
        bytes = bytes.slice(valueLength);
        x += valueLength;
        result[key] = value;
      }
      return [result, x + 5];
    }
    case 0x09: {
      // Get the name length.
      const nameLength = bytes[1];

      // Get the name.
      const name = new TextDecoder().decode(bytes.slice(2, 2 + nameLength));

      // Get the structure.
      const structure = _autoGeneratedStructures[name];
      const x = new structure(bytes);
      if (structure) return [x, 2 + nameLength + x._bytesLength];

      // Skip through the structure.
      let i = 2 + nameLength;
      const itemNameLen = _readUint16Le(bytes, i);
      i += 2 + itemNameLen;
      const itemValueLen = _readUint32Le(bytes, i);
      i += 4 + itemValueLen;

      // Return null.
      return [null, i];
    }
    case 0x0a:
      return [_readInt64Le(bytes, 1), 9];
    case 0x0b:
      return [_readFloat64Le(bytes, 1), 9];
    case 0x0c: {
      const timestamp = _readInt64Le(bytes, 1);
      return [new Date(timestamp), 9];
    }
    case 0x0d: {
      const b = readBytes();
      return [new BigInt(b), (root ? 1 : 5) + b.length];
    }
    case 0x0e: {
      const u = _readUint64Le(bytes, 1);
      return [BigInt(u), 9];
    }

    // Int constants.
    case 0x10:
    case 0x30:
    case 0x60:
      return [0, 1];
    case 0x11:
    case 0x31:
    case 0x61:
      return [1, 1];
    case 0x12:
    case 0x32:
    case 0x62:
      return [2, 1];
    case 0x13:
    case 0x33:
    case 0x63:
      return [3, 1];
    case 0x14:
    case 0x34:
    case 0x64:
      return [4, 1];
    case 0x15:
    case 0x35:
    case 0x65:
      return [5, 1];
    case 0x16:
    case 0x36:
    case 0x66:
      return [6, 1];
    case 0x17:
    case 0x37:
    case 0x67:
      return [7, 1];
    case 0x18:
    case 0x38:
    case 0x68:
      return [8, 1];
    case 0x19:
    case 0x39:
    case 0x69:
      return [9, 1];
    case 0x1a:
    case 0x3a:
    case 0x6a:
      return [10, 1];
    case 0x1b:
    case 0x3b:
    case 0x6b:
      return [11, 1];
    case 0x1c:
    case 0x3c:
    case 0x6c:
      return [12, 1];
    case 0x1d:
    case 0x3d:
    case 0x6d:
      return [13, 1];
    case 0x1e:
    case 0x3e:
    case 0x6e:
      return [14, 1];
    case 0x1f:
    case 0x3f:
    case 0x6f:
      return [15, 1];
    case 0x20:
    case 0x70:
      return [-1, 1];
    case 0x21:
    case 0x71:
      return [-2, 1];
    case 0x22:
    case 0x72:
      return [-3, 1];
    case 0x23:
    case 0x73:
      return [-4, 1];
    case 0x24:
    case 0x74:
      return [-5, 1];
    case 0x25:
    case 0x75:
      return [-6, 1];
    case 0x26:
    case 0x76:
      return [-7, 1];
    case 0x27:
    case 0x77:
      return [-8, 1];
    case 0x28:
    case 0x78:
      return [-9, 1];
    case 0x29:
    case 0x79:
      return [-10, 1];
    case 0x2a:
    case 0x7a:
      return [-11, 1];
    case 0x2b:
    case 0x7b:
      return [-12, 1];
    case 0x2c:
    case 0x7c:
      return [-13, 1];
    case 0x2d:
    case 0x7d:
      return [-14, 1];
    case 0x2e:
    case 0x7e:
      return [-15, 1];
    case 0x2f:
    case 0x7f:
      return [-16, 1];
    case 0x40:
      return [BigInt(0), 1];
    case 0x41:
      return [BigInt(1), 1];
    case 0x42:
      return [BigInt(2), 1];
    case 0x43:
      return [BigInt(3), 1];
    case 0x44:
      return [BigInt(4), 1];
    case 0x45:
      return [BigInt(5), 1];
    case 0x46:
      return [BigInt(6), 1];
    case 0x47:
      return [BigInt(7), 1];
    case 0x48:
      return [BigInt(8), 1];
    case 0x49:
      return [BigInt(9), 1];
    case 0x4a:
      return [BigInt(10), 1];
    case 0x4b:
      return [BigInt(11), 1];
    case 0x4c:
      return [BigInt(12), 1];
    case 0x4d:
      return [BigInt(13), 1];
    case 0x4e:
      return [BigInt(14), 1];
    case 0x4f:
      return [BigInt(15), 1];
    case 0x50:
      return [BigInt(-1), 1];
    case 0x51:
      return [BigInt(-2), 1];
    case 0x52:
      return [BigInt(-3), 1];
    case 0x53:
      return [BigInt(-4), 1];
    case 0x54:
      return [BigInt(-5), 1];
    case 0x55:
      return [BigInt(-6), 1];
    case 0x56:
      return [BigInt(-7), 1];
    case 0x57:
      return [BigInt(-8), 1];
    case 0x58:
      return [BigInt(-9), 1];
    case 0x59:
      return [BigInt(-10), 1];
    case 0x5a:
      return [BigInt(-11), 1];
    case 0x5b:
      return [BigInt(-12), 1];
    case 0x5c:
      return [BigInt(-13), 1];
    case 0x5d:
      return [BigInt(-14), 1];
    case 0x5e:
      return [BigInt(-15), 1];
    case 0x5f:
      return [BigInt(-16), 1];
    default:
      throw new Error(`Unknown type ${type}`);
  }
}

const _valMap = new Map([
  [null, 0x00],
  [false, 0x01],
  [true, 0x02],
  [new Uint8Array(), 0x03],
  ["", 0x04],
]);

const _classMap = new Map([
  [
    Uint8Array,
    (val, root) => {
      const v = new Uint8Array((root ? 1 : 5) + val.length);
      v[0] = 0x05;
      if (!root) {
        v[1] = val.length & 0xff;
        v[2] = (val.length >> 8) & 0xff;
        v[3] = (val.length >> 16) & 0xff;
        v[4] = (val.length >> 24) & 0xff;
      }
      v.set(val, root ? 1 : 5);
      return v;
    },
  ],
  [
    String,
    (val, root) => {
      const b = new TextEncoder().encode(val);
      const v = new Uint8Array((root ? 1 : 5) + b.length);
      v[0] = 0x06;
      if (!root) {
        v[1] = b.length & 0xff;
        v[2] = (b.length >> 8) & 0xff;
        v[3] = (b.length >> 16) & 0xff;
        v[4] = (b.length >> 24) & 0xff;
      }
      v.set(b, root ? 1 : 5);
      return v;
    },
  ],
  [
    Array,
    (val) => {
      const length = val.length;
      const v = new Uint8Array(5);
      v[0] = 0x07;
      v[1] = length & 0xff;
      v[2] = (length >> 8) & 0xff;
      v[3] = (length >> 16) & 0xff;
      v[4] = (length >> 24) & 0xff;
      let offset = 5;
      for (const x of val) {
        const b = _encode(x, _any, false);
        const bLen = b.length;
        const newV = new Uint8Array(offset + bLen);
        newV.set(v);
        newV.set(b, offset);
        v = newV;
        offset += bLen;
      }
      return v;
    },
  ],
  [
    Object,
    (val) => {
      const keys = Object.keys(val);
      const length = keys.length;
      const v = new Uint8Array(5);
      v[0] = 0x08;
      v[1] = length & 0xff;
      v[2] = (length >> 8) & 0xff;
      v[3] = (length >> 16) & 0xff;
      v[4] = (length >> 24) & 0xff;
      let offset = 5;
      for (const key of keys) {
        const b = _encode(key, _any, false);
        const bLen = b.length;
        const newV = new Uint8Array(offset + bLen);
        newV.set(v);
        newV.set(b, offset);
        v = newV;
        offset += bLen;
        const b2 = _encode(val[key], _any, false);
        const bLen2 = b2.length;
        const newV2 = new Uint8Array(offset + bLen2);
        newV2.set(v);
        newV2.set(b2, offset);
        v = newV2;
        offset += bLen2;
      }
      return v;
    },
  ],
  [
    Date,
    (val) => {
      const v = new Uint8Array(9);
      v[0] = 0x0c;
      const timestamp = val.getTime();
      v[1] = timestamp & 0xff;
      v[2] = (timestamp >> 8) & 0xff;
      v[3] = (timestamp >> 16) & 0xff;
      v[4] = (timestamp >> 24) & 0xff;
      v[5] = (timestamp >> 32) & 0xff;
      v[6] = (timestamp >> 40) & 0xff;
      v[7] = (timestamp >> 48) & 0xff;
      v[8] = (timestamp >> 56) & 0xff;
      return v;
    },
  ],
  // TODO
]);

function _encode(value, typeAnnotation, root) {
  let b = _valMap.get(value);
  if (b !== undefined) return new Uint8Array([b]);

  const hn = _classMap.get(value.constructor);
  if (hn !== undefined) return hn(value, root, typeAnnotation);

  if (_autoGeneratedStructures[value.constructor.name]) {
    return value.toBinary();
  }

  throw new Error(`Unknown value ${value}`);
}

class ServerError extends Error {
  constructor(code, message) {
    super(`${code}: ${message}`);
    this.code = code;
    this.message = message;
  }
}

class _BaseModel {
  constructor(values, types, name) {
    this._types = types;
    this._name = name;
    values instanceof Uint8Array
      ? this._fromBinary(values)
      : this._fromObject(values);
  }

  _fromObject(values) {
    const typesCpy = { ...this._types };
    for (const key in values) {
      if (typesCpy[key] === undefined) throw new Error("Unknown key " + key);
      const type = typesCpy[key];
      delete typesCpy[key];
      const v = values[key];
      _validateType(v, type);
      this[key] = v;
    }
    for (const key in typesCpy) {
      let val = typesCpy[key];
      if (
        val === null ||
        (Array.isArray(val) && val.length > 1 && val.includes(null))
      ) {
        this[key] = null;
        continue;
      }
      throw new Error("Missing key " + key);
    }
  }

  _fromBinary(data) {
    if (data[0] !== 0x09)
      throw new Error("Expected struct type, got " + data[0]);
    const nameLength = data[1];
    const name = new TextDecoder("utf-8").decode(data.slice(2, 2 + nameLength));
    if (name !== this._name)
      throw new Error("Expected struct name " + this._name + ", got " + name);
    data = data.slice(2 + nameLength);
    const length = data[0] + data[1] * 256;
    data = data.slice(2);
    const obj = {};
    for (let i = 0; i < length; i++) {
      const keyLength = data[0] + data[1] * 256;
      const key = new TextDecoder("utf-8").decode(data.slice(2, 2 + keyLength));
      data = data.slice(2 + keyLength);
      const valueLength =
        data[0] +
        data[1] * 256 +
        data[2] * 256 * 256 +
        data[3] * 256 * 256 * 256;
      obj[key] = _parseBytes(data.slice(4, 4 + valueLength), true);
      data = data.slice(4 + valueLength);
    }
    this._fromObject(obj);
  }

  toBinary() {
    const parts = [];
    const len = 2;
    parts.push(new Uint8Array([0x09, n.length]));
    parts.push(new TextEncoder().encode(n));
    len += n.length;
    const keys = Object.keys(this._types);
    for (const key of keys) {
      const type = this._types[key];
      const value = this[key];
      if (value === null) continue;
      const keyBytes = new TextEncoder().encode(key);
      const valueBytes = _encode(value, type, false);
      const keyLen = keyBytes.length;
      const valueLen = valueBytes.length;
      len += 2 + keyLen + 4 + valueLen;
    }
    const a = new Uint8Array(len);
    let offset = 0;
    for (const part of parts) {
      a.set(part, offset);
      offset += part.length;
    }
  }
}

// AUTO-GENERATION MARKER: structures

function _parseExceptionPacket(data) {
  const h = data[0];
  if (h === 0x00 || h === 0x01) {
    // Read the next 2 bytes.
    const length = _readUint16Le(data, 1);
    const totalLen = 3 + length;
    const name = data.slice(3, totalLen);

    // The remaining bytes are the value.
    const value = data.slice(totalLen);

    // If h is 0x00, it is a RemixDB error.
    if (h === 0x00) {
      throw new ServerError(
        new TextDecoder().decode(name),
        new TextDecoder().decode(value)
      );
    }

    // Decode value as JSON.
    const json = new TextDecoder().decode(value);
    const obj = JSON.parse(json);

    // Make sure it is a object.
    if (typeof obj !== "object") {
      throw new Error(`Expected object, got ${obj}`);
    }

    // Check if name is a struct name.
    const struct = _autoGeneratedStructures[name];
    if (struct) throw new struct(obj);
    throw new Error(`Unknown struct ${name}`);
  }
}

class Cursor {
  constructor(ws, type) {
    this._ws = ws;
    this._type = type;

    // Buffer messages and errors.
    this._nextId = 0;
    this._messages = [];
    this._resolvers = new Map();
    this._error = null;
    this._rejectors = new Map();

    // Add the event listeners.
    this._ws.addEventListener("message", (event) => {
      const data = event.data;
      const resolvers = this._resolvers;
      this._resolvers = new Map();
      for (const res of resolvers.values()) res(data);
      if (resolvers.size() === 0) this._messages.push(data);
    });
    this._ws.addEventListener("error", (event) => {
      this._error = event;
      const rejectors = this._rejectors;
      this._rejectors = new Map();
      for (const rej of rejectors.values()) rej(event);
    });
  }

  _waitForMessage() {
    const resId = this._nextId++;
    const rejId = this._nextId++;
    return new Promise((resolve, reject) => {
      if (this._error) return reject(this._error);
      if (this._messages.length > 0) return resolve(this._messages.shift());

      this._resolvers.set(resId, (data) => {
        this._resolvers.delete(resId);
        this._rejectors.delete(rejId);
        resolve(data);
      });

      this._rejectors.set(rejId, (error) => {
        this._resolvers.delete(resId);
        this._rejectors.delete(rejId);
        reject(error);
      });
    });
  }

  async next() {
    // Send 0x01 to get the next value.
    this._ws.send(new Uint8Array([0x01]));

    // Wait for the response.
    const data = await this._waitForMessage();

    // Get the header and make sure it is not an exception.
    const h = data[0];
    _parseExceptionPacket(data);

    // If the data is 0x02, chop off the first byte and parse it.
    if (h === 0x02) {
      const [value, _] = _parseBytes(data.slice(1), true);
      _validateType(value, this._type);
      return { value, done: false };
    }

    // If the data is 0x03, then we are done.
    if (h === 0x03) return { done: true };
  }

  close() {
    this._ws.close();
  }
}

class Client {
  constructor(url, options) {
    if (typeof options !== "object") {
      throw new Error("Expected options to be an object");
    }
    this._url = new URL(url);
    this._options = new TextEncoder().encode(JSON.stringify(options) + "\n");
  }

  async _doNonCursorRequest(method, data, schemaHash, type) {
    // Make the request.
    const urlCopy = new URL(this._url);
    urlCopy.pathname = `/rpc/${encodeURIComponent(method)}`;
    const body = new Uint8Array(this._options.length + data.length);
    body.set(this._options);
    body.set(data, this._options.length);
    const res = await fetch(urlCopy.toString(), {
      method: "POST",
      headers: {
        "X-RemixDB-Schema-Hash": schemaHash,
        "Content-Type": "application/x-remixdb-rpc-mixed",
      },
      body,
    });

    // Make sure X-Is-RemixDB is set.
    const isRemixDB = res.headers.get("X-Is-RemixDB");
    if (isRemixDB !== "true") {
      throw new ServerError(
        "response_is_not_remixdb",
        "The response does not appear to be from RemixDB. Does your reverse proxy let through the X-Is-RemixDB header?"
      );
    }

    // Check if the response is a 204.
    if (res.status === 204) {
      _validateType(null, type);
      return null;
    }

    // Check if the response is a 200.
    if (res.status === 200) {
      // Read the bytes.
      const bytes = await res.arrayBuffer();
      const bytesArray = new Uint8Array(bytes);

      // Parse the bytes.
      const [value] = _parseBytes(bytesArray, true);
      _validateType(value, type);
      return value;
    }

    // Check if X-RemixDB-Exception is set.
    const customException = res.headers.get("X-RemixDB-Exception");
    if (customException) {
      // Check if it is in our list of exceptions.
      const exception = _autoGeneratedStructures[customException];

      // Parse the body.
      const json = await res.json();
      if (typeof json !== "object") {
        throw new Error(`Expected object for exception, got ${json}`);
      }

      // Throw the exception.
      if (exception) throw new exception(json);

      // Throw a generic exception.
      throw new Error(`Unknown exception ${customException}`);
    }

    // Parse the body.
    const json = await res.json();
    if (typeof json !== "object") {
      throw new Error(`Expected object, got ${json}`);
    }

    // Check if code and message is set and is a string.
    const code = json.code;
    const message = json.message;
    if (typeof code === "string" && typeof message === "string") {
      throw new ServerError(code, message);
    }
    throw new Error(`Unknown error ${json}`);
  }

  async _doCursorRequest(method, data, schemaHash, type) {
    // Get the URL.
    const urlCopy = new URL(this._url);
    urlCopy.pathname = "/rpc";

    // Make the request.
    const ws = new WebSocket(urlCopy.toString());
    ws.binaryType = "arraybuffer";

    // Wrap it in a cursor.
    const cursor = new Cursor(ws, type);

    // Send the initialization message.
    let msg = new Uint8Array(
      2 + method.length + 2 + schemaHash.length + data.length
    );

    // First 2 bytes are the method length in little endian.
    msg[0] = method.length & 0xff;
    msg[1] = (method.length >> 8) & 0xff;

    // Now we write the method.
    msg.set(method, 2);

    // Now we write the schema hash length.
    msg[2 + method.length] = schemaHash.length & 0xff;
    msg[3 + method.length] = (schemaHash.length >> 8) & 0xff;

    // Now we write the schema hash.
    msg.set(schemaHash, 4 + method.length);

    // Now we write the data.
    msg.set(data, 4 + method.length + schemaHash.length);

    // Send the message.
    ws.send(msg);

    // Wait for the first message.
    msg = await cursor._waitForMessage();

    // Check if it is 0x02.
    if (msg[0] === 0x02) return cursor;

    // Handle exceptions.
    _parseExceptionPacket(msg);
    throw new Error(`Expected 0x02, got ${msg[0]}`);
  }

  // AUTO-GENERATION MARKER: methods
}

/* CJS MODIFICATION NEEDED */ export {
  ServerError,
  Cursor,
  Client, // AUTO-GENERATION MARKER: exports
};