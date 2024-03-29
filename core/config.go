package core

import (
	"fmt"
	"os"
	"runtime/debug"
	"slices"
	"strings"

	"github.com/knadh/koanf/providers/posflag"
	"github.com/rs/zerolog"
)

// Configure configures bgpipe
func (b *Bgpipe) Configure() error {
	// parse CLI args
	err := b.parseArgs(os.Args[1:])
	if err != nil {
		return fmt.Errorf("could not parse CLI flags: %w", err)
	}

	// debugging level
	if ll := b.K.String("log"); len(ll) > 0 {
		lvl, err := zerolog.ParseLevel(ll)
		if err != nil {
			return err
		}
		zerolog.SetGlobalLevel(lvl)
	}

	return nil
}

func (b *Bgpipe) addFlags() {
	f := b.F
	f.SortFlags = false
	f.Usage = b.usage
	f.SetInterspersed(false)
	f.BoolP("version", "v", false, "print detailed version info and quit")
	f.StringP("log", "l", "info", "log level (debug/info/warn/error/disabled)")
	f.StringSliceP("events", "e", []string{"PARSE", "ESTABLISHED"}, "log given events (\"all\" means all events)")
	f.StringSliceP("kill", "k", []string{}, "kill session on given events")
	f.BoolP("stdin", "i", false, "read stdin after session is established (unless explicitly configured)")
	f.BoolP("silent", "s", false, "do not write stdout (unless explicitly configured)")
	f.BoolP("short-asn", "2", false, "use 2-byte ASN numbers")
}

func (b *Bgpipe) usage() {
	fmt.Fprintf(os.Stderr, `Usage: bgpipe [OPTIONS] [--] STAGE [STAGE-OPTIONS] [STAGE-ARGUMENTS] [--] ...

Options:
`)
	b.F.PrintDefaults()
	fmt.Fprintf(os.Stderr, `
Supported stages (run stage -h to get its help)
`)

	// iterate over cmds
	var cmds []string
	for cmd := range b.repo {
		cmds = append(cmds, cmd)
	}
	slices.Sort(cmds)
	for _, cmd := range cmds {
		var descr string

		s := b.NewStage(cmd)
		if s != nil {
			descr = s.Options.Descr
		}

		fmt.Fprintf(os.Stderr, "  %-22s %s\n", cmd, descr)
	}
	fmt.Fprintf(os.Stderr, "\n")
}

// Usage prints usage screen to stderr
func (s *StageBase) usage() {
	var (
		o = &s.Options
		f = o.Flags
		e = os.Stderr
	)

	if len(o.Usage) > 0 {
		fmt.Fprintf(e, "Stage usage: %s", o.Usage)
	} else {
		fmt.Fprintf(e, "Stage usage: %s [OPTIONS] %s",
			s.Cmd, strings.ToUpper(strings.Join(o.Args, " ")))
	}
	fmt.Fprintf(e, "\n\nDescription: %s\n", o.Descr)

	for i, l := range strings.Split(f.FlagUsages(), "\n") {
		if strings.HasPrefix(l, "  -A, --args") {
			fmt.Fprint(e, "\nCommon Options:\n")
		} else if i == 0 {
			fmt.Fprint(e, "\nOptions:\n")
		}
		fmt.Fprintf(e, "%s\n", l)
	}

	// iterate over events?
	if len(o.Events) > 0 {
		fmt.Fprint(e, "Events:\n")
		var events []string
		for e := range o.Events {
			events = append(events, e)
		}
		slices.Sort(events)
		for _, ev := range events {
			fmt.Fprintf(e, "  %-24s %s\n", ev, o.Events[ev])
		}
		fmt.Fprint(e, "\n")
	}
}

// parseArgs adds and configures stages from CLI args
func (b *Bgpipe) parseArgs(args []string) error {
	// parse and export flags into koanf
	if err := b.F.Parse(args); err != nil {
		return err
	} else {
		b.K.Load(posflag.Provider(b.F, ".", b.K), nil)
	}

	// print version and quit?
	if b.K.Bool("version") {
		if bi, ok := debug.ReadBuildInfo(); ok && bi != nil {
			fmt.Fprintf(os.Stderr, "bgpipe build info:\n%s", bi)
		}
		os.Exit(1)
	}

	// parse stages and their args
	args = b.F.Args()
	for idx := 1; len(args) > 0; idx++ {
		// skip empty stages
		if args[0] == "--" {
			args = args[1:]
			continue
		}

		// has a name prefix?
		name := ""
		if args[0][0] == '@' {
			name = args[0]
			args = args[1:]
		}

		// is args[0] a special value, or generic stage command name?
		cmd := args[0]
		switch {
		case IsAddr(cmd):
			cmd = "connect"
		case IsBind(cmd):
			cmd = "listen"
		case IsFile(cmd):
			cmd = "mrt" // TODO: stat -> mrt / exec / json / etc.
		default:
			args = args[1:]
		}

		// get s for cmd
		s, err := b.AddStage(idx, cmd)
		if err != nil {
			return err
		}

		// override the stage name?
		if name != "" {
			s.Name = name
		}

		// find an explicit end of its args
		var nextargs []string
		for i, arg := range args {
			if arg == "--" {
				nextargs = args[i+1:]
				args = args[:i]
				break
			}
		}

		// parse stage args, move on
		if remargs, err := s.parseArgs(args); err != nil {
			return err
		} else {
			args = append(remargs, nextargs...)
		}
	}

	return nil
}

// parseArgs parses CLI flags and arguments and exports to s.K.
// May return unused args.
func (s *StageBase) parseArgs(args []string) (unused []string, err error) {
	o := &s.Options
	f := o.Flags

	// override f.Usage?
	if f.Usage == nil {
		f.Usage = s.usage
	}

	// parse stage flags, export to koanf
	if err := f.Parse(args); err != nil {
		return args, s.Errorf("%w", err)
	} else {
		s.K.Load(posflag.Provider(f, ".", s.K), nil)
	}

	// rewrite required CLI arguments?
	sargs := f.Args()
	for _, name := range o.Args {
		if len(sargs) == 0 {
			return sargs, s.Errorf("needs an argument: %s", name)
		}
		s.K.Set(name, sargs[0])
		sargs = sargs[1:]
	}

	// consume the rest of arguments?
	if v, _ := f.GetBool("args"); v {
		s.K.Set("args", sargs)
		return nil, nil
	}

	return sargs, nil
}
