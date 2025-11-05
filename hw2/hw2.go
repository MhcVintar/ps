package main

import (
	"fmt"
	"math/rand"
	"time"
)

type Meritev struct {
	vrsta    string
	vrednost float32
}

func zazeniSenzor(vrsta string, min, max float32, meritve chan<- Meritev) {
	for {
		meritve <- Meritev{
			vrsta:    vrsta,
			vrednost: min + rand.Float32()*(max-min),
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func main() {
	meritve := make(chan Meritev, 3)

	senzorji := []struct {
		vrsta string
		min   float32
		max   float32
	}{
		{vrsta: "temperatura", min: -10, max: 35},
		{vrsta: "vlaga", min: 0, max: 100},
		{vrsta: "tlak", min: 950, max: 1050},
	}

	for _, senzor := range senzorji {
		go zazeniSenzor(senzor.vrsta, senzor.min, senzor.max, meritve)
	}

	done := make(chan struct{})
	go func() {
		defer close(done)
		fmt.Scanln()
	}()

	for {
		select {
		case meritev := <-meritve:
			fmt.Printf("%v: %.2f\n", meritev.vrsta, meritev.vrednost)
		case <-time.After(5 * time.Second):
			fmt.Println("sistem je neodziven")
			return
		case <-done:
			return
		}
	}
}
