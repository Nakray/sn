# sn - VK Social Network Data Collector

Go-based tool for collecting and monitoring VKontakte social network data.

## Features

- VK API data collection (users, groups, posts, photos, etc.)
- Scheduled monitoring tasks
- Proxy support
- Account management
- Web UI for task management
- PostgreSQL storage

## Installation

```bash
git clone https://github.com/Nakray/sn.git
cd sn
go mod download
```

## Configuration

Copy `config.example.json` to `config.json` and configure:
- Database connection
- Server port
- Monitoring intervals
- VK API settings

## Usage

```bash
go run cmd/sn/main.go
```

Web UI will be available at `http://localhost:8080`

## Database

Use the provided schema from `schemas_structure.sql` (from original project)

## License

MIT
