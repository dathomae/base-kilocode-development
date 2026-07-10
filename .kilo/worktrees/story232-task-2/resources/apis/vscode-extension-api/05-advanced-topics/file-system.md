# File System Providers

File system providers allow extensions to implement custom file systems that VSCode can access. This enables features like remote file access, virtual file systems, and custom storage backends.

## Registering a File System Provider

Use `vscode.workspace.registerFileSystemProvider` to register a custom file system:

```typescript
import * as vscode from "vscode";

class MyFileSystemProvider implements vscode.FileSystemProvider {
  // Implement required methods
}

export function activate(context: vscode.ExtensionContext) {
  context.subscriptions.push(
    vscode.workspace.registerFileSystemProvider(
      "myfs", // URI scheme
      new MyFileSystemProvider(),
      { isCaseSensitive: true },
    ),
  );
}
```

## FileSystemProvider Interface

Implement the `FileSystemProvider` interface:

```typescript
interface FileSystemProvider {
  // Events
  onDidChangeFile: vscode.Event<vscode.FileChangeEvent[]>;

  // Required methods
  stat(uri: vscode.Uri): vscode.FileStat | Thenable<vscode.FileStat>;
  readDirectory(
    uri: vscode.Uri,
  ): [string, vscode.FileType][] | Thenable<[string, vscode.FileType][]>;
  readFile(uri: vscode.Uri): Uint8Array | Thenable<Uint8Array>;

  // Optional methods for writable file systems
  writeFile?(
    uri: vscode.Uri,
    content: Uint8Array,
    options: { create: boolean; overwrite: boolean },
  ): void | Thenable<void>;
  rename?(
    oldUri: vscode.Uri,
    newUri: vscode.Uri,
    options: { overwrite: boolean },
  ): void | Thenable<void>;
  delete?(
    uri: vscode.Uri,
    options: { recursive: boolean },
  ): void | Thenable<void>;
  createDirectory?(uri: vscode.Uri): void | Thenable<void>;

  // Optional methods for search
  watch?(
    uri: vscode.Uri,
    options: { recursive: boolean; excludes: string[] },
  ): vscode.Disposable;
}
```

## Implementing a Read-Only File System

```typescript
class ReadOnlyFileSystemProvider implements vscode.FileSystemProvider {
  private files: Map<string, { content: Uint8Array; mtime: number }> =
    new Map();
  private emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();

  onDidChangeFile = this.emitter.event;

  constructor() {
    // Initialize with some files
    this.files.set("/readme.txt", {
      content: new TextEncoder().encode("Hello from my file system!"),
      mtime: Date.now(),
    });
  }

  stat(uri: vscode.Uri): vscode.FileStat {
    const path = uri.path;
    const file = this.files.get(path);

    if (file) {
      return {
        type: vscode.FileType.File,
        ctime: 0,
        mtime: file.mtime,
        size: file.content.length,
      };
    }

    // Check if it's a directory
    const dirPath = path.endsWith("/") ? path : path + "/";
    for (const key of this.files.keys()) {
      if (key.startsWith(dirPath)) {
        return {
          type: vscode.FileType.Directory,
          ctime: 0,
          mtime: 0,
          size: 0,
        };
      }
    }

    throw vscode.FileSystemError.FileNotFound(uri);
  }

  readDirectory(uri: vscode.Uri): [string, vscode.FileType][] {
    const path = uri.path;
    const dirPath = path.endsWith("/") ? path : path + "/";
    const entries = new Map<string, vscode.FileType>();

    for (const key of this.files.keys()) {
      if (key.startsWith(dirPath)) {
        const relative = key.slice(dirPath.length);
        const name = relative.split("/")[0];
        if (name && !entries.has(name)) {
          const isDir = relative.includes("/");
          entries.set(
            name,
            isDir ? vscode.FileType.Directory : vscode.FileType.File,
          );
        }
      }
    }

    return Array.from(entries.entries());
  }

  readFile(uri: vscode.Uri): Uint8Array {
    const file = this.files.get(uri.path);
    if (!file) {
      throw vscode.FileSystemError.FileNotFound(uri);
    }
    return file.content;
  }

  watch(uri: vscode.Uri): vscode.Disposable {
    return { dispose: () => {} };
  }
}
```

