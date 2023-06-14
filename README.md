<div align="center">
	<h1>Zactem</h1>
	<h4 align="center">
        A Yaml File Template Engine
	</h4>
</div>

<p align="center">
	<a href="#-installation-and-documentation">Installation</a> ❘
	<a href="#-features">Features</a> ❘
	<a href="#-usage">Usage</a>
</p>

## 🚀&nbsp; Installation and Documentation

```bash
go get github.com/zaCloud-net/zactem
```

## Features
- template engine for yaml files
- import files
- parse nested maps

## Example

Docker compose yml file example

```yml
{{import * as LoggingTemplate from "/templates/logging.yml"}}
version: "3.9"
services:
    {{service_name}}:
        image: wordpress:latest
        container_name: {{container_name}}
        port: {{port}}
        logging:
            {{LoggingTemplate.logging}}
```

`logging.yml` template
```yml
logging:
      driver: "json-file"
      options:
        max-size: "1g"
        max-file: "2"
```

### Render the template

```go

```