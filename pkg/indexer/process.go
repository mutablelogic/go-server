package indexer

import (
	"io/fs"

	// Modules
	. "github.com/djthorpe/go-server"
)

func (this *Indexer) ProcessPath(event EventType, path string, info fs.FileInfo) error {
	evt, err := this.NewEvent(event, path, info)
	if err != nil {
		return err
	}
	select {
	case this.c <- evt:
		return nil
	default:
		return ErrChannelBlocked.With(this.name)
	}
}
