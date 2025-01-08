# Product Context

## Purpose
This project is a Go-based console application that sends notifications to ntfy.sh, a pub-sub notification service.

## Problem Statement
The application solves the need for simple, programmatic notification delivery by making HTTP calls to ntfy.sh on startup.

## Expected Behavior
1. Application starts
2. Makes an HTTP POST request to ntfy.sh/dapidi_alerts with the message "Hi"
3. Exits after successful notification delivery or error handling