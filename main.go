package main

import (
    "fmt"
	"io"
    "log"
    "net"
    "strconv"
	"strings" // Needed for Join
    "time"

    tea "github.com/charmbracelet/bubbletea"
    "github.com/charmbracelet/bubbles/list"
    "github.com/charmbracelet/bubbles/spinner"
    "github.com/charmbracelet/bubbles/viewport"
    "github.com/charmbracelet/lipgloss"
    wish "github.com/charmbracelet/wish"
    "github.com/charmbracelet/wish/activeterm"
    wb "github.com/charmbracelet/wish/bubbletea"
    wl "github.com/charmbracelet/wish/logging"
    "github.com/charmbracelet/ssh"
)

// --- Data Structs ---

type educationItem struct {
	School  string
	Status  string
	Dates   string
	Details []string
}

type experienceItem struct {
	Company     string
	Dates       string
	Role        string
	Location    string
	Reporting   string
	Description []string
}

type projectItem struct {
	Name        string
	Dates       string
	Description []string
}

type skillsItem struct {
	Category string
	Details  []string
}

type contactItem struct {
	Line string
}

// Interface for list items
type listItemData interface {
	FilterValue() string // For list filtering (though disabled)
}

// Implement FilterValue for each type
func (e educationItem) FilterValue() string { return e.School }
func (e experienceItem) FilterValue() string { return e.Company + " " + e.Role }
func (p projectItem) FilterValue() string { return p.Name }
func (s skillsItem) FilterValue() string { return s.Category }
func (c contactItem) FilterValue() string { return c.Line }

// --- Updated Résumé Data ---

