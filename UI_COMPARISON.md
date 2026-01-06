# UI Comparison: Legacy CharaXiv vs Current Implementation

This document compares the legacy CharaXiv (https://charaxiv.app/demo/cthulhu6/) with the current Hono/JSX implementation.

## Overall Layout

### Legacy
- Three-column layout: Profile/Memos | Status/Parameters | Skills
- All columns visible simultaneously on desktop
- Purple promotional banner at bottom ("アカウント登録はこちら") - demo page only, not required

### Current
- Similar three-column layout preserved ✓
- Same column ordering ✓

## Header

### Legacy
- CharaXiv logo with icon (left)
- Three header buttons (right): Eye icon, Person icon, Share icon (blue filled)
- System name displayed: "クトゥルフ神話TRPG（第6版）"

### Current
- CharaXiv logo with icon (left) ✓
- Three header buttons match legacy ✓
- **Missing**: System name title ("クトゥルフ神話TRPG（第6版）") is not displayed

## Image Gallery

### Legacy
- Image placeholder when no image
- Navigation arrows: Left `<` | "画像一覧" text | Right `>`
- Action buttons: Delete (red), Upload (blue), Download (green), Pin (yellow), Crop(?)
- Buttons are colored with distinct backgrounds

### Current
- Image placeholder matches ✓
- Navigation arrows match ✓
- Action buttons match in function ✓
- **Difference**: Current buttons appear more subdued/outline style vs legacy's filled colors

## Profile Section

### Legacy
- Name field: "名前" (red text placeholder style)
- Ruby field: "よみがな" (gray subtext)

### Current
- Name field: "名前" ✓
- Ruby field: "よみがな" ✓
- Layout matches

## Status Panel (能力値)

### Legacy
- Title: "能力値" with expand/collapse icon
- Ability rows: Label | `<<` `<` value `>` `>>` | total
- Variables: STR, CON, POW, DEX, APP, SIZ, INT, EDU
- Computed values below in 2x2 grid: 初期SAN, アイデア, 幸運, 知識
- Points display: 職業P, 興味P with values
- Damage bonus row

### Current
- Title: "能力値" with "ランダム" button instead of expand icon
- **Difference**: Legacy has an expand/collapse icon next to title, current has a "ランダム" (Random) button
- Ability rows match the format ✓
- Variables match ✓
- Computed values layout matches ✓
- Points display matches ✓

## Parameters Panel (パラメーター)

### Legacy
- HP, MP, SAN with `<<` `<` value `>` `>>` controls and max values
- DB (Damage Bonus): "+1d4" style text
- 不定 (Indefinite): Numeric value

### Current
- HP, MP, SAN with controls ✓
- DB: Shows "+0" format ✓
- 不定: Shows numeric value ✓
- Layout matches

## Skills Panel (技能)

### Legacy
- Header: "技能" with search/filter icon
- "追加技能ポイント" section with 職業P/興味P adjustment controls
- Skill categories: 戦闘技能, 探索技能, 行動技能, etc.
- Each skill row: Checkbox | Name | Value | Dropdown arrow
- Points breakdown: `55/00/00` format (職/興/増一) visible for some skills
- Floating points display: "職業P 260 興味P 130" (bottom right, fixed position)

### Current
- Header: "技能" without search icon
- **Missing**: Search/filter icon in skills header
- "追加技能ポイント" section matches ✓
- Skill categories match ✓
- Skill rows: Checkbox | Name | `00/00/00` breakdown | Value | Dropdown ✓
- **Difference**: Points breakdown is always visible in current (00/00/00 format), legacy shows breakdown only for modified skills
- Floating points display: "職業P 235 興味P 140" ✓

### Skill Row Details (Expanded)

### Legacy
- Clicking dropdown expands skill details
- Details show: 職業P, 興味P, 増加分, 一時的 inputs

### Current
- Same expansion behavior ✓
- Same detail fields ✓
- Detail panel is indented with gray background ✓

## Memo Sections

### Legacy
- 公開メモ: Markdown editor with toolbar (H, quote, bullet, numbered list, B, I, S, link)
- 秘匿メモ: Same editor with lock icon for visibility toggle
- シナリオ公開メモ, シナリオ秘匿メモ: Additional memo sections for scenarios

### Current
- 公開メモ: Editor with toolbar ✓
- 秘匿メモ: Editor ✓
- シナリオ公開メモ: Editor ✓
- シナリオ秘匿メモ: Editor ✓
- **Difference**: Placeholder text differs slightly
  - Legacy: Shows existing content or empty
  - Current: Shows "公開メモを入力...", "秘匿メモを入力...", etc.

## Multi-Genre Skills (運転, 芸術, etc.)

### Legacy
- Genre categories shown: 運転, 操縦, 製作, 芸術, 母国語, ほかの言語
- Each genre has expandable sub-skills
- Sub-skills show value and totals

### Current
- Same genre categories ✓
- **Addition**: Plus (+) button next to genre categories for adding new sub-skills
- Sub-skills with "(専門を入力)" placeholder for unnamed genres
- Expandable sub-skills ✓

## Custom Skills (独自技能)

### Legacy
- Section at bottom of skills list
- Add button for new custom skills

### Current
- "独自技能" section ✓
- Plus (+) button for adding ✓
- "独自技能なし" message when empty ✓

## Visual Design Differences

### Colors & Styling
| Element | Legacy | Current |
|---------|--------|----------|
| Primary button | Blue filled | Blue filled (matches) |
| Secondary buttons | Colored fills (red/green/yellow) | Outlined/ghost style |
| Text colors | Slate gray scale | Slate gray scale (matches) |
| Card backgrounds | White with shadow | White with shadow (matches) |
| Inactive skill text | Gray | Gray (matches) |

### Typography
- Both use similar font sizing and weights
- Category headers use uppercase/semibold style

## Missing Features in Current Implementation

1. ~~**System title display** - "クトゥルフ神話TRPG（第6版）" not shown~~ **RESOLVED** - Added as editable input with system name as placeholder
2. **Skills filter/search icon** - Present in legacy header, missing in current
3. **Status panel collapse icon** - Legacy has expand/collapse, current has Random button instead
4. ~~**EasyMDE styles** - Markdown editor styles are not being applied correctly~~ **RESOLVED** - EasyMDE styles match legacy

## Additional Features in Current Implementation

1. **Points breakdown always visible** - Shows 00/00/00 for all skills, not just modified ones
2. ~~**Explicit add buttons** - Plus icons for genre skills and custom skills are more prominent~~ **RESOLVED** - Added blue "「X」技能を追加" buttons matching legacy
3. **Random button** - Directly in status panel header for quick stat generation

## Summary of Core Differences

| Feature | Legacy | Current | Priority |
|---------|--------|---------|----------|
| System title | ✓ Shown | ✓ Added | **DONE** |
| EasyMDE styles | ✓ Applied | ✓ Applied | **DONE** |
| Skills filter | ✓ Icon present | ✗ Missing | Low |
| Status expand/collapse | ✓ Present | Different (Random btn) | Low |
| Points breakdown visibility | Conditional | Always visible | Design choice |
| Gallery button colors | Filled colors | Outline style | Style choice |

## Recommendations

1. ~~**Fix EasyMDE styles**: Ensure the EasyMDE CSS is properly loaded/applied to markdown editors~~ **DONE** - Verified EasyMDE renders identically to legacy
2. ~~**Add system title**: Display "クトゥルフ神話TRPG（第6版）" above the status panel~~ **DONE**
3. **Consider skills filter**: If this is a frequently used feature, add the filter icon back
4. **Gallery button styling**: Consider reverting to filled colored buttons for better visual distinction
