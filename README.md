# Port Forward

Port Forward is a simple TCP proxy written in Go that listens on an address and forwards incoming connections to a proxy server.

## Example

To run the proxy listening on port 9090 and forward connections to `10.42.0.1:8080`:

```bash
./portfwd -l :9090 -p 10.42.0.1:8080
```


## TODO

- [ ] Implement support for Proxy Protocol v2
- [ ] Implement graceful server shutdown
- [ ] Improve error handling
- [ ] Add more tests and validations
