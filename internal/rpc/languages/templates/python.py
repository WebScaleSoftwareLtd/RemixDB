# This file is automatically generated by RemixDB. Do not edit.

import base64
import hashlib
import os
import typing
import io
import datetime
import urllib.parse
import urllib.request
import json


# Used internally so we know which classes are auto-generated.
_autogenned_models: typing.Dict[str, typing.Type["_DataModel"]] = {}


def _validate_type(value: typing.Any, type_annotation: typing.Any) -> None:
	"""
	Validates the type of the value against the type annotation. This is included within
	the auto-generated code.
	"""
	# Check if it is something from the typing module.
	if isinstance(type_annotation, typing._SpecialForm):
		# Handle the typing.Any type.
		if type_annotation is typing.Any:
			return 

		# Handle the typing.Union type.
		if type_annotation.__origin__ is typing.Union:
			# Validate each type in the Union.
			for union_type in type_annotation.__args__:
				try:
					_validate_type(value, union_type)
					return
				except ValueError:
					pass
			raise ValueError(f"Value {value} is not of type {type_annotation}")

		# Handle the typing.List type.
		if type_annotation.__origin__ is typing.List:
			# Validate each item in the list.
			for item in value:
				_validate_type(item, type_annotation.__args__[0])
			return

		# Handle the typing.Dict type.
		if type_annotation.__origin__ is typing.Dict:
			# Validate each key and value in the dict.
			for key, item in value.items():
				_validate_type(key, type_annotation.__args__[0])
				_validate_type(item, type_annotation.__args__[1])
			return

		# Handle the typing.Tuple type.
		if type_annotation.__origin__ is typing.Tuple:
			# Validate each item in the tuple.
			for item, item_type in zip(value, type_annotation.__args__):
				_validate_type(item, item_type)
			return

		# Raise an error if the type is not supported.
		raise ValueError(f"Unsupported type: {type_annotation}")

	# Handle the case where the type annotation is a class.
	if not isinstance(value, type_annotation):
		raise ValueError(f"Value {value} is not of type {type_annotation}")


def _parse_bytes(bytes_reader: io.BytesIO, root: bool) -> typing.Any:
	"""Parses the RPC bytes into the correct type."""
	# Read the type of the value.
	value_type = bytes_reader.read(1)[0]

	def read_bytes() -> bytes:
		# If this is the root, then we need to just consume the remaining bytes.
		if root:
			return bytes_reader.read()

		# Otherwise, we need to read the length of the bytes.
		return bytes_reader.read(int.from_bytes(bytes_reader.read(4), "little"))

	# Switch on the type of the value.
	if value_type == 0x00:
		return None
	elif value_type < 0x03:
		return True if value_type == 0x02 else False
	elif value_type == 0x03:
		return bytes()
	elif value_type == 0x04:
		return ""
	elif value_type == 0x05:
		return read_bytes()
	elif value_type == 0x06:
		return read_bytes().decode("utf-8")
	elif value_type == 0x07:
		arr_size = int.from_bytes(bytes_reader.read(4), "little")
		return [_parse_bytes(bytes_reader, False) for _ in range(arr_size)]
	elif value_type == 0x08:
		dict_size = int.from_bytes(bytes_reader.read(4), "little")
		return {
			_parse_bytes(bytes_reader, False): _parse_bytes(bytes_reader, False)
			for _ in range(dict_size)
		}
	elif value_type == 0x09:
		# Get the struct name.
		struct_name_length = bytes_reader.read(1)[0]
		struct_name = bytes_reader.read(struct_name_length).decode("utf-8")

		# Check if the struct name is in the auto-generated classes.
		if struct_name in _autogenned_models:
			# Rewind the bytes reader. The amount we need to rewind is the struct
			# name, struct name length, and the value type.
			bytes_reader.seek(-struct_name_length - 2, io.SEEK_CUR)

			# Create the class instance.
			return _autogenned_models[struct_name](bytes_reader)

		# Skip over the struct information if it isn't in the auto-generated classes.
		struct_item_count = int.from_bytes(bytes_reader.read(2), "little")
		for _ in range(struct_item_count):
			# Get the key length.
			struct_item_name_length = int.from_bytes(bytes_reader.read(2), "little")

			# Skip over the key.
			bytes_reader.seek(struct_item_name_length, io.SEEK_CUR)

			# Get the value length.
			struct_item_value_length = int.from_bytes(bytes_reader.read(4), "little")

			# Skip over the value.
			bytes_reader.seek(struct_item_value_length, io.SEEK_CUR)

		# Return None.
		return None
	elif value_type == 0x0A:
		# Read the next 8 bytes as a int.
		return int.from_bytes(bytes_reader.read(8), "little")
	elif value_type == 0x0B:
		# Read the next 8 bytes as a float.
		return float.from_bytes(bytes_reader.read(8), "little")
	elif value_type == 0x0C:
		# Read the next 8 bytes as a timestamp.
		uint = int.from_bytes(bytes_reader.read(8), "little", signed=False)
		return datetime.datetime.fromtimestamp(uint)
	elif value_type == 0x0D:
		# Read as a string.
		val = bytes_reader.read().decode("utf-8")

		# Parse as a int.
		return int(val)
	elif value_type == 0x0E:
		# Parse as a uint64.
		return int.from_bytes(bytes_reader.read(8), "little", signed=False)
	elif value_type >= 0x10 and value_type <= 0x1F:
		# Remove 0x10 and return as a int.
		return value_type - 0x10
	elif value_type >= 0x20 and value_type <= 0x2F:
		# Remove 0x20 and then subtract that from -1.
		return -1 - (value_type - 0x20)
	elif value_type >= 0x30 and value_type <= 0x3F:
		# Remove 0x30 and return as a int (Python doesn't have unsigned ints).
		return value_type - 0x30
	elif value_type >= 0x40 and value_type <= 0x4F:
		# Remove 0x40 and then you have a bigint which Python doesn't support so a int.
		return value_type - 0x40
	elif value_type >= 0x50 and value_type <= 0x5F:
		# Remove 0x50 and then subtract that from -1 (Python doesn't have bigints).
		return -1 - (value_type - 0x50)
	elif value_type >= 0x60 and value_type <= 0x6F:
		# Remove 0x60 and return as a float.
		return float(value_type - 0x60)
	elif value_type >= 0x70 and value_type <= 0x7F:
		# Remove 0x70 and then subtract that from -1.
		return -1.0 - (value_type - 0x70)
	else:
		raise ValueError(f"Invalid value type: {value_type}")


