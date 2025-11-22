# Migration Guide

## Migration from v1

If you're upgrading from the old API:

**Old API:**
```go
client := &fetchgo.Client{
    BaseURL: "https://api.example.com",
    RequestType: fetchgo.RequestJSON,
}
client.SendRequest("POST", "/users", data, func(result any, err error) {
    // result was any, needed type assertion
})
```

**New API:**
```go
fg := fetchgo.New()
client := fg.NewClient("https://api.example.com", 5000)
client.SendJSON("POST", "/users", data, func(result []byte, err error) {
    // result is always []byte
})
```
