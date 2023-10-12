# Circuit Breaker

The Circuit Breaker pattern, can prevent an application from repeatedly trying to execute an operation that's likely to fail. Allowing it to continue without waiting for the fault to be fixed or wasting CPU cycles while it determines that the fault is long lasting. The Circuit Breaker pattern also enables an application to detect whether the fault has been resolved. If the problem appears to have been fixed, the application can try to invoke the operation.

In this package we provide you an implementation of [Circuit Breaker pattern](https://learn.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker).
 
## Installation

```bash
go get github.com/mrsoftware/circuitbreaker
```
    
## Usage
consider each instance as a circuit, and it's a state matchine. you can create a circuit using `NewCircuit`.

we only support `redis` and `memory` storage.


for code documentation you can use [Go Doc](https://pkg.go.dev/github.com/mrsoftware/circuitbreaker).


## Examples

### using `IsAvailable` and `Done` methods:
```Go
cb := circuitbreaker.NewCircuit(circuitbreaker.WithDefaultOptions())

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
```

### using `Do` method:

```Go
cb := circuitbreaker.NewCircuit(circuitbreaker.WithDefaultOptions())

func Get(ctx context.Context, url string) (res []byte, err error) {
	// no need to check the circuit breaker state and report the result to it, `Do` will do them for you.
	response, err := cb.Do(ctx, func() (interface{}, error) {
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

	})

	if err != nil {
		return nil, err
	}

	return response.([]byte), nil
}
```
### Stat
you can get circuit stat by calling `Stat` method on `Circuit` like below:

```Go
cb := circuitbreaker.NewCircuit(circuitbreaker.WithDefaultOptions())

stat := cb.Stat(context.Background())

```


## Roadmap
- [ ]  Health checker option
- [x]  Stat (State, failures, success)

