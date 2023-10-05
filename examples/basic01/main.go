package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/mrsoftware/circuitbreaker"
)

var cb circuitbreaker.Manager

func main() {
	cb = circuitbreaker.NewCircuit(circuitbreaker.WithDefaultOptions())

	res, err := Get(context.Background(), "https://google.com")
	if err != nil {
		log.Println(err)

		return
	}

	fmt.Println("response: ", string(res))
}

func Get(ctx context.Context, url string) (res []byte, err error) {
	// first you have to check the circuit state.
	if !cb.IsAvailable(ctx) {
		return nil, circuitbreaker.ErrIsOpen // or any error you like
	}

	// after the proccess is done, we need to notify the circuit breaker the the result.
	defer func() { cb.Done(ctx, err) }()

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != 200 {
		return nil, errors.New("response status code is not 200")
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return body, nil
}
