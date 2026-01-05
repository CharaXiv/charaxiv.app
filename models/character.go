package models

// Variable represents an ability score with base, perm, and temp modifiers
type Variable struct {
	Base int `json:"base"`
	Perm int `json:"perm"`
	Temp int `json:"temp"`
	Min  int `json:"min"`
	Max  int `json:"max"`
}

// Sum returns the total value of the variable
func (v Variable) Sum() int {
	return v.Base + v.Perm + v.Temp
}

// Cthulhu6Status represents the status section for CoC 6th edition
type Cthulhu6Status struct {
	Variables  map[string]Variable `json:"variables"`
	Parameters map[string]*int     `json:"parameters"` // nil means use default
	DB         string              `json:"db"`
}

// NewCthulhu6Status creates a new status with default values
func NewCthulhu6Status() *Cthulhu6Status {
	return &Cthulhu6Status{
		Variables: map[string]Variable{
			"STR": {Base: 10, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"CON": {Base: 12, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"POW": {Base: 11, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"DEX": {Base: 13, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"APP": {Base: 9, Perm: 0, Temp: 0, Min: 3, Max: 18},
			"SIZ": {Base: 14, Perm: 0, Temp: 0, Min: 8, Max: 18},
			"INT": {Base: 15, Perm: 0, Temp: 0, Min: 8, Max: 18},
			"EDU": {Base: 16, Perm: 0, Temp: 0, Min: 6, Max: 21},
		},
		Parameters: map[string]*int{
			"HP":  nil,
			"MP":  nil,
			"SAN": nil,
		},
		DB: "",
	}
}

// ComputedValues returns the derived values from variables
func (s *Cthulhu6Status) ComputedValues() map[string]int {
	pow := s.Variables["POW"].Sum()
	inT := s.Variables["INT"].Sum()
	edu := s.Variables["EDU"].Sum()

	return map[string]int{
		"初期SAN": pow * 5,
		"アイデア":  inT * 5,
		"幸運":    pow * 5,
		"知識":    edu * 5,
		"職業P":   edu * 20,
		"興味P":   inT * 10,
	}
}

// DefaultParameters returns the default parameter values derived from variables
func (s *Cthulhu6Status) DefaultParameters() map[string]int {
	con := s.Variables["CON"].Sum()
	pow := s.Variables["POW"].Sum()
	siz := s.Variables["SIZ"].Sum()

	return map[string]int{
		"HP":  (con + siz + 1) / 2,
		"MP":  pow,
		"SAN": pow * 5,
	}
}

// DamageBonus calculates the damage bonus from STR+SIZ
func (s *Cthulhu6Status) DamageBonus() string {
	if s.DB != "" {
		return s.DB
	}
	str := s.Variables["STR"].Sum()
	siz := s.Variables["SIZ"].Sum()
	sum := str + siz

	switch {
	case sum < 13:
		return "-1d6"
	case sum < 17:
		return "-1d4"
	case sum < 25:
		return "+0"
	case sum < 33:
		return "+1d4"
	default:
		return "+1d6"
	}
}

// Indefinite calculates the indefinite insanity threshold
func (s *Cthulhu6Status) Indefinite() int {
	san := s.EffectiveParameter("SAN")
	return (san * 4) / 5
}

// EffectiveParameter returns the parameter value or its default
func (s *Cthulhu6Status) EffectiveParameter(key string) int {
	if val := s.Parameters[key]; val != nil {
		return *val
	}
	return s.DefaultParameters()[key]
}

// SkillCategory represents a skill category
type SkillCategory string

const (
	SkillCategoryCombat        SkillCategory = "戦闘技能"
	SkillCategoryInvestigation SkillCategory = "探索技能"
	SkillCategoryAction        SkillCategory = "行動技能"
	SkillCategorySocial        SkillCategory = "交渉技能"
	SkillCategoryKnowledge     SkillCategory = "知識技能"
)

// SkillCategoryOrder returns the display order for a category
func SkillCategoryOrder(cat SkillCategory) int {
	switch cat {
	case SkillCategoryCombat:
		return 0
	case SkillCategoryInvestigation:
		return 1
	case SkillCategoryAction:
		return 2
	case SkillCategorySocial:
		return 3
	case SkillCategoryKnowledge:
		return 4
	default:
		return 99
	}
}

// Cthulhu6SingleSkill represents a simple skill with point allocations
type Cthulhu6SingleSkill struct {
	Job   int  `json:"job"`
	Hobby int  `json:"hobby"`
	Perm  int  `json:"perm"`
	Temp  int  `json:"temp"`
	Grow  bool `json:"grow"`
}

// Sum returns total allocated points
func (s *Cthulhu6SingleSkill) Sum() int {
	return s.Job + s.Hobby + s.Perm + s.Temp
}

// Cthulhu6SkillGenre represents one specialty within a multi-genre skill
type Cthulhu6SkillGenre struct {
	Label string `json:"label"` // e.g., "自動車" for 運転
	Job   int    `json:"job"`
	Hobby int    `json:"hobby"`
	Perm  int    `json:"perm"`
	Temp  int    `json:"temp"`
	Grow  bool   `json:"grow"`
}

// Sum returns total allocated points for this genre
func (g *Cthulhu6SkillGenre) Sum() int {
	return g.Job + g.Hobby + g.Perm + g.Temp
}

// Cthulhu6MultiSkill represents a skill with multiple specialties
type Cthulhu6MultiSkill struct {
	Genres []Cthulhu6SkillGenre `json:"genres"`
}

// TotalPoints returns sum of all genre points
func (m *Cthulhu6MultiSkill) TotalPoints() (job, hobby int) {
	for _, g := range m.Genres {
		job += g.Job
		hobby += g.Hobby
	}
	return
}

// Cthulhu6Skill wraps either a single or multi skill (exactly one will be non-nil)
type Cthulhu6Skill struct {
	Order  int                  `json:"order"`
	Single *Cthulhu6SingleSkill `json:"single,omitempty"`
	Multi  *Cthulhu6MultiSkill  `json:"multi,omitempty"`
}

// IsSingle returns true if this is a single skill
func (s Cthulhu6Skill) IsSingle() bool {
	return s.Single != nil
}

// IsMulti returns true if this is a multi-genre skill
func (s Cthulhu6Skill) IsMulti() bool {
	return s.Multi != nil
}

// Cthulhu6SkillExtra represents extra skill points
type Cthulhu6SkillExtra struct {
	Job   int `json:"job"`
	Hobby int `json:"hobby"`
}

// Cthulhu6SkillCategoryData represents a category of skills
type Cthulhu6SkillCategoryData struct {
	Skills map[string]Cthulhu6Skill `json:"skills"`
	Order  int                      `json:"order"`
}

// Cthulhu6Skills represents all skills for a character
type Cthulhu6Skills struct {
	Categories map[SkillCategory]Cthulhu6SkillCategoryData `json:"categories"`
	Custom     []Cthulhu6Skill                             `json:"custom"`
	Extra      Cthulhu6SkillExtra                          `json:"extra"`
}

// singleSkill creates a single skill
func singleSkill(order int) Cthulhu6Skill {
	return Cthulhu6Skill{Order: order, Single: &Cthulhu6SingleSkill{}}
}

// multiSkill creates an empty multi-genre skill
func multiSkill(order int) Cthulhu6Skill {
	return Cthulhu6Skill{Order: order, Multi: &Cthulhu6MultiSkill{Genres: []Cthulhu6SkillGenre{}}}
}

// multiSkillWithGenre creates a multi-genre skill with one initial genre
func multiSkillWithGenre(order int, label string) Cthulhu6Skill {
	return Cthulhu6Skill{Order: order, Multi: &Cthulhu6MultiSkill{Genres: []Cthulhu6SkillGenre{{Label: label}}}}
}

// NewCthulhu6Skills creates skills with default values
func NewCthulhu6Skills() *Cthulhu6Skills {
	return &Cthulhu6Skills{
		Categories: map[SkillCategory]Cthulhu6SkillCategoryData{
			SkillCategoryCombat: {
				Order: 0,
				Skills: map[string]Cthulhu6Skill{
					"回避":       singleSkill(0),
					"キック":      singleSkill(1),
					"組み付き":     singleSkill(2),
					"こぶし":      singleSkill(3),
					"頭突き":      singleSkill(4),
					"投擲":       singleSkill(5),
					"マーシャルアーツ": singleSkill(6),
					"拳銃":       singleSkill(7),
					"サブマシンガン":  singleSkill(8),
					"ショットガン":   singleSkill(9),
					"マシンガン":    singleSkill(10),
					"ライフル":     singleSkill(11),
				},
			},
			SkillCategoryInvestigation: {
				Order: 1,
				Skills: map[string]Cthulhu6Skill{
					"目星":    singleSkill(0),
					"聞き耳":   singleSkill(1),
					"図書館":   singleSkill(2),
					"応急手当":  singleSkill(3),
					"隠れる":   singleSkill(4),
					"隠す":    singleSkill(5),
					"変装":    singleSkill(6),
					"忍び歩き":  singleSkill(7),
					"追跡":    singleSkill(8),
					"ナビゲート": singleSkill(9),
					"写真術":   singleSkill(10),
					"鍵開け":   singleSkill(11),
					"精神分析":  singleSkill(12),
				},
			},
			SkillCategoryAction: {
				Order: 2,
				Skills: map[string]Cthulhu6Skill{
					"登攀":    singleSkill(0),
					"跳躍":    singleSkill(1),
					"運転":    multiSkill(2),
					"操縦":    multiSkill(3),
					"重機械操作": singleSkill(4),
					"機械修理":  singleSkill(5),
					"電気修理":  singleSkill(6),
					"製作":    multiSkill(7),
					"芸術":    multiSkill(8),
					"乗馬":    singleSkill(9),
					"水泳":    singleSkill(10),
				},
			},
			SkillCategorySocial: {
				Order: 3,
				Skills: map[string]Cthulhu6Skill{
					"言いくるめ": singleSkill(0),
					"信用":    singleSkill(1),
					"説得":    singleSkill(2),
					"値切り":   singleSkill(3),
				},
			},
			SkillCategoryKnowledge: {
				Order: 4,
				Skills: map[string]Cthulhu6Skill{
					"クトゥルフ神話": singleSkill(0),
					"心理学":     singleSkill(1),
					"母国語":     multiSkillWithGenre(2, ""),
					"ほかの言語":   multiSkill(3),
					"オカルト":    singleSkill(4),
					"歴史":      singleSkill(5),
					"法律":      singleSkill(6),
					"経理":      singleSkill(7),
					"人類学":     singleSkill(8),
					"考古学":     singleSkill(9),
					"博物学":     singleSkill(10),
					"医学":      singleSkill(11),
					"薬学":      singleSkill(12),
					"生物学":     singleSkill(13),
					"化学":      singleSkill(14),
					"コンピューター": singleSkill(15),
					"電子工学":    singleSkill(16),
					"物理学":     singleSkill(17),
					"天文学":     singleSkill(18),
					"地質学":     singleSkill(19),
				},
			},
		},
		Custom: []Cthulhu6Skill{},
		Extra:  Cthulhu6SkillExtra{Job: 0, Hobby: 0},
	}
}

// SkillInitialValue returns the initial value for a skill based on character stats
func (s *Cthulhu6Status) SkillInitialValue(skillKey string) int {
	switch skillKey {
	case "回避":
		return s.Variables["DEX"].Sum() * 2
	case "母国語":
		return s.Variables["EDU"].Sum() * 5
	case "キック", "組み付き", "投擲", "目星", "聞き耳", "図書館", "跳躍", "水泳", "ライフル":
		return 25
	case "こぶし":
		return 50
	case "頭突き", "隠れる", "忍び歩き", "追跡", "ナビゲート", "写真術", "電気修理":
		return 10
	case "応急手当", "ショットガン":
		return 30
	case "隠す", "サブマシンガン", "マシンガン", "信用", "説得":
		return 15
	case "登攀":
		return 40
	case "運転", "機械修理", "拳銃", "歴史":
		return 20
	case "製作", "芸術", "乗馬", "言いくるめ", "値切り", "オカルト", "心理学", "法律", "医学":
		return 5
	case "経理", "博物学":
		return 10
	case "マーシャルアーツ", "変装", "鍵開け", "精神分析", "操縦", "重機械操作",
		"ほかの言語", "薬学", "生物学", "化学", "コンピューター", "電子工学", "物理学", "天文学", "人類学", "考古学", "地質学":
		return 1
	case "クトゥルフ神話":
		return 0
	default:
		return 1
	}
}

// RemainingPoints calculates remaining skill points
func (s *Cthulhu6Status) RemainingPoints(skills *Cthulhu6Skills) (job int, hobby int) {
	computed := s.ComputedValues()
	totalJob := computed["職業P"] + skills.Extra.Job
	totalHobby := computed["興味P"] + skills.Extra.Hobby

	usedJob := 0
	usedHobby := 0
	for _, catData := range skills.Categories {
		for _, skill := range catData.Skills {
			if skill.IsMulti() {
				for _, g := range skill.Multi.Genres {
					usedJob += g.Job
					usedHobby += g.Hobby
				}
			} else if skill.IsSingle() {
				usedJob += skill.Single.Job
				usedHobby += skill.Single.Hobby
			}
		}
	}
	for _, skill := range skills.Custom {
		if skill.IsSingle() {
			usedJob += skill.Single.Job
			usedHobby += skill.Single.Hobby
		}
	}

	return totalJob - usedJob, totalHobby - usedHobby
}