## Implementing a Writable File System

```typescript
class WritableFileSystemProvider implements vscode.FileSystemProvider {
  private files: Map<string, { content: Uint8Array; mtime: number }> =
    new Map();
  private emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();

  onDidChangeFile = this.emitter.event;

  stat(uri: vscode.Uri): vscode.FileStat {
    const file = this.files.get(uri.path);
    if (file) {
      return {
        type: vscode.FileType.File,
        ctime: 0,
        mtime: file.mtime,
        size: file.content.length,
      };
    }
    throw vscode.FileSystemError.FileNotFound(uri);
  }

  readDirectory(uri: vscode.Uri): [string, vscode.FileType][] {
    // Implementation
    return [];
  }

  readFile(uri: vscode.Uri): Uint8Array {
    const file = this.files.get(uri.path);
    if (!file) {
      throw vscode.FileSystemError.FileNotFound(uri);
    }
    return file.content;
  }

  writeFile(
    uri: vscode.Uri,
    content: Uint8Array,
    options: { create: boolean; overwrite: boolean },
  ): void {
    const existing = this.files.get(uri.path);

    if (!existing && !options.create) {
      throw vscode.FileSystemError.FileNotFound(uri);
    }

    if (existing && !options.overwrite) {
      throw vscode.FileSystemError.FileExists(uri);
    }

    this.files.set(uri.path, {
      content,
      mtime: Date.now(),
    });

    this.emitter.fire([
      {
        type: existing
          ? vscode.FileChangeType.Changed
          : vscode.FileChangeType.Created,
        uri,
      },
    ]);
  }

  createDirectory(uri: vscode.Uri): void {
    // Directories are implicit in this implementation
  }

  delete(uri: vscode.Uri, options: { recursive: boolean }): void {
    if (!this.files.has(uri.path)) {
      throw vscode.FileSystemError.FileNotFound(uri);
    }

    this.files.delete(uri.path);

    this.emitter.fire([
      {
        type: vscode.FileChangeType.Deleted,
        uri,
      },
    ]);
  }

  rename(
    oldUri: vscode.Uri,
    newUri: vscode.Uri,
    options: { overwrite: boolean },
  ): void {
    const file = this.files.get(oldUri.path);
    if (!file) {
      throw vscode.FileSystemError.FileNotFound(oldUri);
    }

    if (this.files.has(newUri.path) && !options.overwrite) {
      throw vscode.FileSystemError.FileExists(newUri);
    }

    this.files.delete(oldUri.path);
    this.files.set(newUri.path, file);

    this.emitter.fire([
      { type: vscode.FileChangeType.Deleted, uri: oldUri },
      { type: vscode.FileChangeType.Created, uri: newUri },
    ]);
  }

  watch(uri: vscode.Uri): vscode.Disposable {
    return { dispose: () => {} };
  }
}
```

## File Change Events

Notify VSCode when files change:

```typescript
// File created
this.emitter.fire([
  {
    type: vscode.FileChangeType.Created,
    uri,
  },
]);

// File changed
this.emitter.fire([
  {
    type: vscode.FileChangeType.Changed,
    uri,
  },
]);

// File deleted
this.emitter.fire([
  {
    type: vscode.FileChangeType.Deleted,
    uri,
  },
]);
```

## File System Errors

Use `FileSystemError` for standard error handling:

```typescript
// File not found
throw vscode.FileSystemError.FileNotFound(uri);

// File already exists
throw vscode.FileSystemError.FileExists(uri);

// Not a file
throw vscode.FileSystemError.FileNotADirectory(uri);

// Not a directory
throw vscode.FileSystemError.NotADirectory(uri);

// No permissions
throw vscode.FileSystemError.NoPermissions(uri);

// Unavailable
throw vscode.FileSystemError.Unavailable(uri);
```

## Using Custom File Systems

