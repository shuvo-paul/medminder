# MedMinder Project Context

## Project Overview
MedMinder is a medication reminder application built with Go. The project is structured following Go best practices with a clear separation of concerns across different directories. The module name is `github.com/shuvo-paul/medminder`.

## Directory Structure
The project follows a standard Go project layout with the following key directories:

```
/Users/shuvo/Projects/medminder/
├───api/
│   └───postman/
├───cmd/
│   └───server/
├───configs/
├───internal/
│   ├───common/
│   ├───config/
│   ├───features/
│   ├───middleware/
│   ├───router/
│   └───server/
├───migrations/
├───pkg/
├───scripts/
├───sql/
├───tests/
│   ├───e2e/
│   ├───integration/
│   └───testutil/
├───tools/
├───web/
├───.gitignore
├───go.mod
├───QWEN.md
└───README.md
```

## Technologies Used
- Go (version 1.25.0 as per go.mod)

## Building and Running
To build and run the application:
```bash
# Initialize dependencies
go mod tidy

# Run the application (assuming main.go exists in cmd/server)
go run cmd/server/main.go
```

Note: The actual build/run commands may vary depending on the final implementation in the cmd/server directory, which currently appears to be empty.

## Development Conventions
- Follows Go project layout conventions
- Modular structure separating internal code from public packages
- Organized testing structure with unit, integration, and end-to-end tests
- Configuration management in dedicated directory

## Current Status
The project appears to be in early stages of development with most directories currently empty. This suggests it's a newly initialized project that will be populated with features related to medication reminders.

## Key Files
- `go.mod` - Go module definition
- `README.md` - Basic project information
- `.gitignore` - Git ignore rules