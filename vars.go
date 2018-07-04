package ankat

const (
	CoutCh1 = 0
	TppCh1  = 642
	UwCh1   = 648
	UrCh1   = 650
	WorkCh1 = 652
	RefCh1  = 654
	Var1Ch1 = 656
	Var3Ch1 = 660

	CoutCh2 = 2
	TppCh2  = 674
	UwCh2   = 680
	UrCh2   = 682
	WorkCh2 = 684
	RefCh2  = 686
	Var1Ch2 = 688
	Var3Ch2 = 692
)

type GasCode int

const (
	CloseGasBlock GasCode = 0
	Nitrogen      GasCode = 1
)

func MainVars1() []int {
	return []int{
		CoutCh1,
		TppCh1,
		UwCh1,
		UrCh1,
		WorkCh1,
		RefCh1,
		Var1Ch1,
		Var3Ch1,
	}
}

func MainVars2() []int {
	return []int{
		CoutCh2,
		TppCh2,
		UwCh2,
		UrCh2,
		WorkCh2,
		RefCh2,
		Var1Ch2,
		Var3Ch2,
	}
}
