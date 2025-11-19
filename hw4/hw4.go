package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type narocilo interface {
	obdelaj()
}

var (
	promet    = 0.0
	stNarocil = 0
	obdelava  sync.Mutex
)

type izdelek struct {
	imeIzdelka string
	cena       float64
	teza       float64
}

var _ narocilo = (*izdelek)(nil)

func (i *izdelek) obdelaj() {
	obdelava.Lock()
	defer obdelava.Unlock()
	promet += i.cena
	stNarocil++

	fmt.Println("Številka naročila:", stNarocil)
	fmt.Println("Ime izdelka:", i.imeIzdelka)
	fmt.Println("Cena:", i.cena, "€")
	fmt.Println("Teža:", i.teza, "kg")
	fmt.Println("---")
}

type eknjiga struct {
	naslovKnjige string
	cena         float64
}

var _ narocilo = (*eknjiga)(nil)

func (e *eknjiga) obdelaj() {
	obdelava.Lock()
	defer obdelava.Unlock()
	promet += e.cena
	stNarocil++

	fmt.Println("Številka naročila:", stNarocil)
	fmt.Println("Naslov knjige:", e.naslovKnjige)
	fmt.Println("Cena:", e.cena, "€")
	fmt.Println("---")
}

type spletniTecaj struct {
	imeTecaja   string
	trajanjeUre int
	cenaUre     float64
}

var _ narocilo = (*spletniTecaj)(nil)

func (s *spletniTecaj) obdelaj() {
	obdelava.Lock()
	defer obdelava.Unlock()
	promet += s.cenaUre
	stNarocil++

	fmt.Println("Številka naročila:", stNarocil)
	fmt.Println("Ime tecaja:", s.imeTecaja)
	fmt.Println("Trajanje ure:", s.trajanjeUre)
	fmt.Println("Cena ure:", s.cenaUre, "€")
	fmt.Println("---")
}

func proizvajalec(narocila chan<- narocilo, konec <-chan struct{}) {
	moznaNarocila := []narocilo{
		&izdelek{imeIzdelka: "Prenosnik", cena: 899.99, teza: 1.5},
		&eknjiga{naslovKnjige: "Go programiranje", cena: 19.99},
		&spletniTecaj{imeTecaja: "Uvod v Go", trajanjeUre: 10, cenaUre: 49.99},
		&izdelek{imeIzdelka: "Miška", cena: 25.50, teza: 0.2},
		&eknjiga{naslovKnjige: "Algoritmi in podatkovne strukture", cena: 29.99},
		&spletniTecaj{imeTecaja: "Napredni Go", trajanjeUre: 15, cenaUre: 79.99},
		&izdelek{imeIzdelka: "Tipkovnica", cena: 65.00, teza: 0.8},
		&eknjiga{naslovKnjige: "Umelna inteligenca 101", cena: 34.99},
		&spletniTecaj{imeTecaja: "Full-Stack Web Development", trajanjeUre: 20, cenaUre: 99.99},
		&izdelek{imeIzdelka: `Monitor 27"`, cena: 189.90, teza: 3.2},
	}

	for {
		time.Sleep(500 * time.Millisecond)
		select {
		case <-konec:
			return
		default:
			narocila <- moznaNarocila[rand.Intn(len(moznaNarocila))]
		}
	}
}

func porabnik(narocila <-chan narocilo) {
	for narocilo := range narocila {
		time.Sleep(500 * time.Millisecond)
		narocilo.obdelaj()
	}
}

func main() {
	konec := make(chan struct{})

	narocila := make(chan narocilo, 5)

	var proizvajalci sync.WaitGroup
	proizvajalci.Add(5)
	for range 5 {
		go func() {
			proizvajalec(narocila, konec)
			proizvajalci.Done()
		}()
	}

	go func() {
		fmt.Scanln()
		close(konec)
		proizvajalci.Wait()
		close(narocila)
	}()

	var porabniki sync.WaitGroup
	porabniki.Add(2)
	for range 2 {
		go func() {
			porabnik(narocila)
			porabniki.Done()
		}()
	}

	porabniki.Wait()

	fmt.Printf("Skupni znesek: %.2f\n €", promet)
	fmt.Println("Število naročil:", stNarocil)
}
