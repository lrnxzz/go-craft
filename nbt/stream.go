package nbt

import "io"

func Read(r io.Reader) (Compound, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return Decode(data)
}

func ReadNamed(r io.Reader) (string, Compound, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", nil, err
	}

	return DecodeNamed(data)
}

func Write(w io.Writer, root Compound) error {
	_, err := w.Write(Encode(root))

	return err
}

func WriteNamed(w io.Writer, name string, root Compound) error {
	_, err := w.Write(EncodeNamed(name, root))

	return err
}
