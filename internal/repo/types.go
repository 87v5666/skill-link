package repo

// Skill 代表仓库中的一个可链接 skill
type Skill struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Linked      bool   `json:"linked"`
}

// Category 代表一个分类
type Category struct {
	Name   string  `json:"name"`
	Skills []Skill `json:"skills"`
}