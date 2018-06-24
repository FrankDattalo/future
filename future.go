// Copyright Frank Dattalo 2018

package future

// Future represents any value which will be
// available at a later time.
type Future interface {

	// Awaits for the Future to complete,
	// returning the value and any error that occurred.
	// Await will panic if the reciever is nil.
	Await() (interface{}, error)
}

// NewFuture concurrently runs the function f and returns
// a Future encapsulating its retult. NewFuture will panic if f
// is nil.
func NewFuture(f func() (interface{}, error)) Future {

	if f == nil {
		panic("f was nil!")
	}

	ret := new(futureImpl)

	// this channel implements a turnstile/monitor mechanism for
	// Await calls. It ensures that only one gorutine
	// will ever be live within the method. It is all that is used to
	// ensure no race condition happens between readers and writer
	ret.turnstile = make(chan bool, 1)

	// setting the turnstile to false is crutial
	// as it prevents any thread from terminating
	// before allowed
	ret.turnstile <- false

	go func() {
		data, err := f()

		ret.data = data
		ret.err = err

		// signal to the one waiting thread that
		// we are done
		ret.turnstile <- true
	}()

	return ret
}

// Join joins the two Futures by first waiting on f1
// then waiting on f2, and finally applying the returned
// results to the reduction function parameter c. Join will panic
// if f1, f2, or c are nil.
func Join(f1, f2 Future, c func(interface{}, error, interface{}, error) (interface{}, error)) Future {

	if f1 == nil {
		panic("f1 was nil!")
	}

	if f2 == nil {
		panic("f2 was nil!")
	}

	if c == nil {
		panic("c was nil!")
	}

	return NewFuture(func() (interface{}, error) {
		data1, err1 := f1.Await()
		data2, err2 := f2.Await()

		return c(data1, err1, data2, err2)
	})
}

// Then applys the function c to the result of f1 and returns
// a Future of that result. Then will panic if f1 or c are nil.
func Then(f1 Future, c func(interface{}, error) (interface{}, error)) Future {

	if f1 == nil {
		panic("f1 was nil!")
	}

	if c == nil {
		panic("c was nil!")
	}

	return NewFuture(func() (interface{}, error) {

		data, err := f1.Await()

		return c(data, err)
	})
}

// JoinAll joins all given Futures by waiting for each Future
// then applying the results of each to the reduction
// parameter function c. JoinAll will panic if fs, fs' element, or c
// are nil.
func JoinAll(fs []Future, c func([]interface{}, []error) (interface{}, error)) Future {

	if fs == nil {
		panic("f was nil!")
	}

	if c == nil {
		panic("c was nil!")
	}

	return NewFuture(func() (interface{}, error) {

		datas := make([]interface{}, len(fs))
		errs := make([]error, len(fs))

		for i, f := range fs {
			di, ei := f.Await()
			datas[i] = di
			errs[i] = ei
		}

		return c(datas, errs)
	})
}

// private implementation of the future interface
type futureImpl struct {

	// data result from function call
	data interface{}

	// error result from function call
	err error

	// reference to the turnstile that this
	// future uses to prevent race conditions
	turnstile chan bool
}

// implementation of Await for the future interface
func (f *futureImpl) Await() (interface{}, error) {

	if f == nil {
		panic("Reciever was nil!")
	}

	// ensures that only one thread is in
	// the rest of the await function at any time
	isDone := <-f.turnstile

	// will pass the value that we previously
	// on the turnstile to the next thread
	defer func() {
		f.turnstile <- isDone
	}()

	// this if statement will only be triggered
	// after a true is recieved from the turnstile
	if isDone {
		return f.data, f.err
	}

	// waits until we get the done signal
	// from the writer thread. now all
	// pulls from the turnstile will be true; triggering
	// the above if statement for all calls to Await going
	// forward
	isDone = <-f.turnstile

	return f.data, f.err
}