def _autogen(cls: typing.Type["_DataModel"]) -> typing.Type["_DataModel"]:
	"""Adds the auto-generated class to the _autogenned_models dict."""
	_autogenned_models[cls.__name__] = cls
	return cls


# Used internally to track the state of a int when encoding.
_int_state_int = 0
_int_state_uint = 1
_int_state_bigint = 2


def _encode_value(
	value: typing.Any, root: bool,
	int_state: typing.Union[int, typing.Tuple[int, int]]
) -> bytes:
	"""Encodes the value into bytes."""
	if value is None:
		return bytes([0x00])
	elif isinstance(value, bool):
		return bytes([0x02 if value else 0x01])
	elif isinstance(value, bytes):
		# Handle if it is blank.
		if len(value) == 0:
			return bytes([0x03])

		# Return 0x05 + length (if not root) + value.
		if root:
			return bytes([0x05]) + value
		return bytes([0x05]) + len(value).to_bytes(4, "little") + value
	elif isinstance(value, str):
		# Handle if it is blank.
		if len(value) == 0:
			return bytes([0x04])

		# Return 0x06 + length (if not root) + value.
		if root:
			return bytes([0x06]) + value.encode("utf-8")
		return bytes([0x06]) + len(value).to_bytes(4, "little") + value.encode("utf-8")
	elif isinstance(value, list):
		# Create a bytes writer.
		bytes_writer = io.BytesIO()

		# Write the type.
		bytes_writer.write(bytes([0x07]))

		# Write the length.
		bytes_writer.write(len(value).to_bytes(4, "little"))

		# Write each item.
		for item in value:
			bytes_writer.write(_encode_value(item, int_state))

		# Return the bytes.
		return bytes_writer.getvalue()
	elif isinstance(value, dict):
		# Create a bytes writer.
		bytes_writer = io.BytesIO()

		# Write the type.
		bytes_writer.write(bytes([0x08]))

		# Write the length.
		bytes_writer.write(len(value).to_bytes(4, "little"))

		# Write each item.
		for key, item in value.items():
			bytes_writer.write(_encode_value(key, int_state[0]))
			bytes_writer.write(_encode_value(item, int_state[1]))

		# Return the bytes.
		return bytes_writer.getvalue()
	elif isinstance(value, _DataModel):
		# Return the bytes.
		return value.encode_to_remixdb_bytes()
	elif isinstance(value, int):
		if int_state == _int_state_bigint:
			# Check if this can be packed into a single byte.
			if value >= 0 and value <= 16:
				return bytes([value + 0x40])
			if value >= -17 and value <= -1:
				return bytes([0x50 + (value * -1)])

			# Encode this into a string.
			enc = str(value).encode("utf-8")
			if root:
				return bytes([0x0D]) + enc
			return bytes([0x0D]) + len(enc).to_bytes(4, "little") + enc

		if int_state == _int_state_uint:
			# Check if this can be packed into a single byte.
			if value >= 0 and value <= 16:
				return bytes([value + 0x30])

			# Handle if this is a negative number.
			if value < 0:
				raise ValueError(f"Invalid uint value: {value}")

			# Encode this into a 8-byte uint.
			return bytes([0x0E]) + value.to_bytes(8, "little", signed=False)

		# Check if this is 0-16.
		if value >= 0 and value <= 16:
			return bytes([value + 0x10])

		# Check if this is -1 to -17.
		if value >= -17 and value <= -1:
			# TODO: Test this
			return bytes([0x20 + (value * -1)])

		# Encode into a int.
		return bytes([0x0A]) + value.to_bytes(8, "little")
	elif isinstance(value, float):
		# Check if this is 0-16.
		if value >= 0 and value <= 16:
			return bytes([value + 0x60])

		# Check if this is -1 to -17.
		if value >= -17 and value <= -1:
			return bytes([0x70 + (value * -1)])

		# Encode into a float.
		return bytes([0x0B]) + value.to_bytes(8, "little")
	elif isinstance(value, datetime.datetime):
		# Encode into a timestamp.
		return bytes([0x0C]) + int(value.timestamp()).to_bytes(8, "little")
	else:
		raise ValueError(f"Unsupported type: {type(value)}")


