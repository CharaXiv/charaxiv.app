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

// Cthulhu6Skill represents a single skill
type Cthulhu6Skill struct {
	Job   int  `json:"job"`
	Hobby int  `json:"hobby"`
	Perm  int  `json:"perm"`
	Temp  int  `json:"temp"`
	Grow  bool `json:"grow"`
	Order int  `json:"order"`
}

// Sum returns total allocated points
func (s Cthulhu6Skill) Sum() int {
	return s.Job + s.Hobby + s.Perm + s.Temp
}

// Cthulhu6SkillExtra represents extra skill points
type Cthulhu6SkillExtra struct {
	Job   int `json:"job"`
	Hobby int `json:"hobby"`
}

// Cthulhu6Skills represents all skills for a character
type Cthulhu6Skills struct {
	Skills map[string]Cthulhu6Skill `json:"skills"`
	Extra  Cthulhu6SkillExtra       `json:"extra"`
}

// NewCthulhu6Skills creates skills with default values
func NewCthulhu6Skills() *Cthulhu6Skills {
	return &Cthulhu6Skills{
		Skills: map[string]Cthulhu6Skill{
			// 戦闘技能 (Combat Skills) - Order 0-11
			"回避":       {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 0},
			"キック":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 1},
			"組み付き":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 2},
			"こぶし":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 3},
			"頭突き":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 4},
			"投擲":       {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 5},
			"マーシャルアーツ": {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 6},
			"拳銃":       {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 7},
			"サブマシンガン":  {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 8},
			"ショットガン":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 9},
			"マシンガン":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 10},
			"ライフル":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 11},

			// 探索技能 (Investigation Skills) - Order 12-24
			"目星":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 12},
			"聞き耳":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 13},
			"図書館":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 14},
			"応急手当":  {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 15},
			"隠れる":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 16},
			"隠す":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 17},
			"変装":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 18},
			"忍び歩き":  {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 19},
			"追跡":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 20},
			"ナビゲート": {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 21},
			"写真術":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 22},
			"鍵開け":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 23},
			"精神分析":  {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 24},

			// 行動技能 (Action Skills) - Order 25-35
			"登攀":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 25},
			"跳躍":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 26},
			"運転":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 27},
			"操縦":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 28},
			"重機械操作": {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 29},
			"機械修理":  {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 30},
			"電気修理":  {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 31},
			"製作":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 32},
			"芸術":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 33},
			"乗馬":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 34},
			"水泳":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 35},

			// 交渉技能 (Social Skills) - Order 36-39
			"言いくるめ": {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 36},
			"信用":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 37},
			"説得":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 38},
			"値切り":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 39},

			// 知識技能 (Knowledge Skills) - Order 40-59
			"クトゥルフ神話": {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 40},
			"心理学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 41},
			"母国語":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 42},
			"ほかの言語":   {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 43},
			"オカルト":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 44},
			"歴史":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 45},
			"法律":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 46},
			"経理":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 47},
			"人類学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 48},
			"考古学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 49},
			"博物学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 50},
			"医学":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 51},
			"薬学":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 52},
			"生物学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 53},
			"化学":      {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 54},
			"コンピューター": {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 55},
			"電子工学":    {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 56},
			"物理学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 57},
			"天文学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 58},
			"地質学":     {Job: 0, Hobby: 0, Perm: 0, Temp: 0, Grow: false, Order: 59},
		},
		Extra: Cthulhu6SkillExtra{Job: 0, Hobby: 0},
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
	for _, skill := range skills.Skills {
		usedJob += skill.Job
		usedHobby += skill.Hobby
	}

	return totalJob - usedJob, totalHobby - usedHobby
}
