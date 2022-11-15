## AUR metadata processor

This is a simple library for processing AUR metadata.

It retrieves the metadata from the AUR, parses it and allows you to query it.

### Usage

```go
package main

func main() {
	aurCache, err := metadata.NewAURCache()
	if err != nil {
        fmt.Println(err)
		return 
	}
}
```