class _DataModel(object):
	"""
	Defines a base class for all data models. This is included in the
	auto-generated code and the struct types will be based upon it.
	"""
	def __init__(self, data_packet: io.BytesIO) -> None:
		"""Handles a struct from a data packet."""
		# Get the annotations for the class.
		try:
			annotations = self.__annotations__.copy()
		except AttributeError:
			# In this case, the class does not have any annotations.
			annotations = {}

		# Make sure the packet type is 0x09.
		packet_type = data_packet.read(1)[0]
		if packet_type != 0x09:
			raise ValueError(f"Invalid packet type: {packet_type}")

		# Read the length of the struct name.
		struct_name_length = data_packet.read(1)[0]

		# Read the struct name.
		struct_name = data_packet.read(struct_name_length).decode("utf-8")

		# Make sure the struct name is the same as the class name.
		if struct_name != self.__class__.__name__:
			raise ValueError(f"Invalid struct name: {struct_name}")

		# Get the number of struct items in the packet as a uint16 little endian.
		struct_item_count = int.from_bytes(data_packet.read(2), "little")

		# Iterate over each struct item.
		for _ in range(struct_item_count):
			# Read the length of the struct item name (uint16 little endian).
			struct_item_name_length = int.from_bytes(data_packet.read(2), "little")

			# Read the struct item name.
			struct_item_name = data_packet.read(struct_item_name_length).decode("utf-8")

			# Read the length of the struct item value (uint32 little endian).
			struct_item_value_length = int.from_bytes(data_packet.read(4), "little")

			# Read the struct item value.
			struct_item_value = data_packet.read(struct_item_value_length)

			# Skip this item if it isn't in the annotations.
			if struct_item_name not in annotations:
				continue

			# Parse the item.
			struct_item_value = _parse_bytes(io.BytesIO(struct_item_value), True)

			# Parse and then set the attribute if all is well.
			self.__setattr__(struct_item_name, struct_item_value)

			# Remove the item from the annotations.
			del annotations[struct_item_name]

	def _handle_none_annotations(self, remaining_annotations: typing.Dict[str, typing.Any]) -> None:
		"""This method handles the case where the items can be optional and it isn't in the packet."""
		for value in remaining_annotations.values():
			# Validate None against the type annotation.
			_validate_type(None, value)

	def __setattr__(self, key: str, value: typing.Any) -> None:
		"""Overrides the __setattr__ method to validate the type of the value."""
		self._validate_and_add(key, value)

	def _validate_and_add(
		self, key: str, value: typing.Any, return_if_not_exists: bool = False
	) -> None:
		"""Validates the type of the value and adds it to the instance."""
		# Get the type annotation of the attribute.
		try:
			type_annotation = self.__annotations__[key]
		except (KeyError, AttributeError):
			# This would mean that it is not a valid attribute.
			if return_if_not_exists: return
			raise ValueError(f"Invalid attribute: {key}")

		# Validate the type of the value.
		_validate_type(value, type_annotation)

		# Add the value to the instance bypassing the set __setattr__ method.
		super().__setattr__(key, value)

	def encode_to_remixdb_bytes(self) -> bytes:
		"""Encodes the struct to RemixDB bytes."""
		# Create a bytes writer.
		bytes_writer = io.BytesIO()

		# Write the packet type.
		bytes_writer.write(bytes([0x09]))

		# Write the struct name.
		bytes_writer.write(bytes([len(self.__class__.__name__)]))
		bytes_writer.write(self.__class__.__name__.encode("utf-8"))

		# Get the annotations for the class.
		try:
			annotations = self.__annotations__
		except AttributeError:
			# In this case, the class does not have any annotations.
			annotations = {}

		# Write the number of struct items.
		bytes_writer.write(len(annotations).to_bytes(2, "little"))

		# Get the injected dict.
		try:
			injected = self._injected
		except AttributeError:
			# In this case, there are no injected items.
			injected = {}

		# Iterate over each struct item.
		for key in annotations.keys():
			# Get the value.
			value = getattr(self, key)

			# Write the struct item name.
			bytes_writer.write(len(key).to_bytes(2, "little"))
			bytes_writer.write(key.encode("utf-8"))

			# Check the injected dict for the key.
			injected_val = injected.get(key, _int_state_int)

			# Encode the value.
			bytes_writer.write(_encode_value(value, injected_val))

		# Return the bytes.
		return bytes_writer.getvalue()


