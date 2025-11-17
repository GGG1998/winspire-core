# MCP Asana Configuration

This document describes how to configure MCP (Model Context Protocol) Asana server in Cursor IDE.

## Setup Instructions

### Option 1: Using the official MCP Asana server

1. **Get your Asana Personal Access Token:**
   - Go to [Asana Developer Settings](https://app.asana.com/0/developer-console)
   - Create a new Personal Access Token
   - Copy the token

2. **Add to Cursor Settings:**
   - Open Cursor Settings (Cmd/Ctrl + ,)
   - Search for "MCP" or "Model Context Protocol"
   - Add the following configuration to your `mcp.json` or settings:

```json
{
  "mcpServers": {
    "asana": {
      "command": "npx",
      "args": [
        "-y",
        "@modelcontextprotocol/server-asana"
      ],
      "env": {
        "ASANA_PERSONAL_ACCESS_TOKEN": "YOUR_TOKEN_HERE"
      }
    }
  }
}
```

### Option 2: Using alternative MCP Asana implementations

#### MCP Asana by wwwaldo (GitHub)

```json
{
  "mcpServers": {
    "asana": {
      "command": "node",
      "args": [
        "/path/to/mcp-asana/dist/index.js"
      ],
      "env": {
        "ASANA_ACCESS_TOKEN": "YOUR_TOKEN_HERE"
      }
    }
  }
}
```

#### Asana MCP Server by CData

```json
{
  "mcpServers": {
    "asana-cdata": {
      "command": "java",
      "args": [
        "-jar",
        "/path/to/asana-mcp-server.jar"
      ],
      "env": {
        "ASANA_CONNECTION_STRING": "YOUR_CONNECTION_STRING"
      }
    }
  }
}
```

## Environment Variables

Set your Asana token as an environment variable:

```bash
export ASANA_PERSONAL_ACCESS_TOKEN="your_token_here"
```

Or add it to your shell profile (`~/.zshrc`, `~/.bashrc`, etc.)

## Verification

After configuration, restart Cursor and verify the MCP server is connected:
- Check Cursor's MCP status in settings
- Try using Asana-related commands in chat

## Security Notes

- Never commit your Asana token to version control
- Use environment variables or secure credential storage
- Rotate tokens regularly
- Use the minimum required permissions

## References

- [MCP Asana Server](https://mcp.so/server/mcp-asana/wwwaldo)
- [Asana API Documentation](https://developers.asana.com/docs)
- [Cursor MCP Documentation](https://docs.cursor.com/mcp)

