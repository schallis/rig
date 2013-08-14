package main

type Runnable interface {
	Start() error
	Stop() error
}

