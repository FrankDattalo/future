# Future

A small future implmentation for go!

## Introduction

TODO

## Example

```go
package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/frankdattalo/future"
)

func listFn() (interface{}, error) {
	return []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, nil
}

func filterFn() (interface{}, error) {
	return map[int]bool{3: true, 4: true, 5: true, 7: true}, nil
}

func namesFn() (interface{}, error) {
	return map[int]string{
		1:  "One",
		2:  "Tooo",
		3:  "Three",
		4:  "Floor",
		5:  "Fliveee",
		6:  "Shicks",
		7:  "Szeven",
		8:  "Ate",
		9:  "Nyne",
		10: "Tan",
	}, nil
}

func makeSleeper(
	name string,
	f func() (interface{}, error)) func() (interface{}, error) {

	return func() (interface{}, error) {
		sleepTime := time.Second * time.Duration(rand.Intn(10))
		fmt.Printf("%v sleeping for %v\n", name, sleepTime)
		defer func() {
			fmt.Printf("%v done!\n", name)
		}()
		time.Sleep(sleepTime)
		return f()
	}
}

func castException(name, to string) error {
	return &castExceptionImpl{name: name, to: to}
}

type castExceptionImpl struct {
	name, to string
}

func (c *castExceptionImpl) Error() string {
	return fmt.Sprintf("Could not cast %v to %v", c.name, c.to)
}

func applyFilterFn(
	listData interface{},
	listError error,
	filerData interface{},
	filterErr error) (interface{}, error) {

	list := listData.([]int)

	filter := filerData.(map[int]bool)

	ret := make([]int, 0)

	for _, val := range list {
		_, contains := filter[val]
		if contains {
			ret = append(ret, val)
		}
	}

	return ret, nil
}

func nameFn(
	filterData interface{},
	filterErr error,
	namesData interface{},
	namesErr error) (interface{}, error) {

	filtered := filterData.([]int)
	names := namesData.(map[int]string)

	ret := make([]string, 0)

	for _, v := range filtered {
		ret = append(ret, names[v])
	}

	return ret, nil
}

func exclaimFn(data interface{}, err error) (interface{}, error) {
	if err != nil {
		return nil, err
	}

	casted := data.([]string)

	ret := make([]string, len(casted))

	for i, v := range casted {
		ret[i] = v + "!!!"
	}

	return ret, nil
}

func main() {

	list := future.NewFuture(makeSleeper("listFn", listFn))

	filter := future.NewFuture(makeSleeper("filterFn", filterFn))

	names := future.NewFuture(makeSleeper("namesFn", namesFn))

	filtered := future.Join(list, filter, applyFilterFn)

	named := future.Join(filtered, names, nameFn)

	exclaimed := future.Then(named, exclaimFn)

	data, err := exclaimed.Await()

	if err != nil {
		panic(err)
	}

	fmt.Println(data)
}

//=> namesFn sleeping for 7s
//=> listFn sleeping for 7s
//=> filterFn sleeping for 1s
//=> filterFn done!
//=> namesFn done!
//=> listFn done!
//=> [Three!!! Floor!!! Fliveee!!! Szeven!!!]
```