var resumeData = map[string][]listItemData{
	"Education": {
		educationItem{
			School:  "Germantown Friends School",
			Status:  "Junior",
			Dates:   "Present",
			Details: []string{"Relevant Coursework: CS 1-3, CS Capstone, Differential Calculus"},
		},
	},
	"Experience": {
		experienceItem{
			Company:  "Honeycake",
			Dates:    "February 2025 - March 2025",
			Role:     "Dev Ops Intern",
			Location: "Philadelphia, PA",
			Reporting:"Reported to Monica Quigg, Co-Founder / VP Engineering",
			Description: []string{
				"Set up and deployed python api in Google Cloud Platform",
				"Installed postgres database",
				"Trained other employees on access, and procedures for code and database updates",
			},
		},
		experienceItem{
			Company:  "Human Security",
			Dates:    "January 2025",
			Role:     "Data Analyst Intern",
			Location: "New York, NY",
			Reporting:"Reported to Francis Kitrick, Manager, Strategic Customer Research",
			Description: []string{
				"Analyzed suspicious internet traffic to determine origin and whether it was malicious",
				"Leveraged multiple RDBMS databases",
				"Successfully detected multiple \"cash-out\" domains", // Use standard escaped quotes
			},
		},
		experienceItem{
			Company:  "Good-Loop.com (B Corp)",
			Dates:    "June 2024 – August 2024",
			Role:     "Software Development Intern",
			Location: "Edinburgh, UK",
			Reporting:"Worked directly under Craig Robertson, PhD, Head of Engineering",
			Description: []string{
				"Designed and implemented a headless web scraper using Playwright and python for the purpose of automating the classification of websites using AI.",
				"Analyzed the features extracted through the web scraper with a Self Organizing Map",
				"Patent application in progress",
			},
		},
		experienceItem{
			Company:  "Seeds of Fortune (Non-Profit)",
			Dates:    "June 2023 - August 2023",
			Role:     "Software Development Intern",
			Location: "Remote",
			Reporting:"Reported to Executive Director Nitiya Walker",
			Description: []string{
				"Created a gamified financial simulation using React and Next.js for the purpose of providing disadvantaged high school students the ability to create budgets for college life.",
				"Implemented continuous revisions based on customer feedback",
			},
		},
	},
	"Projects": {
		projectItem{
			Name:  "EMP",
			Dates: "April 2025 - Present",
			Description: []string{
				"Researched EMP design, feasibility, and safety",
				"Designed a small EMP generator and current WIP",
			},
		},
		projectItem{
			Name:  "Micropantry (Volunteer)",
			Dates: "February 2025 - Present",
			Description: []string{
				"Collaborated with classmates to design an app that tracks your pantry to recommend personalized healthy recipes using AI.",
				"Wrote the backend in Express and deployed it as a Firebase Function.",
				"Won Comcast student competition",
			},
		},
		projectItem{
			Name:  "Epicliem.com",
			Dates: "January 2022 - present",
			Description: []string{
				"Functionally serves as a repo and showcase for my work",
				"Deployed to AWS using their stack (Route 53, etc.)",
				"Migrated from a static site hosted on an S3 bucket to a dynamic, interactive webpage using Vercel and Next.js.",
			},
		},
		projectItem{
			Name:  "Wake On Lan Server",
			Dates: "September - October 2023",
			Description: []string{
				"Deployed Ubuntu on a used Dell Optiplex.",
				"Ran a containerized Wake On Lan Server written in Python that sent a magic packet to my computer to wake it up using Docker and deployed the server on a Dell Optiplex in my garage.",
				"Installed Docker containers, wireguard vpn, bitwarden, and a minecraft server",
			},
		},
		projectItem{
			Name:  "ChessAI",
			Dates: "June 2023",
			Description: []string{
				"Researched muzero and alphazero papers to determine best approach",
				"Implemented muzero",
				"Trained the model using Google Cloud TPUs",
			},
		},
		projectItem{
			Name:  "Miller-Rabin Primality Test",
			Dates: "October 2022",
			Description: []string{
				"Researched prime number algorithms",
				"Implemented in Rust as a means to better learn that language",
			},
		},
	},
	"Skills & Interests": {
		skillsItem{
			Category: "Programming Languages",
			Details:  []string{"Python", "JavaScript", "Rust", "Java", "fish and bash shell scripting"},
		},
		skillsItem{
			Category: "Tools",
			Details:  []string{"GitHub", "AWS", "Google Cloud Platform", "Docker", "RDMS", "Wireguard", "Linux", "Windows", "Mac", "Playwright"},
		},
		skillsItem{
			Category: "Interests",
			Details:  []string{"I am passionate about cooking, and always on the lookout for new kitchen tools to expand my kitchen capabilities!"},
		},
		skillsItem{
			Category: "Volunteering",
			Details:  []string{"SAT tutor at Schoolhouse", "Manna", "Also note that Seeds of Fortune and Micropantry are volunteer / non-profit."},
		},
	},
	"Contact": {
		contactItem{Line: "Liem Luttrell"},
		contactItem{Line: "819 N 4th St"},
		contactItem{Line: "+1 (267) 800-4362"}, // Use standard hyphen
		contactItem{Line: "liem@epicliem.com"},
		contactItem{Line: "https://epicliem.com"},
	},
}

// Order of sections for tabs
var sectionOrder = []string{
	"Education",
	"Experience",
	"Projects",
	"Skills & Interests",
	"Contact",
}

// --- Bubbles list.Item wrapper ---
type item struct {
	data listItemData
}

func (i item) Title() string       { return "" } // Not used by custom delegate
func (i item) Description() string { return "" } // Not used by custom delegate
func (i item) FilterValue() string { return i.data.FilterValue() }

// --- itemDelegate ---

var (
	// Define styles used in rendering items
	activeTabColor      = lipgloss.Color("78") // Bright Green for active tab/selection
	styleItemTitle      = lipgloss.NewStyle().Bold(true)
	styleItemSubtitle   = lipgloss.NewStyle().Faint(true)
	styleItemDesc       = lipgloss.NewStyle().PaddingLeft(1) // Base indent for descriptions
	styleSelectedBorder = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, false, true).
				BorderForeground(activeTabColor).
				PaddingLeft(1)
	styleNormal = lipgloss.NewStyle().PaddingLeft(2)
)

type itemDelegate struct{}

