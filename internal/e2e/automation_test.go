//go:build !js

package e2e

import (
	"errors"
	"testing"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/automation"
)

func TestContentWindowAutomation(t *testing.T) {
	on := false
	w := newWindow(t, mygo.WindowOptions{Title: "Automation", Width: 320, Height: 240,
		Content: ui.View(func(c *ui.Context) { ui.Checkbox(c, &on, "Agree") }).Automatable()})
	driver := mygo.Automate(w)
	var snapshot automation.Snapshot
	eventually(t, "the semantic tree", func() bool {
		var err error
		snapshot, err = driver.Inspect()
		if err != nil {
			t.Fatal(err)
		}
		_, err = snapshot.Resolve(0, "Agree")
		return err == nil
	})
	for _, action := range []automation.Action{automation.Press, automation.Click} {
		target, err := snapshot.Resolve(0, "Agree")
		if err != nil {
			t.Fatal(err)
		}
		if err := driver.Perform(target, automation.Request{Action: action}); err != nil {
			t.Fatal(err)
		}
		snapshot, err = driver.Inspect()
		if err != nil {
			t.Fatal(err)
		}
		for _, node := range snapshot.Nodes {
			if node.Name == "Agree" && node.Checked != (action == automation.Press) {
				t.Fatalf("state after %s: %+v", action, node)
			}
		}
	}
	w.Destroy()
	if _, err := driver.Inspect(); !errors.Is(err, automation.ErrClosed) {
		t.Fatalf("closed: %v", err)
	}
}

func TestContentWindowAutomationInput(t *testing.T) {
	value := "before"
	w := newWindow(t, mygo.WindowOptions{Width: 320, Height: 240,
		Content: ui.View(func(c *ui.Context) { ui.TextInput(c, &value).Label("Name") }).Automatable()})
	driver := mygo.Automate(w)
	var target automation.Target
	eventually(t, "the text field", func() bool {
		s, err := driver.Inspect()
		if err != nil {
			t.Fatal(err)
		}
		target, err = s.Resolve(0, "Name")
		return err == nil
	})
	if err := driver.Perform(target, automation.Request{Action: automation.Focus}); err != nil {
		t.Fatal(err)
	}
	for _, input := range []automation.Input{
		{Kind: automation.KeyInput, Key: "CmdOrCtrl+A"},
		{Kind: automation.TextInput, Text: "after 🌍"},
	} {
		if err := driver.Input(input); err != nil {
			t.Fatal(err)
		}
	}
	snapshot, err := driver.Inspect()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := snapshot.Resolve(target.ID, ""); err != nil {
		t.Fatal(err)
	}
	for _, node := range snapshot.Nodes {
		if node.ID == target.ID && node.Value != "after 🌍" {
			t.Fatalf("text input: %q", node.Value)
		}
	}
}