# AUTO-GENERATION MARKER: structs
class _DottedDict(dict):
	"""Defines a dotted dict to allow for dot access."""
	def __getattr__(self, key: str) -> typing.Any:
		"""Overrides the __getattr__ method to allow for dot access."""
		try:
			return self[key]
		except KeyError:
			raise AttributeError(f"Invalid attribute: {key}")

	def __setattr__(self, key: str, value: typing.Any) -> None:
		"""Overrides the __setattr__ method to allow for dot access."""
		self[key] = value


class Config(_DottedDict):
	"""Defines the config for the API."""
	# AUTO-GENERATION MARKER: config


class ServerError(Exception):
	"""Defines an exception for server errors."""
	__slots__ = ("code", "message")

	def __init__(self, code: str, message: str) -> None:
		self.code = code
		self.message = message

	def __str__(self) -> str:
		return f"ServerError({self.code}: {self.message})"


def _parse_exception(custom_exception: typing.Optional[str], body: bytes) -> None:
	"""Parses the specified exception and then throws it."""
	if custom_exception is None:
		# Parse the exception.
		exception = json.loads(body.decode("utf-8"))

		# Make sure code and message are strings.
		if not isinstance(exception["code"], str):
			raise ServerError("invalid_exception", "The exception code must be a string.")
		if not isinstance(exception["message"], str):
			raise ServerError("invalid_exception", "The exception message must be a string.")

		# Raise the exception.
		raise ServerError(exception["code"], exception["message"])

	# Check if the exception is in the structs.
	if custom_exception not in _autogenned_models:
		raise ServerError("invalid_exception", "The exception is not in the structs.")

	# Attempt to parse the exception.
	j = json.loads(body.decode("utf-8"))
	raise _autogenned_models[custom_exception](**j)


