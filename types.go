package aur

// Pkg holds package information.
type Pkg struct {
	ID             int      `json:"ID"`
	Name           string   `json:"Name"`
	PackageBaseID  int      `json:"PackageBaseID"`
	PackageBase    string   `json:"PackageBase"`
	Version        string   `json:"Version"`
	Description    string   `json:"Description"`
	URL            string   `json:"URL"`
	NumVotes       int      `json:"NumVotes"`
	Popularity     float64  `json:"Popularity"`
	OutOfDate      int      `json:"OutOfDate"`
	Maintainer     string   `json:"Maintainer"`
	Submitter      string   `json:"Submitter"`
	FirstSubmitted int      `json:"FirstSubmitted"`
	LastModified   int      `json:"LastModified"`
	URLPath        string   `json:"URLPath"`
	Depends        []string `json:"Depends"`
	MakeDepends    []string `json:"MakeDepends"`
	CheckDepends   []string `json:"CheckDepends"`
	Conflicts      []string `json:"Conflicts"`
	Provides       []string `json:"Provides"`
	Replaces       []string `json:"Replaces"`
	OptDepends     []string `json:"OptDepends"`
	Groups         []string `json:"Groups"`
	License        []string `json:"License"`
	Keywords       []string `json:"Keywords"`
	CoMaintainers  []string `json:"CoMaintainers"`
}

func (p *Pkg) String() string {
	return p.Name
}

// By specifies what to search by in RPC searches.
type By int

const (
	Name By = iota + 1
	NameDesc
	Maintainer
	Submitter
	Depends
	MakeDepends
	OptDepends
	CheckDepends
	None
	Provides
	Conflicts
	Replaces
	Keywords
	Groups
	CoMaintainers
)

func (by By) String() string {
	switch by {
	case Name:
		return "name"
	case NameDesc:
		return "name-desc"
	case Maintainer:
		return "maintainer"
	case Submitter:
		return "submitter"
	case Depends:
		return "depends"
	case MakeDepends:
		return "makedepends"
	case OptDepends:
		return "optdepends"
	case CheckDepends:
		return "checkdepends"
	case None:
		return ""
	case Provides:
		return "provides"
	case Conflicts:
		return "conflicts"
	case Replaces:
		return "replaces"
	case Keywords:
		return "keywords"
	case Groups:
		return "groups"
	case CoMaintainers:
		return "comaintainers"
	default:
		panic("invalid By")
	}
}
