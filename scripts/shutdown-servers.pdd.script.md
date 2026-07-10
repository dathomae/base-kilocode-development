# Shutdown Servers

This script gracefully shuts down server processes running on standard ports used by the ChronoCone application. By default, it terminates processes on ports 8001 (Core Service) and 8002 (Metrics Service), but can be configured to shut down services on any specified ports.

## Parameters

| Name  | Type    | Description                                     | Required | Default Value  |
| ----- | ------- | ----------------------------------------------- | -------- | -------------- |
| ports | list    | List of port numbers to shut down               | No       | `[8001, 8002]` |
| force | boolean | Force kill processes if graceful shutdown fails | No       | `false`        |
| quiet | boolean | Suppress non-essential output messages          | No       | `false`        |

## Prerequisites

Before running this script, ensure the following prerequisites are met:

1. The `fuser` command is available on the system (part of `psmisc` package on most Linux distributions).
2. You have appropriate permissions to terminate processes on the specified ports (may require `sudo` for processes owned by other users).
3. No critical operations are in progress on the services being shut down.

## Operations

1. Validate that all parameters have acceptable values:
   - Ports must be valid integers in the range 1-65535.
   - If no ports are specified, use the default ports `[8001, 8002]`.

2. Display the shutdown configuration, including the list of ports to be checked.

3. For each port in the list:
   a. Check if a process is listening on the port using `fuser <port>/tcp` or `ss -tlnp | grep <port>`.
   b. If a process is found, record the process ID for reporting.

4. Display a summary of found processes:
   - List each port with its associated process ID (if any).
   - Indicate which ports have no active processes.

5. If processes are found, proceed with shutdown:
   a. Build the shutdown command using `fuser -k` with the port list:
   - Command format: `fuser -k <port1>/tcp <port2>/tcp ... 2>/dev/null`
     b. If `force` is true, add the `-9` flag for SIGKILL instead of SIGTERM.

6. Execute the shutdown command and capture the exit code.

7. Wait briefly (1-2 seconds) for processes to terminate.

8. Verify that ports are freed:
   a. Re-check each port to confirm no processes are listening.
   b. If any processes remain and `force` is false, offer to retry with force enabled.

9. Display a summary of the shutdown results, including:
   - Which ports were successfully freed.
   - Which ports (if any) still have running processes.
   - Any errors encountered during shutdown.

## Output

This script produces the following outputs:

- **Console output**: Real-time progress messages during shutdown, including:
  - List of ports being checked
  - Process IDs found on each port
  - Shutdown command being executed
  - Verification results

- **Shutdown summary**: A formatted summary displaying:
  - Ports that were successfully freed
  - Ports that had no processes (already free)
  - Ports that could not be freed (with error details)
  - Process IDs that were terminated

- **Return code**: The script returns:
  - Success code (0) if all specified ports are freed
  - Error code (1) if any port could not be freed or if permission is denied

## Error Handling

Common error scenarios and their handling:

- **No process found on port**: The script will report that the port is already free and continue with other ports. This is not treated as an error.

- **Permission denied**: If the user lacks permissions to terminate a process, the script will display an error suggesting the use of `sudo`. The script will continue with other ports.

- **Process refuses to terminate**: If a process does not respond to SIGTERM, the script will offer to retry with `force=true` (SIGKILL).

- **Invalid port number**: If an invalid port is specified, the script will display an error and skip that port, continuing with valid ports.

- **fuser command not found**: If the `fuser` command is not available, the script will attempt to use alternative methods (`ss` or `lsof`) or report an error if no alternatives are available.

- **Port still in use after shutdown**: If verification shows a port is still in use, the script will report the remaining process and suggest manual intervention.

During execution, if a critical error is encountered, the script pauses and asks whether to continue with remaining ports or terminate. The user can review the error context before deciding how to proceed.
