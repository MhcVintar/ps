package main

import "fmt"

type Student struct {
	ime     string
	priimek string
	ocene   []int
}

func dodajOceno(studenti map[string]Student, vpisnaStevilka string, ocena int) {
	if ocena < 0 || ocena > 10 {
		fmt.Println("ocena mora biti na intervalu [0, 10]")
		return
	}

	student, ok := studenti[vpisnaStevilka]
	if !ok {
		fmt.Printf("student %q ne obstaja\n", vpisnaStevilka)
		return
	}

	student.ocene = append(student.ocene, ocena)

	studenti[vpisnaStevilka] = student
}

func povprecje(studenti map[string]Student, vpisnaStevilka string) float64 {
	student, ok := studenti[vpisnaStevilka]
	if !ok {
		return -1
	}

	if len(student.ocene) < 6 {
		return 0
	}

	sum := 0
	for _, ocena := range student.ocene {
		sum += ocena
	}

	return float64(sum) / float64(len(student.ocene))
}

func izpisRedovalnice(studenti map[string]Student) {
	fmt.Println("REDOVALNICA:")
	for vpisna, student := range studenti {
		fmt.Printf("%v - %v %v: %v\n", vpisna, student.ime, student.priimek, student.ocene)
	}
}

func izpisiKoncniUspeh(studenti map[string]Student) {
	for vpisna, student := range studenti {
		povprecje := povprecje(studenti, vpisna)

		var komentar string
		if povprecje >= 9 {
			komentar = "Odličen študent!"
		} else if povprecje >= 6 {
			komentar = "Povprečen študent"
		} else {
			komentar = "Neuspešen študent"
		}

		fmt.Printf("%v %v: povprečna ocena %.1f -> %v\n", student.ime, student.priimek, povprecje, komentar)
	}
}

func main() {
	studenti := map[string]Student{
		"63210001": {
			ime:     "Ana",
			priimek: "Novak",
			ocene:   []int{10, 9, 8, 10, 10},
		},
		"63210002": {
			ime:     "Boris",
			priimek: "Kralj",
			ocene:   []int{6, 7, 5, 8, 8, 8},
		},
		"63210003": {
			ime:     "Janez",
			priimek: "Novak",
			ocene:   []int{4, 5, 3, 5},
		},
	}

	fmt.Println("dodajOceno - uspešno")
	dodajOceno(studenti, "63210001", 10)

	fmt.Println("dodajOceno - študent ne obstaja")
	dodajOceno(studenti, "63210000", 10)

	fmt.Println("dodajOceno - napačna ocena")
	dodajOceno(studenti, "63210001", 11)

	fmt.Println("\n---\n")

	fmt.Println("povpracje - uspešno")
	fmt.Println(povprecje(studenti, "63210001"))

	fmt.Println("povpracje - študent ne obstaja")
	fmt.Println(povprecje(studenti, "63210000"))

	fmt.Println("povpracje - premalo ocen")
	fmt.Println(povprecje(studenti, "63210003"))

	fmt.Println("\n---\n")

	fmt.Println("izpisRedovalnice")
	izpisRedovalnice(studenti)

	fmt.Println("\n---\n")

	fmt.Println("izpisiKoncniUspeh")
	izpisiKoncniUspeh(studenti)
}