// Use a smaller fixed Height, plus Spacing between items
func (d itemDelegate) Height() int  { return 4 } // Smaller fixed height
func (d itemDelegate) Spacing() int { return 1 } // Spacing BETWEEN items
func (d itemDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(item)
	if !ok {
		return
	}

	itemWidth := m.Width()
	var contentStyle lipgloss.Style
	if index == m.Index() {
		contentStyle = styleSelectedBorder
	} else {
		contentStyle = styleNormal
	}
	contentWidth := itemWidth - contentStyle.GetHorizontalFrameSize() - contentStyle.GetHorizontalPadding()
	if contentWidth < 0 { contentWidth = 0 }

	var title, subtitle, details string
	var finalContent string
	data := it.data

	switch d := data.(type) {
	case educationItem:
		title = styleItemTitle.Render(d.School)
		subtitle = styleItemSubtitle.Render(fmt.Sprintf("%s • %s", d.Status, d.Dates))
		details = styleItemDesc.Copy().Width(contentWidth).Render(strings.Join(d.Details, "\n"))
		finalContent = lipgloss.JoinVertical(lipgloss.Left, title, subtitle, details)

	case experienceItem:
		title = styleItemTitle.Render(d.Company + " — " + d.Role)
		subtitle = styleItemSubtitle.Render(fmt.Sprintf("%s • %s", d.Dates, d.Location))
		descItems := make([]string, len(d.Description))
		for i, desc := range d.Description {
			descItems[i] = "• " + desc
		}
		details = styleItemDesc.Copy().Width(contentWidth).Render(strings.Join(descItems, "\n"))
		finalContent = lipgloss.JoinVertical(lipgloss.Left, title, subtitle, details)

	case projectItem:
		title = styleItemTitle.Render(d.Name)
		subtitle = styleItemSubtitle.Render(d.Dates)
		descItems := make([]string, len(d.Description))
		for i, desc := range d.Description {
			descItems[i] = "• " + desc
		}
		details = styleItemDesc.Copy().Width(contentWidth).Render(strings.Join(descItems, "\n"))
		finalContent = lipgloss.JoinVertical(lipgloss.Left, title, subtitle, details)

	case skillsItem:
		title = styleItemTitle.Render(d.Category)
		detailsList := make([]string, len(d.Details))
		for i, detail := range d.Details {
			detailsList[i] = "- " + detail
		}
		details = styleItemDesc.Copy().PaddingLeft(2).Width(contentWidth - 2).Render(strings.Join(detailsList, "\n")) 
		finalContent = lipgloss.JoinVertical(lipgloss.Left, title, details)

	case contactItem:
		centeredLineStyle := lipgloss.NewStyle().Width(contentWidth).Align(lipgloss.Center)
		finalContent = centeredLineStyle.Render(d.Line)
	}

	// Remove the PaddingBottom(1) added previously
	// contentStyle = contentStyle.PaddingBottom(1) 

	// Apply selection styling (left border) or normal padding to the whole block
	if index == m.Index() {
		styledBlock := contentStyle.Render(finalContent) 
		fmt.Fprint(w, styledBlock)
	} else {
		styledBlock := contentStyle.Render(finalContent)
		fmt.Fprint(w, styledBlock)
	}
}

// --- Model ---

type state int

const (
	splash state = iota
	mainUI
)

type loadedMsg struct{}

type model struct {
	s       state
	spin    spinner.Model
	w, h    int
	active  int // active tab index (index into sectionOrder)
	lst     list.Model
	vp      viewport.Model
	gotSize bool
}

func newModel() *model {
	sp := spinner.New()
	sp.Spinner = spinner.Pulse // Try pulse spinner
	sp.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("78")) // Use the active green color
	vp := viewport.New(0, 0)
	
	// Initialize list model here with defaults to prevent nil pointer panic
	lst := list.New([]list.Item{}, itemDelegate{}, 0, 0) // Empty items, delegate, zero size
	lst.SetShowHelp(false)
	lst.SetShowPagination(false)
	lst.SetShowStatusBar(false)
	lst.SetFilteringEnabled(false)
	lst.Title = ""
	lst.Styles.Title = lipgloss.NewStyle()

	return &model{s: splash, spin: sp, vp: vp, lst: lst} // Assign initialized list
}

