package gui

import (
	"fmt"
	"strings"

	"github.com/drgarcia1986/gonews/story"
	"github.com/drgarcia1986/gonews/utils"
	"github.com/fatih/color"
	"github.com/jroimartin/gocui"
)

var (
	downKeys = []interface{}{'j', gocui.KeyArrowDown}
	upKeys   = []interface{}{'k', gocui.KeyArrowUp}
	quitKeys = []interface{}{'q', gocui.KeyCtrlC}

	keybindingMap = []struct {
		keys     []interface{}
		viewName string
		event    func(*gocui.Gui, *gocui.View) error
	}{
		{quitKeys, "", quit},
		{downKeys, "", cursorDown},
		{upKeys, "", cursorUp},
		{[]interface{}{'?'}, "main", helpMsg},
	}
)

type Gui struct {
	items        []*story.Story
	providerName string
}

func (gui *Gui) openStory(g *gocui.Gui, v *gocui.View) error {
	s, err := gui.getStoryOfCurrentLine(v)
	if err == nil && s != nil {
		return utils.OpenURL(s.URL)
	}
	return err
}

func (gui *Gui) preview(g *gocui.Gui, v *gocui.View) error {
	s, err := gui.getStoryOfCurrentLine(v)
	if err != nil {
		return err
	}

	if s == nil {
		return nil
	}

	content, err := utils.GetPreview(s.URL)
	if err != nil {
		return err
	}

	if content == "" {
		content = "No preview available"
	}
	return showPreview(g, s.Title, content)
}

func (gui *Gui) comments(g *gocui.Gui, v *gocui.View) error {
	s, err := gui.getStoryOfCurrentLine(v)
	if err == nil && s != nil {
		return utils.OpenURL(s.CommentsURL)
	}
	return err
}

func (gui *Gui) getStoryOfCurrentLine(v *gocui.View) (*story.Story, error) {
	_, cy := v.Cursor()
	line, err := v.Line(cy)
	if err != nil {
		return nil, err
	}

	for _, s := range gui.items {
		if strings.Contains(line, s.Title) {
			return s, nil
		}
	}
	return nil, nil
}

func (gui *Gui) layout(g *gocui.Gui) error {
	maxX, maxY := g.Size()
	if v, err := g.SetView("main", 0, 0, maxX-1, maxY-1); err != nil {
		if err != gocui.ErrUnknownView {
			return err
		}
		v.Title = fmt.Sprintf("GoNews - %s ('?' for help)", gui.providerName)
		v.Highlight = true
		v.SelBgColor = gocui.ColorBlue
		v.SelFgColor = gocui.ColorBlack
		v.FgColor = gocui.ColorWhite

		colored := func(s string) string {
			code := int(color.FgMagenta)
			attr := []color.Attribute{color.Attribute(code)}
			return color.New(attr...).SprintFunc()(s)
		}

		pretty := func(s *story.Story) string {

			comments := ""
			switch {
			case s.CommentsCount > 100:
				comments = "[" + colored(fmt.Sprintf("%d", s.CommentsCount)) + "]"
			case s.CommentsCount > 0:
				comments = fmt.Sprintf("[%d]", s.CommentsCount)
			}

			return fmt.Sprintf(
				"%s %s %s",
				colored(s.Title),
				s.Domain(gui.providerName),
				comments,
			)
		}

		for _, story := range gui.items {
			fmt.Fprintln(v, pretty(story))
		}

		if _, err := g.SetCurrentView("main"); err != nil {
			return err
		}
	}
	return nil
}

func (gui *Gui) keybindings(g *gocui.Gui) error {
	for _, bm := range keybindingMap {
		for _, key := range bm.keys {
			if err := g.SetKeybinding(bm.viewName, key, gocui.ModNone, bm.event); err != nil {
				return err
			}
		}
	}

	if err := g.SetKeybinding("main", gocui.KeyEnter, gocui.ModNone, gui.openStory); err != nil {
		return err
	}

	if err := g.SetKeybinding("main", 'p', gocui.ModNone, gui.preview); err != nil {
		return err
	}

	if err := g.SetKeybinding("main", 'c', gocui.ModNone, gui.comments); err != nil {
		return err
	}

	return nil
}

func (gui *Gui) Run() error {
	g, err := gocui.NewGui(gocui.OutputNormal)
	if err != nil {
		return err
	}
	defer g.Close()

	g.SetManagerFunc(gui.layout)
	if err := gui.keybindings(g); err != nil {
		return err
	}

	if err := g.MainLoop(); err != nil && err != gocui.ErrQuit {
		return err
	}
	return nil
}

func New(items []*story.Story, providerName string) *Gui {
	guiItems := make([]*story.Story, len(items))
	copy(guiItems, items)
	return &Gui{items: guiItems, providerName: providerName}
}
