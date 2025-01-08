# System Patterns

## Architecture
- Simple console application architecture
- Single responsibility: HTTP notification delivery
- Function-based organization with clear separation of concerns

## Technical Decisions
1. Using net/http package for HTTP requests
   - Standard library provides all needed functionality
   - No external dependencies required
   
2. Error handling pattern
   - Immediate error reporting and exit
   - Clear error messages for debugging

## Code Organization
- main.go: Entry point and notify function
- Flat structure due to simple requirements