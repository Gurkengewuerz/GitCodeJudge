# System Architecture

```
                   ┌─────────────┐
                   │   Gitea     │
                   │  (Git Host) │
                   └─────┬───────┘
                         │ webhook
                         ▼
┌──────────────┐   ┌─────────────┐   ┌─────────────┐
│  Test Cases  │   │   Judge     │   │   Docker    │
│  Repository  │──▶│   Server    │──▶│  Containers │
└──────────────┘   └─────────────┘   └─────────────┘
                         │
                         │ results
                         ▼
                   ┌─────────────┐
                   │  Commit     │
                   │  Comments   │
                   └─────────────┘
```

## Overview

The GitCodeJudge system consists of multiple components working together to provide a secure and scalable automated
testing environment. In this example students uses ContainerSSH to connect to a development container to work on their code.
When they are ready to submit their code, they push it to a Git server. The Git server instance triggers a webhook to the Judge server.

## Components Diagram

```mermaid
flowchart TD
    Student([Student])
    Git[Git Host]
    Judge[Judge Server]
    TestCases[Test Cases\nRepository]
    Docker[Docker\nContainers]
    Comments[Commit\nStatus]
    SSH[ContainerSSH]
    
    %% Student interactions
    Student -->|git push| Git
    Student -->|ssh connect| SSH
    
    %% Main judge flow
    Git -->|webhook| Judge
    TestCases -->|load| Judge
    Judge -->|execute| Docker
    Judge -->|write| Comments
    
    %% ContainerSSH flow
    SSH -->|create| Docker
    SSH -->|authenticate| Git
    
    classDef system fill:#f9f,stroke:#333,stroke-width:2px;
    classDef storage fill:#bbf,stroke:#333,stroke-width:2px;
    classDef user fill:#dfd,stroke:#333,stroke-width:2px;
    
    class Git,Judge,SSH system;
    class TestCases,Comments storage;
    class Student user;
    class Docker system;
```
