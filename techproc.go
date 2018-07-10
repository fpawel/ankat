package ankat

const (
	TechLin1 = "LIN1"
	TechLin2 = "LIN2"
)

func TechProcessName(s string) string {
	return techProcessName[s]
}

func TechProcesses() (xs []string) {
	for s := range techProcessName {
		xs = append(xs, s)
	}
	return
}

var techProcessName = map[string]string{
	TechLin1: "линеаризация канала 1",
	TechLin2: "линеаризация канала 2",
}
