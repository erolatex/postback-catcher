# Postback Catcher

Postback Catcher is a small and simple Go server for receiving and storing postbacks from various sources. It is designed
to be easy to use and deploy, requiring minimal configuration.

## Features

- Receive postbacks via multiple HTTP methods (GET, POST, PUT, DELETE, etc.)
- Store postbacks in a BoltDB database
- Query stored postbacks with customizable limit
- Delete stored postbacks by ID
- Test URL redirection with custom headers
- Health check endpoint

## Installation

1. Clone the repository:

```bash
git clone https://github.com/erolatex/postback-catcher.git
```

2. Change into the project directory:

```bash
cd postback-catcher
```
3. Build the binary:

```bash
go build -o postback-catcher
```

4. Run the binary:

```bash
./postback-catcher
```
The server will start on port `8081` by default. You can change the port by modifying the `port` constant in the source code.
## Usage

### Send a postback

Make an HTTP request to the server with the desired method and parameters:

```bash
curl -X POST http://localhost:8081/somepath?param1=value1 -d "request body"
```

### Retrieve stored postbacks

Get the stored postbacks with an optional limit:

```bash
curl http://localhost:8081/get?limit=5
```
### Delete a postback

Delete a postback by ID:

```bash
curl -X DELETE http://localhost:8081/delete/your_postback_id_here
```
### Test URL redirection

Redirect to a test URL with custom headers:

```bash
curl http://localhost:8081/test-url?header=Location&value=https://example.com
```

### Health check

Check the server health:

```bash
curl http://localhost:8081/health
```
## Contributing
Please feel free to submit issues, fork the repository and send pull requests!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.