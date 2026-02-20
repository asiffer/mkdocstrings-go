from __future__ import annotations

import ctypes
import platform
from pathlib import Path

import pydantic

_ext = {"Darwin": "dylib", "Windows": "dll"}.get(platform.system(), "so")
_lib_path = Path(__file__).parent / f"libcollector.{_ext}"
_lib = ctypes.CDLL(str(_lib_path))

_lib.collect.restype = ctypes.c_int
_lib.collect.argtypes = [
    ctypes.c_char_p,  # importPath
    ctypes.c_char_p,  # dir
    ctypes.c_char_p,  # object
    ctypes.c_char_p,  # outBuf
    ctypes.c_int,  # outSize
]

_DEFAULT_BUF_SIZE = 65536


_type_to_symbol = {
    "function": "func",
    "method": "meth",
}


class _GoObject(pydantic.BaseModel):
    """A class representing a Go object, used for type annotations in the handler."""

    path: str
    name: str
    type: str
    doc: str
    tag: dict[str, str] | None
    definition: str
    body: str
    items: list[_GoObject]
    location: str
    start_line: int
    end_line: int

    @property
    def signature(self) -> str:
        """Get the signature of the object, if applicable."""
        return self.definition

    @property
    def symbol(self) -> str:
        """Get the symbol of the object, if applicable."""
        return _type_to_symbol.get(self.type, self.type)


def _collect(import_path: str, directory: str, identifier: str, buf_size: int = _DEFAULT_BUF_SIZE) -> _GoObject:
    """Collect documentation data for a Go object.

    Parameters:
        import_path: The module import path, e.g. ``github.com/user/project``.
        directory: Absolute path to the project root on disk.
        identifier: Fully qualified object, e.g. ``github.com/user/project/pkg.MyType``.
        buf_size: Output buffer size in bytes. Increase for very large objects.

    Returns:
        A dict representing the collected ``GoObject`` tree.

    Raises:
        ValueError: If the Go library returns an error.
    """
    buf = ctypes.create_string_buffer(buf_size)
    n = _lib.collect(
        import_path.encode(),
        directory.encode(),
        identifier.encode(),
        buf,
        ctypes.c_int(buf_size),
    )
    if n < 0:
        raise ValueError(f"collect failed for {identifier!r} (error code {n})")

    return _GoObject.model_validate_json(buf.raw[:n])
