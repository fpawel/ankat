package uiworks

import (
	"github.com/fpawel/ankat/internal/db/worklog"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Task struct {
	parent      *Task
	children    []*Task
	descendants []*Task
	name        string
	action      Action
	ordinal     int
}

type Action func() error

var errorInterrupted = errors.New("выполнение прервано")

const (
	csUncheckedNormal = iota
	csUncheckedPressed
	csCheckedNormal
)

func (x *Task) Root() (root *Task) {
	root = x
	for root.parent != nil {
		root = root.parent
	}
	return
}

func (x *Task) enumDescendants(descendants *[]*Task) {
	*descendants = append(*descendants, x)
	for _, y := range x.children {
		y.enumDescendants(descendants)
	}
}

func (x *Task) Descendants() []*Task {
	return x.descendants
}

func (x *Task) Checked(dbConfig *sqlx.DB) bool {
	checkState := x.CheckState(dbConfig)
	return checkState != csUncheckedNormal && checkState != csUncheckedPressed
}

func (x *Task) CheckState(dbConfig *sqlx.DB) (checkState int) {

	var xs []int

	err := dbConfig.Select(&xs, `SELECT checked FROM work_checked WHERE work_order = $1;`, x.ordinal)
	if err != nil {
		panic(err)
	}
	if len(xs) == 0 {
		checkState = csCheckedNormal
	} else {
		checkState = xs[0]
	}
	return
}

func (x *Task) perform(ui Runner, dbConfig *sqlx.DB) error {

	if !x.Checked(dbConfig) {
		return nil
	}

	ui.notifyWork(x, true)
	defer func() {
		ui.notifyWork(x, false)
	}()

	if ui.Interrupted() {
		return errorInterrupted
	}

	if x.action != nil {
		err := x.action()
		if err != nil {
			ui.WriteLog(0, worklog.Error, err.Error())
		}
		return err
	}
	for _, y := range x.children {
		if err := y.perform(ui, dbConfig); err != nil {
			return err
		}
		if ui.Interrupted() {
			return errorInterrupted
		}
	}
	return nil
}

func (x *Task) GetTaskByOrdinal(ordinal int) (m *Task) {
	for i, m := range x.Descendants() {
		if i == ordinal {
			return m
		}
	}
	panic("unexpected")
}

func (x *Task) Key() worklog.Work {
	return worklog.Work{
		Index: x.ordinal,
		Name:  x.name,
	}
}

func (x *Task) ParentKeys() (xs []worklog.Work) {
	xs = append(xs, x.Key())
	for y := x; y.parent != nil; y = y.parent {
		xs = append([]worklog.Work{y.parent.Key()}, xs...)
	}
	return
}

func (x *Task) Text() (text string) {
	text = x.name
	for y := x.parent; y != nil && y.parent != nil; y = y.parent {
		text = y.name + ". " + text
	}
	return
}

type TaskInfo struct {
	Name       string
	Ordinal    int
	HasError   bool
	HasMessage bool
	Children   []TaskInfo
}

func (x *Task) Name() string {
	return x.name
}

func (x *Task) Ordinal() int {
	return x.ordinal
}

func (x *Task) Info(dbLog *sqlx.DB) (m TaskInfo) {
	w := worklog.GetLastWorkInfo(dbLog, x.ordinal, x.name)
	m = TaskInfo{
		Name:       x.name,
		Ordinal:    x.ordinal,
		HasError:   w.HasError,
		HasMessage: w.HasMessage,
	}

	for _, y := range x.children {
		m.Children = append(m.Children, y.Info(dbLog))
	}
	return
}
