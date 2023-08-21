package support

func NewHelm() (*Helm, error) {

	h := Helm{}

	return &h, nil
}

type Helm struct {
}

func (h *Helm) Install() error {

	return nil
}