func (m *model) setSize(w, h int) {
	m.w = w
	m.h = h
	m.gotSize = true

	headerHeight := lipgloss.Height(styleHeaderText.Render(" "))
	tabBarHeight := 1
	separatorLineHeight := 1
	helpViewHeight := 1
	contentHeight := m.h - headerHeight - tabBarHeight - separatorLineHeight - helpViewHeight
	if contentHeight < 1 { contentHeight = 1 }
	contentWidth := m.w

	m.lst.SetSize(contentWidth, contentHeight)
	m.vp.Width = contentWidth
	m.vp.Height = contentHeight
}

func (m *model) buildSkillsContent() string {
	var skillsBuilder strings.Builder
	skillsData, _ := resumeData["Skills & Interests"]

	contentWidth := m.vp.Width - styleItemDesc.GetHorizontalPadding()
	if contentWidth < 0 { contentWidth = 0 }

	for i, itemData := range skillsData {
		if skillsItem, ok := itemData.(skillsItem); ok {
			skillsBuilder.WriteString(styleItemTitle.Render(skillsItem.Category))
			skillsBuilder.WriteString("\n")
			detailsList := make([]string, len(skillsItem.Details))
			for j, detail := range skillsItem.Details {
				detailsList[j] = "- " + detail
			}
			details := styleItemDesc.Copy().PaddingLeft(2).Width(contentWidth).Render(strings.Join(detailsList, "\n")) 
			skillsBuilder.WriteString(details)
			if i < len(skillsData)-1 {
				skillsBuilder.WriteString("\n\n")
			}
		}
	}
	return skillsBuilder.String()
}

func (m *model) Init() tea.Cmd {
	return tea.Batch(
		m.spin.Tick,
		func() tea.Msg {
			// Longer delay for splash animation
			// time.Sleep(600 * time.Millisecond) // Old duration
			time.Sleep(3 * time.Second) // Set duration to 3 seconds
			return loadedMsg{}
		},
	)
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.setSize(msg.Width, msg.Height)
		if m.s == mainUI && sectionOrder[m.active] == "Skills & Interests" {
			m.vp.SetContent(m.buildSkillsContent())
		}

	case loadedMsg:
		if m.gotSize {
			m.enterMain()
		}
	}

	if m.s == mainUI {
		activeSection := sectionOrder[m.active]

		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "q", "ctrl+c":
				return m, tea.Quit
			case "left", "h":
				m.active = (m.active - 1 + len(sectionOrder)) % len(sectionOrder)
				m.rebuildList()
				return m, nil
			case "right", "l":
				m.active = (m.active + 1) % len(sectionOrder)
				m.rebuildList()
				return m, nil
			case "1", "2", "3", "4", "5":
				 idx, err := strconv.Atoi(keyMsg.String())
				 if err == nil && idx >= 1 && idx <= len(sectionOrder) {
					 m.active = idx - 1
					 m.rebuildList()
					 return m, nil
				 }
			}
		}

		if activeSection == "Skills & Interests" {
			m.vp, cmd = m.vp.Update(msg)
			cmds = append(cmds, cmd)
		} else if activeSection != "Contact" {
			m.lst, cmd = m.lst.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	if m.s == splash {
		m.spin, cmd = m.spin.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m *model) enterMain() {
	m.s = mainUI
	m.rebuildList()
}

func (m *model) rebuildList() {
	activeSection := sectionOrder[m.active]
	if activeSection == "Skills & Interests" {
		if m.gotSize {
			m.vp.SetContent(m.buildSkillsContent())
			m.vp.GotoTop()
		}
	} else if activeSection != "Contact" {
		idx := 0
		if m.lst.Items() != nil && len(m.lst.Items()) > 0 {
			idx = m.lst.Index()
		}
		m.lst = m.newList()
		if idx >= len(m.lst.Items()) {
			idx = 0
			if len(m.lst.Items()) > 0 {
				idx = len(m.lst.Items()) - 1
			}
		}
		if len(m.lst.Items()) > 0 {
			m.lst.Select(idx)
		}
	}
}

func (m *model) newList() list.Model {
	activeSectionTitle := sectionOrder[m.active]
	
	var listItems []list.Item
	if activeSectionTitle != "Skills & Interests" && activeSectionTitle != "Contact" {
		itemsData, exists := resumeData[activeSectionTitle]
		if exists {
			listItems = make([]list.Item, len(itemsData))
			for i, data := range itemsData {
				listItems[i] = item{data: data}
			}
		}
	}

	d := itemDelegate{}
	listHeight := 1
	if m.gotSize {
		headerHeight := lipgloss.Height(styleHeaderText.Render(" "))
		tabBarHeight := 1
		separatorLineHeight := 1
		helpViewHeight := 1
		listHeight = m.h - headerHeight - tabBarHeight - separatorLineHeight - helpViewHeight
		if listHeight < 1 { listHeight = 1 }
	}

	l := list.New(listItems, d, m.w, listHeight)
	l.Title = ""
	l.SetFilteringEnabled(false)
	l.Styles.Title = lipgloss.NewStyle()
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)
	l.SetShowPagination(false)
	return l
}

