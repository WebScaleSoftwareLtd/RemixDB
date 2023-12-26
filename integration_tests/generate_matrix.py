# RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
# Author: Astrid Gealer <astrid@gealer.email>

import os
import json

def _validate_versions(versions) -> list[str]:
    """Make sure that versions is a string array."""
    if not isinstance(versions, list):
        raise ValueError("versions is not an array")
    for v in versions:
        if not isinstance(v, str):
            raise ValueError("versions contains a non-string")
    return versions

def _pipe_to_env_file_path(pipe: str, data: str) -> None:
    """Pipe data to a file path in the environment."""
    with open(os.environ[pipe], "w") as f:
        f.write(data)

def main() -> None:
    # Get the folder this script is in.
    script_dir = os.path.dirname(os.path.realpath(__file__))

    # Defines matrix include items.
    includes = []

    # Get the folders in the folder this script is in.
    files = os.listdir(script_dir)
    for f in files:
        if f.startswith("_"):
            # Skip files starting with an underscore.
            continue

        folder_join = os.path.join(script_dir, f)
        if not os.path.isdir(folder_join):
            # Skip non-folders.
            continue

        # Try to read the manifest file.
        try:
            with open(os.path.join(folder_join, "manifest.json")) as manifest_file:
                manifest = json.load(manifest_file)
        except FileNotFoundError:
            print(f"Skipping {f} - no manifest.json")
            continue

        # Make sure the manifest has a versions array with strings in.
        try:
            versions = _validate_versions(manifest["versions"])
        except BaseException as e:
            print(f"Skipping {f} - {e.__class__.__name__}: {e}")
            continue

        # Add the versions to the matrix.
        for v in versions:
            includes.append({"version": v, "name": f})

    # Output the matrix so GitHub Actions can use it.
    _pipe_to_env_file_path("GITHUB_OUTPUT", f"matrix={json.dumps({'include': includes})}")

if __name__ == "__main__":
    main()
