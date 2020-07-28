package interfaces

type IProxy interface {
	Run()
	AddEvents()
	RemoveEvents()
	RegisterMessages()
	CancelMessages()
	Destroy()
}
