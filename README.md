## Available commands:

- `clone`: Clone the specified GitHub repositories.
- `docker-up`: Run Docker Compose in the downloaded repositories.
- `list-services`: List all available services in the downloaded repositories.
- `run-service`: Run a particular service within all downloaded repositories.

For more information about each command, run `myproject [command] --help`.

## Configuration

The tool requires a GitHub personal access token to access the repositories. Set the `GITHUB_TOKEN` environment variable to your personal access token.

Additionally, set the `REPOSITORIES` environment variable to a comma-separated list of the GitHub repositories you want to download.

## How to run commands
1. Open a terminal and navigate to the directory where you saved the file.
2. Run the following command to compile the program: go build main.go
3. Run the compiled executable with the list-services action: ./main -action=list-services on Unix-based systems, or main.exe -action=list-services on Windows.