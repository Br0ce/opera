# Opera: A Tool-Discovering Multi-Agent Agent Framework

![Build Status](https://github.com/Br0ce/opera/actions/workflows/ci.yml/badge.svg)
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/Br0ce/opera)](https://github.com/Br0ce/opera)
[![Go Reference](https://pkg.go.dev/badge/github.com/Br0ce/opera.svg)](https://pkg.go.dev/github.com/Br0ce/opera)
[![Go Report Card](https://goreportcard.com/badge/github.com/Br0ce/opera)](https://goreportcard.com/report/github.com/Br0ce/opera)
![License](https://img.shields.io/badge/license-MIT-green)

**Note**: This project is evolving towards a multi-agent framework, leveraging the Raft algorithm.

Opera is a lightweight, extensible multi-agent framework written in Go, designed to enable the creation of intelligent agent meshes. Leveraging the Raft consensus algorithm, Opera forms a decentralized network where agents can coordinate and collaborate to solve complex tasks. Additionally agents can dynamically identify and interact with other services or tools within the same network, streamlining workflows and fostering better service orchestration. 

## 🚀 Features  

- **Dynamic Multi-Agent Configuration:** Adapts to given tasks by dynamically selecting and configuring the most appropriate multi-agent setup.
- **Dynamic Tool Discovery**: Automatically detects tools and services available in the Docker network.  
- **Observability with OpenTelemetry**: Built-in support for distributed tracing using [OpenTelemetry](https://opentelemetry.io/), providing visibility into tool interactions and network activity.  
- **Go-Powered Performance**: Written in Go for high performance and scalability.  
- **Plug-and-Play Design**: Easily integrate with new tools and services without significant configuration overhead.  
- **Extensibility**: Add custom agents and behaviors tailored to your specific use cases.  
- **Lightweight**: Focused on efficiency, Opera ensures minimal overhead for tool communication and discovery.  
- **Docker-Native**: Specifically designed to work seamlessly within Dockerized environments.  

## 🎯 Motivation  

Modern distributed systems and complex tasks often require coordinated efforts from multiple agents. Opera aims to provide a robust framework that simplifies the creation and management of such multi-agent systems.

The move to a multi-agent framework with Raft consensus addresses the need for reliable, scalable, and adaptable agent networks. By dynamically configuring agent setups based on tasks, Opera enables efficient and flexible task resolution.

Observability is critical in such scenarios to understand system behavior, diagnose issues, and optimize performance. Opera leverages OpenTelemetry to provide detailed, end-to-end visibility into the interactions between agents and tools within the Docker network.  

Opera was born out of the need for an agent framework that goes beyond static configurations and manual integrations. By enabling dynamic discovery of tools and incorporating observability, Opera simplifies service orchestration and empowers developers to focus on building functionality rather than managing infrastructure.  

## 🤝 The Raft Consensus

Opera incorporates the Raft consensus algorithm to build a robust and reliable multi-agent system. This choice provides several key benefits: 

- automatic leader election ensures high availability, preventing single points of failure
- state replication guarantees consistency across the agent network, enhancing data integrity
- fault tolerance enables the system to gracefully handle agent failures, maintaining continuous operation.

These features are crucial for complex multi-agent tasks, where reliability and coordination are paramount. For a deeper understanding of the Raft algorithm, refer to "In Search of an Understandable Consensus Algorithm" (https://raft.github.io/raft.pdf) by Diego Ongaro and John Ousterhout.

**Note: The multi-agent capabilities are under active development. Expect frequent updates and improvements.**


## 🛠️ Development

- [JaegerUI](http://localhost:16686)

### Prerequisites  
- [Go](https://golang.org/) (version 1.23 or newer).  
- [Docker](https://www.docker.com/) installed on your machine.  

### License
Opera is licensed under the MIT License.