# Matrix GroupMe Go Bridge
A Matrix-GroupMe puppeting bridge

[Features & Roadmap](./ROADMAP.md)

## Discussion
Matrix room: [#groupme-go-bridge:malhotra.cc](https://matrix.to/#/#groupme-go-bridge:malhotra.cc)

## Credits


## Build Instructions

### Windows

To build on Windows (or systems without libolm), use the `nocrypto` tag:

```powershell
go build -tags nocrypto .
```

This will produce `groupme.exe`.

### Linux/macOS

If you have libolm installed:

```bash
go build .
```