// --- View ---

var (
	// Define styles specific to the View function (tabs, help, filler)
	inactiveTabFg   = lipgloss.Color("248")
	helpColor       = lipgloss.Color("241")

	// Base style for tabs
	styleTabBase = lipgloss.NewStyle().Padding(0, 1)

	// Inactive tabs: dim foreground, no background
	styleTabInactive = styleTabBase.Copy().
				Foreground(inactiveTabFg)

	// Active tab: Use active color (green) foreground, bold, NO border or background
	styleTabActive = styleTabBase.Copy().
				Bold(true).
				Foreground(activeTabColor)

	// Filler has no background
	styleFiller = lipgloss.NewStyle()

	// Help text style
	styleHelp = lipgloss.NewStyle().
				Foreground(helpColor).
				Padding(0, 1)
	
	// Style for the header text above tabs
	styleHeaderText = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("213")). // A magenta/pink color like Soft Serve title
				Padding(1, 0, 0, 1) // Padding top/bottom/left

	// Style for the separator line below tabs
	styleSeparatorLine = lipgloss.NewStyle().
				Foreground(inactiveTabFg) // Use faint color

	// Style for the centered contact block
	styleContactBlock = lipgloss.NewStyle().Align(lipgloss.Center)
)

func (m *model) View() string {
	if !m.gotSize {
		return "Initializing..."
	}

	if m.s == splash {
		// Changed splash text
		introText := " Liem Luttrell - Portfolio"
		ui := lipgloss.JoinHorizontal(lipgloss.Center, m.spin.View(), introText)
		splashStyle := lipgloss.NewStyle().Padding(1, 2).Foreground(lipgloss.Color("248")) // Dim text color
		// Place centered
		return lipgloss.Place(m.w, m.h, lipgloss.Center, lipgloss.Center, splashStyle.Render(ui))
	}

	// --- Render Static Parts --- 
	headerText := "Liem Luttrell - SSH Portfolio"
	renderedHeader := styleHeaderText.Render(headerText)
	headerHeight := lipgloss.Height(renderedHeader)

	var renderedTabStrings []string 
	separator := lipgloss.NewStyle().Foreground(inactiveTabFg).Render(" | ")
	for i, title := range sectionOrder {
		style := styleTabInactive
        if i == m.active {
			style = styleTabActive
		}
		tabTitle := fmt.Sprintf("%d. %s", i+1, title)
		renderedTabStrings = append(renderedTabStrings, style.Render(tabTitle))
	}
	joinedTabs := strings.Join(renderedTabStrings, separator)
	remainingWidth := m.w - lipgloss.Width(joinedTabs)
	if remainingWidth < 0 { remainingWidth = 0 }
	filler := styleFiller.Copy().Width(remainingWidth).Render("")
	tabBar := lipgloss.JoinHorizontal(lipgloss.Bottom, joinedTabs, filler)
	tabBarHeight := lipgloss.Height(tabBar)

	separatorLine := styleSeparatorLine.Render(strings.Repeat("─", m.w))
	separatorLineHeight := 1

	helpView := styleHelp.Render("←/→ or h/l: switch • ↑/↓: navigate • q: quit")
	helpViewHeight := 1

	// --- Calculate Content Area Dimensions --- 
	contentHeight := m.h - headerHeight - tabBarHeight - separatorLineHeight - helpViewHeight
	if contentHeight < 1 { contentHeight = 1 }
	contentWidth := m.w

	// --- Render Content Area (Conditional) --- 
	var contentView string
	activeSectionTitle := sectionOrder[m.active]

	if activeSectionTitle == "Contact" {
		// Special rendering for Contact tab
		contactData, _ := resumeData["Contact"]
		var contactLines []string
		for _, itemData := range contactData {
			if contactItem, ok := itemData.(contactItem); ok {
				contactLines = append(contactLines, contactItem.Line)
			}
		}
		contactBlockStr := strings.Join(contactLines, "\n")
		styledContactBlock := styleContactBlock.Render(contactBlockStr)
		contentView = lipgloss.Place(contentWidth, contentHeight, lipgloss.Center, lipgloss.Center, styledContactBlock)

	} else if activeSectionTitle == "Skills & Interests" {
		m.vp.Width = contentWidth
		m.vp.Height = contentHeight
		if m.vp.TotalLineCount() == 0 {
			m.vp.SetContent(m.buildSkillsContent())
		}
		contentView = m.vp.View()

	} else {
		m.lst.SetSize(contentWidth, contentHeight)
		contentView = m.lst.View()
	}

	// --- Create Explicit Container for Content Area --- 
	contentAreaStyle := lipgloss.NewStyle().Width(contentWidth).Height(contentHeight)
	renderedContentArea := contentAreaStyle.Render(contentView)

	// --- Final Layout --- 
	return lipgloss.JoinVertical(lipgloss.Left,
		renderedHeader, 
		tabBar,
		separatorLine, 
		renderedContentArea,
		helpView,
	)
}


