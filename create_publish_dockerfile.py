# RemixDB. Copyright (C) 2023 Web Scale Software Ltd.
# Author: Astrid Gealer <astrid@gealer.email>

import sys


def mutate_dockerignore() -> None:
    f = open(".dockerignore", "r")
    contents = f.read().replace("frontend/dist\n", "")
    f.close()
    open(".dockerignore", "w").write(contents)


def generate_deploy_dockerfile() -> None:
    # Generate the publish Dockerfile.
    a = []
    dockerfile_started = False
    for line in open("Dockerfile", "r").readlines():
        if dockerfile_started:
            # Handle if the publish Dockerfile was started.
            if line == "# -- ^ REMOVE IN PUBLISH DOCKERFILE ^ --\n":
                a.pop()
            else:
                a.append(line)
        elif line == "# -- PUBLISH DOCKERFILE START --\n":
            # Start the publish Dockerfile.
            dockerfile_started = True
    dockerfile = "".join(a).strip() + "\n"

    # Write the publish Dockerfile.
    open("Dockerfile.ghcr-deploy", "w").write(dockerfile)


def main() -> None:
    # Check if the arg is 'mutate_dockerignore'.
    if len(sys.argv) == 2 and sys.argv[1] == "mutate_dockerignore":
        mutate_dockerignore()

    generate_deploy_dockerfile()


if __name__ == "__main__":
    main()
