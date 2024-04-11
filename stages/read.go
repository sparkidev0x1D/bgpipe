package stages

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/bgpfix/bgpipe/core"
	"github.com/bgpfix/bgpipe/pkg/extio"
)

type Read struct {
	*core.StageBase
	eio   *extio.Extio
	fpath string
	flag  int
	fh    *os.File
}

func NewRead(parent *core.StageBase) core.Stage {
	s := &Read{StageBase: parent}

	o := &s.Options
	o.IsProducer = true
	o.Bidir = true
	o.Descr = "read messages from file"
	o.Args = []string{"path"}

	s.eio = extio.NewExtio(parent, 1)
	return s
}

func (s *Read) Attach() error {
	k := s.K

	s.fpath = k.String("path")
	if len(s.fpath) == 0 {
		return errors.New("path must be set")
	}
	s.fpath = filepath.Clean(s.fpath)
	s.flag = os.O_RDONLY

	k.Set("read", true)
	return s.eio.Attach()
}

func (s *Read) Prepare() error {
	s.Info().Msgf("opening %s", s.fpath)
	fh, err := os.OpenFile(s.fpath, s.flag, 0666)
	if err != nil {
		return err
	}
	s.fh = fh // closed in .Stop()
	return nil
}

func (s *Read) Run() error {
	return s.eio.ReadStream(s.fh, nil)
}

func (s *Read) Stop() error {
	s.eio.InputClose()
	s.fh.Close()
	return nil
}
