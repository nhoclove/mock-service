# mock-service

# Description
This is a simple mock server that allows quickly creating HTTP services
via definition in JSON files in ./data folder.
# How to
1. Define route in a JSON file in ./data folder
    - Path: URI
    - Method: HTTP methods(post, get ...)
    - Request:
        + Header: Key-value pairs
        + Body: Request body if it is a POST/PUT/PATCH request
    - Response:
        + StatusCode: HTTP standard status code
        + Header: Key-value pairs
        + Body: Response body
2. Start server:

    `# go run main.go`

# Notes:
- No external dependencies
- Only support POST, GET method for now