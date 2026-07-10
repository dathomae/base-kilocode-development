# Debug Adapters

Debug adapters enable VSCode to debug programs written in any language. They implement the Debug Adapter Protocol (DAP), which separates the debug UI from the actual debugger implementation.

## Debug Adapter Protocol Overview

The Debug Adapter Protocol defines how VSCode communicates with debug adapters:

1. VSCode sends requests to the adapter (e.g., "launch", "setBreakpoints", "continue")
2. The adapter responds with results
3. The adapter sends events to VSCode (e.g., "stopped", "output", "terminated")

This separation allows the same debug adapter to work with different editors that support DAP.

## Registering a Debugger

Declare the debugger in `package.json`:

```json
{
  "contributes": {
    "debuggers": [
      {
        "type": "myDebugger",
        "label": "My Debugger",
        "program": "./out/debugAdapter.js",
        "runtime": "node",
        "configurationAttributes": {
          "launch": {
            "required": ["program"],
            "properties": {
              "program": {
                "type": "string",
                "description": "The program to debug"
              },
              "stopOnEntry": {
                "type": "boolean",
                "description": "Stop on program entry",
                "default": true
              }
            }
          }
        },
        "initialConfigurations": [
          {
            "type": "myDebugger",
            "request": "launch",
            "name": "Launch Program",
            "program": "${workspaceFolder}/main.js"
          }
        ],
        "configurationSnippets": [
          {
            "label": "My Debugger: Launch",
            "description": "Launch a program with My Debugger",
            "body": {
              "type": "myDebugger",
              "request": "launch",
              "name": "Launch Program",
              "program": "^\"\\${workspaceFolder}/${1:main.js}\""
            }
          }
        ]
      }
    ]
  }
}
```

## Debug Adapter Factory

Create a factory that produces debug adapter descriptors:

```typescript
import * as vscode from "vscode";

class MyDebugAdapterDescriptorFactory
  implements vscode.DebugAdapterDescriptorFactory
{
  createDebugAdapterDescriptor(
    session: vscode.DebugSession,
    executable: vscode.DebugAdapterExecutable | undefined,
  ): vscode.ProviderResult<vscode.DebugAdapterDescriptor> {
    // Return an executable that runs the debug adapter
    return new vscode.DebugAdapterExecutable("node", [
      path.join(context.extensionPath, "out", "debugAdapter.js"),
    ]);
  }
}

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.debug.registerDebugAdapterDescriptorFactory(
      "myDebugger",
      new MyDebugAdapterDescriptorFactory(),
    ),
  );
}
```

### Types of Debug Adapter Descriptors

**DebugAdapterExecutable**: Run an external program as the debug adapter

```typescript
new vscode.DebugAdapterExecutable("/path/to/executable", ["--arg1", "value1"], {
  env: { CUSTOM_VAR: "value" },
});
```

**DebugAdapterServer**: Connect to an already-running debug server

```typescript
new vscode.DebugAdapterServer(4711, "localhost");
```

**DebugAdapterInlineImplementation**: Run the adapter in the extension host process

```typescript
new vscode.DebugAdapterInlineImplementation(new MyDebugAdapter());
```

## Implementing a Debug Adapter

Debug adapters implement the DAP protocol. The simplest approach is to use the `vscode-debugadapter` npm package:

```typescript
import {
  DebugSession,
  InitializedEvent,
  StoppedEvent,
  TerminatedEvent,
  OutputEvent,
  Thread,
  StackFrame,
  Scope,
  Source,
  Breakpoint,
} from "vscode-debugadapter";

class MyDebugSession extends DebugSession {
  private debugger: MyDebugger;

  protected initializeRequest(
    response: DebugProtocol.InitializeResponse,
    args: DebugProtocol.InitializeRequestArguments,
  ): void {
    // Build and return capabilities
    response.body = {
      supportsConfigurationDoneRequest: true,
      supportsSetVariable: true,
      supportsConditionalBreakpoints: true,
      supportsEvaluateForHovers: true,
    };

    this.sendResponse(response);
    this.sendEvent(new InitializedEvent());
  }

  protected launchRequest(
    response: DebugProtocol.LaunchResponse,
    args: DebugProtocol.LaunchRequestArguments,
  ): void {
    // Start the program
    this.debugger = new MyDebugger(args.program);
    this.debugger.on("stop", () => {
      this.sendEvent(new StoppedEvent("entry", 1));
    });
    this.debugger.on("output", (text) => {
      this.sendEvent(new OutputEvent(text));
    });
    this.debugger.on("terminate", () => {
      this.sendEvent(new TerminatedEvent());
    });

    this.debugger.launch();
    this.sendResponse(response);
  }

  protected setBreakPointsRequest(
    response: DebugProtocol.SetBreakpointsResponse,
    args: DebugProtocol.SetBreakpointsArguments,
  ): void {
    // Set breakpoints in the debugger
    const breakpoints =
      args.breakpoints?.map((bp) => {
        const verified = this.debugger.setBreakpoint(args.source.path, bp.line);
        return new Breakpoint(verified, bp.line);
      }) || [];

    response.body = { breakpoints };
    this.sendResponse(response);
  }

  protected threadsRequest(response: DebugProtocol.ThreadsResponse): void {
    // Return available threads
    response.body = {
      threads: [new Thread(1, "Main Thread")],
    };
    this.sendResponse(response);
  }

  protected stackTraceRequest(
    response: DebugProtocol.StackTraceResponse,
    args: DebugProtocol.StackTraceArguments,
  ): void {
    // Return the call stack
    const frames = this.debugger
      .getStackTrace()
      .map(
        (frame, i) =>
          new StackFrame(
            i,
            frame.name,
            new Source(frame.file, frame.path),
            frame.line,
            frame.column,
          ),
      );

    response.body = { stackFrames: frames };
    this.sendResponse(response);
  }

  protected scopesRequest(
    response: DebugProtocol.ScopesResponse,
    args: DebugProtocol.ScopesArguments,
  ): void {
    // Return variable scopes
    response.body = {
      scopes: [new Scope("Local", 1, false), new Scope("Global", 2, true)],
    };
    this.sendResponse(response);
  }

  protected variablesRequest(
    response: DebugProtocol.VariablesResponse,
    args: DebugProtocol.VariablesArguments,
  ): void {
    // Return variables in a scope
    const variables = this.debugger.getVariables(args.variablesReference);
    response.body = { variables };
    this.sendResponse(response);
  }

  protected continueRequest(
    response: DebugProtocol.ContinueResponse,
    args: DebugProtocol.ContinueArguments,
  ): void {
    this.debugger.continue();
    this.sendResponse(response);
  }

  protected nextRequest(
    response: DebugProtocol.NextResponse,
    args: DebugProtocol.NextArguments,
  ): void {
    this.debugger.stepOver();
    this.sendResponse(response);
  }

  protected stepInRequest(
    response: DebugProtocol.StepInResponse,
    args: DebugProtocol.StepInArguments,
  ): void {
    this.debugger.stepInto();
    this.sendResponse(response);
  }

  protected stepOutRequest(
    response: DebugProtocol.StepOutResponse,
    args: DebugProtocol.StepOutArguments,
  ): void {
    this.debugger.stepOut();
    this.sendResponse(response);
  }

  protected disconnectRequest(
    response: DebugProtocol.DisconnectResponse,
    args: DebugProtocol.DisconnectArguments,
  ): void {
    this.debugger.terminate();
    this.sendResponse(response);
  }
}

// Start the debug adapter
DebugSession.run(MyDebugSession);
```

