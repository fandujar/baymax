# Baymax

Baymax is a Slack bot powered by OpenAI's GPT models, designed to assist users within Slack channels. It listens to messages where it's mentioned, processes them using OpenAI's Chat Completion API, and responds accordingly. Baymax supports plugin extensions, allowing for custom tool integrations.

## Features

- **Slack Integration**: Listens and responds to messages in Slack channels.
- **OpenAI GPT Integration**: Uses OpenAI's GPT models for generating responses.
- **Plugin Support**: Load custom plugins to extend functionality.
- **NATS Messaging**: Utilizes NATS for internal messaging between services.
- **Health Checks**: Provides liveness and readiness endpoints.
- **Docker and Helm Support**: Easily deployable using Docker and Kubernetes Helm charts.

## Table of Contents

- [Architecture Overview](#architecture-overview)
- [Installation](#installation)
  - [Prerequisites](#prerequisites)
  - [Clone the Repository](#clone-the-repository)
  - [Build from Source](#build-from-source)
  - [Using Docker](#using-docker)
- [Configuration](#configuration)
- [Usage](#usage)
  - [Running Locally](#running-locally)
  - [Running with Docker](#running-with-docker)
  - [Deploying to Kubernetes with Helm](#deploying-to-kubernetes-with-helm)
- [Plugin Development](#plugin-development)
  - [Plugin Interface](#plugin-interface)
- [Health Checks](#health-checks)
- [Logging](#logging)
- [Contributing](#contributing)
- [License](#license)
- [Contact](#contact)

## Architecture Overview

Baymax consists of the following components:

- **`cmd/baymax`**: The main application entry point.
- **`pkg/providers`**: Providers for Slack, OpenAI, and NATS services.
- **`pkg/services`**: Business logic for handling Slack and OpenAI interactions.
- **`pkg/transport`**: Event handlers and transport mechanisms.
- **`pkg/plugins`**: Plugin system for extending Baymax functionalities.
- **`chart/baymax`**: Helm chart for deploying Baymax on Kubernetes.

## Installation

### Prerequisites

- **Go**: Version 1.22 or later.
- **Docker**: For containerization (optional).
- **Kubernetes**: Cluster for deployment (optional).
- **Slack App**: With proper tokens.
- **OpenAI API Key**: For accessing OpenAI services.
- **NATS Server**: Optional, Baymax can start its own embedded server.

### Clone the Repository

```bash
git clone https://github.com/fandujar/baymax.git
cd baymax
```

#### Build from Source
```bash
go build -o baymax ./cmd/baymax
```

#### Using Docker
Build the Docker image:

```bash
docker build -t baymax:latest -f docker/Dockerfile .
```

### Configuration
Baymax requires the following environment variables:

- `SLACK_APP_TOKEN`: Slack App-level token (starts with xapp-).
- `SLACK_BOT_TOKEN`: Slack Bot token (starts with xoxb-).
- `OPENAI_API_KEY`: OpenAI API key.

Optional environment variables:
- `LOG_LEVEL`: Defines de log level for the application.
- `BAYMAX_NAME`: The name Baymax will introduce itself as (default: "Baymax").
- `BAYMAX_PLUGINS_DIR`: Directory to load plugins from (default: current directory).
- `OPENAI_SYSTEM_MESSAGE`: Defines a custom system role message.
- `OPENAI_MODEL`: Defines the OpenAI Model to be used. Defaults to gpt-4o-mini.

Set these variables in your environment or pass them when running the application.

### Usage

#### Running Locally
Set the required environment variables and run:

```bash
./baymax
```

#### Running with Docker
```bash
docker run -e SLACK_APP_TOKEN=your_app_token \
           -e SLACK_BOT_TOKEN=your_bot_token \
           -e OPENAI_API_KEY=your_openai_key \
           baymax:latest
```

### Deploying to Kubernetes with Helm
First, create a Kubernetes secret with your tokens:

```bash
kubectl create secret generic baymax-secrets \
  --from-literal=slack-app-token=your_app_token \
  --from-literal=slack-bot-token=your_bot_token \
  --from-literal=openai-api-key=your_openai_key
```

#### Install the Helm chart:

```bash
helm install baymax ./chart/baymax
```

### Plugin Development
Baymax supports plugins to extend its functionality. Plugins should implement the Plugin interface defined in pkg/plugins/plugins.go. Compile your plugin as a Go plugin (.so file) and place it in the directory specified by `BAYMAX_PLUGINS_DIR`.

#### Plugin Interface
```go
type Plugin interface {
    GetTools() []openai.Tool
    RunTool(toolName string, parameters string, messages []openai.ChatCompletionMessage, tools []openai.Tool) (string, error)
    RunEventLoop(natsClient *nats.Conn)
}
```

### Health Checks
Baymax exposes health check endpoints:

- Liveness Probe: /liveness on port 8081.
- Readiness Probe: /readiness on port 8081.

### Logging
Baymax uses zerolog for structured logging. The log level is set to Info by default.

## Contributing
Contributions are welcome! Please open an issue or submit a pull request on GitHub.

## License
This project is licensed under the MIT License - see the LICENSE file for details.

## Contact
For questions or support, please open an issue on the GitHub repository.