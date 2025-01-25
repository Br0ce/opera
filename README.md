# Opera: A Tool-Discovering Agent Framework  

![Go Version](https://img.shields.io/badge/Go-1.20-blue)  
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/Br0ce/opera)](https://github.com/Br0ce/opera)
![License](https://img.shields.io/badge/license-MIT-green)  

Opera is a lightweight, extensible agent framework written in Go, designed to enable seamless discovery and utilization of tools available in a Docker network. With Opera, developers can build intelligent agents that can dynamically identify and interact with other services or tools within the same network, streamlining workflows and fostering better service orchestration.  

## 🚀 Features  

- **Dynamic Tool Discovery**: Automatically detects tools and services available in the Docker network.  
- **Observability with OpenTelemetry**: Built-in support for distributed tracing using [OpenTelemetry](https://opentelemetry.io/), providing visibility into tool interactions and network activity.  
- **Go-Powered Performance**: Written in Go for high performance and scalability.  
- **Plug-and-Play Design**: Easily integrate with new tools and services without significant configuration overhead.  
- **Extensibility**: Add custom agents and behaviors tailored to your specific use cases.  
- **Lightweight**: Focused on efficiency, Opera ensures minimal overhead for tool communication and discovery.  
- **Docker-Native**: Specifically designed to work seamlessly within Dockerized environments.  

## 🎯 Motivation  

Modern distributed systems and microservice architectures often involve numerous tools and services. However, managing these tools and enabling communication between them can be a challenge.  

Observability is critical in such scenarios to understand system behavior, diagnose issues, and optimize performance. Opera leverages OpenTelemetry to provide detailed, end-to-end visibility into the interactions between agents and tools within the Docker network.  

Opera was born out of the need for an agent framework that goes beyond static configurations and manual integrations. By enabling dynamic discovery of tools and incorporating observability, Opera simplifies service orchestration and empowers developers to focus on building functionality rather than managing infrastructure.  

## 🛠️ Getting Started  

Follow these steps to get started with Opera:  

### Prerequisites  
- [Go](https://golang.org/) (version 1.20 or newer).  
- [Docker](https://www.docker.com/) installed on your machine.  
- Basic knowledge of Docker networking and services.  

### Installation  
1. Clone the Opera repository:  
   ```bash  
   git clone https://github.com/Br0ce/opera.git  
   cd opera  

### License
Opera is licensed under the MIT License.