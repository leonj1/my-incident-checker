# System Patterns

## Architecture
- Console application with concurrent operations
- Multiple responsibilities:
  1. HTTP notification delivery to ntfy.sh
  2. Incident polling and monitoring
  3. Regular heartbeat checks to nosnch.in
- Function-based organization with clear separation of concerns

## Technical Decisions
1. Using net/http package for HTTP requests
   - Standard library provides all needed functionality
   - No external dependencies required
   
2. Error handling pattern
   - Immediate error reporting and exit for critical startup errors
   - Logged errors with continued execution for non-critical operations
   - Clear error messages for debugging

3. Concurrency pattern
   - Goroutines for independent operations
   - Time-based tickers for regular intervals
   - Non-blocking execution of heartbeats

## Code Organization
- main.go: Entry point with multiple concurrent functions:
  - notify: Send notifications to ntfy.sh
  - pollIncidents: Monitor and report incidents
  - sendHeartbeat: Regular health checks to nosnch.in
- Flat structure with modular functions