Access files in custom file systems using the registered scheme:

```typescript
// Open a file from the custom file system
const uri = vscode.Uri.parse("myfs:/path/to/file.txt");
const document = await vscode.workspace.openTextDocument(uri);
await vscode.window.showTextDocument(document);

// Read file content
const content = await vscode.workspace.fs.readFile(uri);

// Write file content
await vscode.workspace.fs.writeFile(uri, new TextEncoder().encode("Hello"));

// List directory
const entries = await vscode.workspace.fs.readDirectory(
  vscode.Uri.parse("myfs:/"),
);

// Create directory
await vscode.workspace.fs.createDirectory(vscode.Uri.parse("myfs:/newdir"));

// Delete
await vscode.workspace.fs.delete(uri);
```

## Watching Files

Implement file watching:

```typescript
private watchers = new Map<string, Set<vscode.Disposable>>();

watch(uri: vscode.Uri, options: { recursive: boolean; excludes: string[] }): vscode.Disposable {
    const key = uri.toString();

    if (!this.watchers.has(key)) {
        this.watchers.set(key, new Set());
    }

    const disposable = { dispose: () => {
        const set = this.watchers.get(key);
        if (set) {
            set.delete(disposable);
            if (set.size === 0) {
                this.watchers.delete(key);
            }
        }
    }};

    this.watchers.get(key)!.add(disposable);
    return disposable;
}
```

## File Stat

Return file metadata:

```typescript
interface FileStat {
  // File or directory
  type: vscode.FileType;

  // Creation time (milliseconds since epoch)
  ctime: number;

  // Modification time
  mtime: number;

  // File size in bytes
  size: number;

  // Optional permissions
  permissions?: vscode.FilePermission;
}
```

File permissions:

- `Readonly`: File cannot be modified

## Best Practices

**Handle errors properly**: Use `FileSystemError` for standard error cases.

**Fire change events**: Notify VSCode when files change externally.

**Implement watch**: Allow VSCode to watch for file changes.

**Support cancellation**: Handle cancellation tokens for long operations.

**Consider performance**: Cache file metadata when appropriate.

**Test thoroughly**: Test all file operations including edge cases.

## Common Patterns

### Memory File System

```typescript
class MemoryFileSystem implements vscode.FileSystemProvider {
  private root = new Directory();
  private emitter = new vscode.EventEmitter<vscode.FileChangeEvent[]>();
  onDidChangeFile = this.emitter.event;

  // ... implementation
}

class Directory {
  entries: Map<string, Directory | File> = new Map();
}

class File {
  constructor(
    public content: Uint8Array,
    public mtime: number,
  ) {}
}
```

### Remote File System

```typescript
class RemoteFileSystemProvider implements vscode.FileSystemProvider {
  constructor(private client: RemoteClient) {}

  async stat(uri: vscode.Uri): Promise<vscode.FileStat> {
    return this.client.stat(uri.path);
  }

  async readFile(uri: vscode.Uri): Promise<Uint8Array> {
    return this.client.readFile(uri.path);
  }

  async writeFile(uri: vscode.Uri, content: Uint8Array): Promise<void> {
    await this.client.writeFile(uri.path, content);
    this.emitter.fire([{ type: vscode.FileChangeType.Changed, uri }]);
  }

  // ... other methods
}
```

### Virtual File System

```typescript
class VirtualFileSystemProvider implements vscode.FileSystemProvider {
  private files: Map<string, VirtualFile> = new Map();

  constructor() {
    // Create virtual files from generated content
    this.generateVirtualFiles();
  }

  private generateVirtualFiles(): void {
    // Generate files from some source
    const content = this.generateContent();
    this.files.set("/generated.ts", {
      content: new TextEncoder().encode(content),
      mtime: Date.now(),
    });
  }

  // ... implementation
}
```

## Next Steps

- [Tasks](tasks.md): Define and run tasks
- [Terminal](terminal.md): Create terminal profiles
- [VSCode File System Documentation](https://code.visualstudio.com/api/extension-guides/virtual-documents)
