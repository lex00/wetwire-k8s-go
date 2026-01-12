# Kiro CLI Integration

Use Kiro CLI with wetwire-k8s for AI-assisted infrastructure design in Kubernetes environments.

## Prerequisites

- Go 1.23+ installed
- Kiro CLI installed ([installation guide](https://kiro.dev/docs/cli/installation/))
- AWS Builder ID or GitHub/Google account (for Kiro authentication)

---

## Step 1: Install wetwire-k8s

### Option A: Using Go (recommended)

```bash
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

### Option B: Pre-built binaries

Download from [GitHub Releases](https://github.com/lex00/wetwire-k8s-go/releases):

```bash
# macOS (Apple Silicon)
curl -LO https://github.com/lex00/wetwire-k8s-go/releases/latest/download/wetwire-k8s-darwin-arm64
chmod +x wetwire-k8s-darwin-arm64
sudo mv wetwire-k8s-darwin-arm64 /usr/local/bin/wetwire-k8s

# macOS (Intel)
curl -LO https://github.com/lex00/wetwire-k8s-go/releases/latest/download/wetwire-k8s-darwin-amd64
chmod +x wetwire-k8s-darwin-amd64
sudo mv wetwire-k8s-darwin-amd64 /usr/local/bin/wetwire-k8s

# Linux (x86-64)
curl -LO https://github.com/lex00/wetwire-k8s-go/releases/latest/download/wetwire-k8s-linux-amd64
chmod +x wetwire-k8s-linux-amd64
sudo mv wetwire-k8s-linux-amd64 /usr/local/bin/wetwire-k8s
```

### Verify installation

```bash
wetwire-k8s --version
```

---

## Step 2: Install Kiro CLI

```bash
# Install Kiro CLI
curl -fsSL https://cli.kiro.dev/install | bash

# Verify installation
kiro-cli --version

# Sign in (opens browser)
kiro-cli login
```

---

## Step 3: Configure Kiro for wetwire-k8s

Run the design command with `--provider kiro` to auto-configure:

```bash
# Create a project directory
mkdir my-k8s-infra && cd my-k8s-infra

# Initialize Go module
go mod init my-k8s-infra

# Run design with Kiro provider (auto-installs configs on first run)
wetwire-k8s design --provider kiro "Create a deployment for nginx with 3 replicas"
```

This automatically installs:

| File | Purpose |
|------|---------|
| `~/.kiro/agents/wetwire-k8s-runner.json` | Kiro agent configuration |
| `.kiro/mcp.json` | Project MCP server configuration |

### Manual configuration (optional)

The MCP server is provided as a subcommand `wetwire-k8s mcp`. If you prefer to configure manually:

**~/.kiro/agents/wetwire-k8s-runner.json:**
```json
{
  "name": "wetwire-k8s-runner",
  "description": "Infrastructure code generator using wetwire-k8s",
  "prompt": "You are an infrastructure design assistant...",
  "model": "claude-sonnet-4",
  "mcpServers": {
    "wetwire": {
      "command": "wetwire-k8s",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  },
  "tools": ["*"]
}
```

**.kiro/mcp.json:**
```json
{
  "mcpServers": {
    "wetwire": {
      "command": "wetwire-k8s",
      "args": ["mcp"],
      "cwd": "/path/to/your/project"
    }
  }
}
```

> **Note:** The `cwd` field ensures MCP tools resolve paths correctly in your project directory. When using `wetwire-k8s design --provider kiro`, this is configured automatically.

---

## Step 4: Run Kiro with wetwire design

### Using the wetwire-k8s CLI

```bash
# Start Kiro design session
wetwire-k8s design --provider kiro "Create a deployment with nginx and expose it via service"
```

This launches Kiro CLI with the wetwire-k8s-runner agent and your prompt.

### Using Kiro CLI directly

```bash
# Start chat with wetwire-k8s-runner agent
kiro-cli chat --agent wetwire-k8s-runner

# Or with an initial prompt
kiro-cli chat --agent wetwire-k8s-runner "Create a StatefulSet with persistent storage"
```

---

## Available MCP Tools

The wetwire-k8s MCP server exposes three tools to Kiro:

| Tool | Description | Example |
|------|-------------|---------|
| `wetwire_init` | Initialize a new project | `wetwire_init(path="./myapp")` |
| `wetwire_lint` | Lint code for issues | `wetwire_lint(path="./k8s/...")` |
| `wetwire_build` | Generate Kubernetes YAML manifests | `wetwire_build(path="./k8s/...", format="yaml")` |

---

## Example Session

```
$ wetwire-k8s design --provider kiro "Create a deployment with nginx and expose it via service"

Installed Kiro agent config: ~/.kiro/agents/wetwire-k8s-runner.json
Installed project MCP config: .kiro/mcp.json
Starting Kiro CLI design session...

> I'll help you create a deployment with nginx and expose it via a service.

Let me initialize the project and create the infrastructure code.

[Calling wetwire_init...]
[Calling wetwire_lint...]
[Calling wetwire_build...]

I've created the following files:
- k8s/nginx.go

The infrastructure includes:
- Deployment with 3 nginx replicas
- Service exposing port 80
- Proper label selectors for pod discovery

Would you like me to add any additional configurations?
```

---

## Workflow

The Kiro agent follows this workflow:

1. **Explore** - Understand your requirements
2. **Plan** - Design the Kubernetes architecture
3. **Implement** - Generate Go code using wetwire-k8s patterns
4. **Lint** - Run `wetwire_lint` to check for issues
5. **Build** - Run `wetwire_build` to generate Kubernetes manifests

---

## Deploying Generated Manifests

After Kiro generates your infrastructure code:

```bash
# Build the Kubernetes manifests
wetwire-k8s build ./k8s > manifests.yaml

# Deploy with kubectl
kubectl apply -f manifests.yaml

# Or deploy directly without saving to file
wetwire-k8s build ./k8s | kubectl apply -f -
```

---

## Troubleshooting

### MCP server not found

```
Mcp error: -32002: No such file or directory
```

**Solution:** Ensure `wetwire-k8s` is in your PATH:

```bash
which wetwire-k8s

# If not found, add to PATH or reinstall
go install github.com/lex00/wetwire-k8s-go/cmd/wetwire-k8s@latest
```

### Kiro CLI not found

```
kiro-cli not found in PATH
```

**Solution:** Install Kiro CLI:

```bash
curl -fsSL https://cli.kiro.dev/install | bash
```

### Authentication issues

```
Error: Not authenticated
```

**Solution:** Sign in to Kiro:

```bash
kiro-cli login
```

---

## Known Limitations

### Automated Testing

When using `wetwire-k8s test --provider kiro`, tests run in non-interactive mode (`--no-interactive`). This means:

- The agent runs autonomously without waiting for user input
- Persona simulation is limited - all personas behave similarly
- The agent won't ask clarifying questions

For true persona simulation with multi-turn conversations, use the Anthropic provider:

```bash
wetwire-k8s test --provider anthropic --persona expert "Create a StatefulSet with PVCs"
```

### Interactive Design Mode

Interactive design mode (`wetwire-k8s design --provider kiro`) works fully as expected:

- Real-time conversation with the agent
- Agent can ask clarifying questions
- Lint loop executes as specified in the agent prompt

---

## See Also

- [CLI Reference](CLI.md) - Full wetwire-k8s CLI documentation
- [Quick Start](QUICK_START.md) - Getting started with wetwire-k8s
- [Kiro CLI Installation](https://kiro.dev/docs/cli/installation/) - Official installation guide
- [Kiro CLI Docs](https://kiro.dev/docs/cli/) - Official Kiro documentation
