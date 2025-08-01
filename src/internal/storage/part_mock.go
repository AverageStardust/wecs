package storage

type partMock uint32

func (part partMock) PartId() PartId { return PartId(part) }
