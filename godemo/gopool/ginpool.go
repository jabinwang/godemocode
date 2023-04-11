package gopool

import "context"

var defaultPool Pool

func init() {
	defaultPool = newPool()
}

func Go(f func()) {
	defaultPool.Go(f)
}

func SetPanicHandler(f func(context.Context, interface{}))  {
	defaultPool.SetPanicHandler(f)
}