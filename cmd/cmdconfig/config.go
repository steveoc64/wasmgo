package cmdconfig

type Config struct {
	Port      int
	Index     string
	Template  string
	Json      bool
	Verbose   bool
	Open      bool
	Command   string
	Flags     string
	BuildTags string
	Path      string
	Cache     bool
}