class _SyncWebSocket(object):
	"""Defines a client that takes a URL and creates a WebSocket connection."""
	def __init__(self, url: urllib.parse.ParseResult, timeout: typing.Union[int, None] = 10) -> None:
		self._url = urllib.parse.urlparse(url.geturl())
		self._url.path = self._url.path + "/rpc"
		self._timeout = timeout
		self._socket = self._connect()

	def _connect(self) -> typing.Any:
		"""Makes a WebSocket connection."""
		# Create the websocket key.
		key = os.urandom(16)
		key_enc = base64.b64encode(key).decode("utf-8")

		# Create the headers.
		headers = {
			"Connection": "Upgrade",
			"Upgrade": "websocket",
			"Sec-WebSocket-Version": "13",
			"Sec-WebSocket-Key": key_enc
		}

		# Make the request.
		response = urllib.request.urlopen(
			self._url.geturl(),
			None,
			headers,
			timeout=self._timeout
		)

		# Check the status code.
		if response.status != 101:
			raise ServerError(
				"invalid_status_code",
				f"Invalid status code: {response.status}"
			)

		# Check the upgrade header.
		if response.getheader("Upgrade") != "websocket":
			raise ServerError(
				"invalid_upgrade_header",
				f"Invalid upgrade header: {response.getheader('Upgrade')}"
			)

		# Check the connection header.
		if response.getheader("Connection") != "Upgrade":
			raise ServerError(
				"invalid_connection_header",
				f"Invalid connection header: {response.getheader('Connection')}"
			)

		# Check the Sec-WebSocket-Accept header.
		accept = response.getheader("Sec-WebSocket-Accept")
		if accept != base64.b64encode(
			hashlib.sha1(key + b"258EAFA5-E914-47DA-95CA-C5AB0DC85B11").digest()
		).decode("utf-8"):
			raise ServerError(
				"invalid_sec_websocket_accept_header",
				f"Invalid Sec-WebSocket-Accept header: {accept}"
			)

		# Return the socket.
		return response.fp._sock.fp.raw._sock

	def send_small(self, data: bytes) -> None:
		"""Sends small packets over the WebSocket."""
		# Create the header.
		header = bytes([0x82, len(data)])

		# Send the data.
		self._socket.sendall(header + data)

	def read(self) -> bytes:
		"""Reads the next packet from the WebSocket."""
		# Read the header.
		header = self._socket.recv(2)

		# Check the opcode.
		opcode = header[0] & 0x0F
		if opcode != 0x02:
			raise ServerError(
				"invalid_opcode",
				f"Invalid opcode: {opcode}"
			)

		# Read the length.
		length = header[1] & 0x7F
		if length == 126:
			length = int.from_bytes(self._socket.recv(2), "big")
		elif length == 127:
			length = int.from_bytes(self._socket.recv(8), "big")

		# Read the data.
		data = self._socket.recv(length)

		# Return the data.
		return data

	def close(self) -> None:
		"""Closes the WebSocket."""
		self._socket.close()


T = typing.TypeVar("T")


class SyncCursor(typing.Generic[T]):
	"""Defines a sync cursor with generic typings."""
	# TODO

class Client(object):
	"""Defines the client class for non-async clients."""
	def __init__(
		self, base_url: str, config: Config,
		timeout: typing.Union[int, None] = 10
	) -> None:
		self._url = urllib.parse.urlparse(base_url)
		self._config = json.dumps(config).encode("utf-8") + bytes(["\n"])
		self._timeout = timeout

	def _non_cursor_do(self, schema_hash: str, method: str, body: bytes) -> bytes:
		"""Handles a network request that handles non-cursors."""
		# Create the URL.
		url = urllib.parse.urlunparse((
			self._url.scheme,
			self._url.netloc,
			f"/rpc/{method}",
			"",
			"",
			""
		))

		# Create the headers.
		headers = {
			"Content-Type": "application/x-remixdb-rpc-mixed",
			"X-RemixDB-Schema-Hash": schema_hash
		}

		# Make the request.
		response = urllib.request.urlopen(
			url,
			self._config + body,
			headers,
			timeout=self._timeout,
			method="POST"
		)

		# Check X-Is-RemixDB is true.
		if response.getheader("X-Is-RemixDB") != "true":
			raise ServerError(
				"response_is_not_remixdb",
				"The response does not appear to be from RemixDB. Does your reverse proxy let through the X-Is-RemixDB header?"
			)

		# Check the status code.
		if response.status != 200 and response.status != 204:
			_parse_exception(
				response.getheader("X-RemixDB-Exception"),
				response.read()
			)

		# Return the response.
		return response.read()

	# TODO: cursors

	# AUTO-GENERATION MARKER: sync_client


# TODO: async