// --- Wish boilerplate ---

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	pty, _, active := s.Pty()
	if !active {
		fmt.Fprintln(s.Stderr(), "No active PTY required.") // Use Stderr
		return nil, nil // Return nil model if no PTY
	}
	
	// Create model *after* potentially getting PTY dims
	m := newModel()
	// Initial dimensions might be 0, wait for WindowSizeMsg
	m.w = pty.Window.Width 
	m.h = pty.Window.Height
	if m.w > 0 && m.h > 0 {
		m.gotSize = true // Mark size as received if PTY provided it
	}

	// Use AltScreen and potentially Mouse
	opts := []tea.ProgramOption{tea.WithAltScreen(), tea.WithMouseCellMotion()}
	return m, opts
}


func main() {
	// Consider adding host key generation/loading if needed
	// e.g., hostKey := wish.WithHostKeyPath(".ssh/soft_serve_server_key")
    srv, err := wish.NewServer(
        wish.WithAddress(net.JoinHostPort("0.0.0.0", "23234")),
		// hostKey, // Add host key option if defined
        wish.WithMiddleware(
			// Order matters: BubbleTea middleware should run after PTY is set up
			// but before logging potentially? Let's keep it first for now.
            wb.Middleware(teaHandler),
			activeterm.Middleware(), // Handles PTY requests
            wl.Middleware(),
        ),
    )
    if err != nil {
		log.Fatalf("Could not create server: %v", err)
    }

	log.Printf("Starting SSH server on port 23234...")
	log.Printf("Connect with: ssh <user>@<host> -p 23234") // More generic connect string
    log.Fatal(srv.ListenAndServe())
}