## Debug Configuration

### Configuration Attributes

Define required and optional configuration properties:

```json
{
  "configurationAttributes": {
    "launch": {
      "required": ["program"],
      "properties": {
        "program": {
          "type": "string",
          "description": "The program to debug"
        },
        "args": {
          "type": "array",
          "items": { "type": "string" },
          "description": "Command line arguments"
        },
        "env": {
          "type": "object",
          "description": "Environment variables"
        },
        "stopOnEntry": {
          "type": "boolean",
          "default": true
        }
      }
    },
    "attach": {
      "required": ["port"],
      "properties": {
        "port": {
          "type": "number",
          "description": "Debug port to attach to"
        },
        "host": {
          "type": "string",
          "default": "localhost"
        }
      }
    }
  }
}
```

### Initial Configurations

Provide default configurations:

```json
{
  "initialConfigurations": [
    {
      "type": "myDebugger",
      "request": "launch",
      "name": "Launch Program",
      "program": "${workspaceFolder}/main.js"
    },
    {
      "type": "myDebugger",
      "request": "attach",
      "name": "Attach to Process",
      "port": 9229
    }
  ]
}
```

### Configuration Snippets

Provide snippets for the launch.json editor:

```json
{
  "configurationSnippets": [
    {
      "label": "My Debugger: Launch",
      "description": "Launch a program",
      "body": {
        "type": "myDebugger",
        "request": "launch",
        "name": "Launch ${1:Program}",
        "program": "^\"\\${workspaceFolder}/${2:main.js}\""
      }
    }
  ]
}
```

## Debug Session Events

Listen for debug session events:

```typescript
// Session started
vscode.debug.onDidStartDebugSession((session) => {
  console.log("Debug session started:", session.id);
});

// Session terminated
vscode.debug.onDidTerminateDebugSession((session) => {
  console.log("Debug session terminated:", session.id);
});

// Active session changed
vscode.debug.onDidChangeActiveDebugSession((session) => {
  console.log("Active session changed:", session?.id);
});

// Breakpoint changed
vscode.debug.onDidChangeBreakpoints((e) => {
  console.log("Breakpoints changed:", e.added, e.removed, e.changed);
});
```

## Custom Debug Commands

Register commands that interact with the debugger:

```typescript
vscode.commands.registerCommand("myExtension.stepBack", async () => {
  const session = vscode.debug.activeDebugSession;
  if (session) {
    await session.customRequest("stepBack");
  }
});
```

Handle custom requests in the debug adapter:

```typescript
protected customRequest(command: string, response: DebugProtocol.Response, args: any): void {
    switch (command) {
        case 'stepBack':
            this.debugger.stepBack();
            this.sendResponse(response);
            break;
        default:
            response.success = false;
            response.message = `Unknown command: ${command}`;
            this.sendResponse(response);
    }
}
```

## Best Practices

**Implement required requests**: At minimum, implement `initialize`, `launch`/`attach`, `threads`, `stackTrace`, `scopes`, and `variables`.

**Report capabilities accurately**: Only advertise features your adapter supports.

**Handle errors gracefully**: Return error responses with clear messages.

**Use the vscode-debugadapter package**: It provides base classes and utilities that simplify implementation.

**Test thoroughly**: Debug adapters are complex; test all scenarios including edge cases.

**Document configuration options**: Provide clear descriptions for all configuration properties.

## Next Steps

- [Tasks](tasks.md): Provide custom tasks
- [Terminal](terminal.md): Integrate with the terminal
- [VSCode Debug Adapter Documentation](https://code.visualstudio.com/api/extension-guides/debugger-extension)
